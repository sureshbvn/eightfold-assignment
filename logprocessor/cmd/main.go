// Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains main file for log-processor.
//
// The log-processor is a microservice that reads input files directory and writes one line at a time to kafka.
//
// Important requirements that are inferred from the problem.
//
// 1. A single log file cannot be fit into memory. So its important we read the entire file from the memory.
// 2. There is a no control on the number of processes running on the node. This means reading one file at a time is not
//    an option. We need to read to N files in parallel but only read few lines (Say M files) at a time. N and M must be
//    configurable to be adjusted based on issues like OOM and number of open FDs etc..
//
// High-level approach:
//
// 1. Scan the input folder for files. We are going to process the files in batches.
//
// 2. The batch size is determined by the configuration parameter "max_files_per_batch" in the defaults.yaml.
//
// 3. For each file in the batch establish, start a goroutine (a thread equivalent) to process the file.
//
// 4. The file processing involves reading one line at a time from the file and writing to a buffered channel.
//
// 5. At this point we have "max_files_per_batch" number of go routines which are concurrently running. Each go routine
//    is processing one file from the batch. It is reading one line at a time and writing it into the
//    buffered channel. The buffered channel is created with a capacity of "max_parallel_lines", another
//    configuration parameter defined in the defualts.yaml. This dictates total number of lines across all the
//    files in batch to be processed. Both these parameters will help us improve the parallelism but control the
//    compute utilization per replica.
//
// 6. We can assume buffered channel as a thread safe in memory FIFO queue. The size of this buffered channel is usually
//    larger than total number of files that are processed in a given batch.
//
// 7. For each batch of files establish a go-routine(consumer thread) to listen to buffered channel (FIFO) queue.
//    Dequeue from the buffered channel and write the message to kafka. Most kafka clients do in-memory buffering
//    any way. No need to additional buffering.
//
// 8. Till now we have the following.
//         a) Producer :- go routines which are reading the files in "max_files_per_batch" batch and writing one line
//                         at a time and writing to buffered channel (Thread safe FIFO queue).
//         b) Consumer :- one go routine per batch which listens to the buffered channel and writes them to kafka.
//
// Kafka partitioning strategy:
//
// 1. Since we are reading a number of files in parallel and writing to kafka, we need to make sure that consumer of
//    of the topic is reading all the changes from a given (process-id, thread-id) sequentially.
//
// 2. So this translates to having hash based partitioning. Kafka uses murmur2 for hash based partitioning. We will be
//    partitioning based on the (process-id, thread-id) so that all changes related to given process-id and thread-id
//    will be serialized.
//
// 3. We could choose to have increased number of partitions depending on our scale.
//
// 4. Partitioning simply based on process-id is also an option. But there could be hotspots here. There could be a very
//    large process on this IRCTC node with way too many threads and this particular partition which is handling this
//    process-id can become a hotspot. The developer only cares about having sanitized logs at a thread level. Hence an
//    assumption is made to serialize at (process-id, thread-id) level.
//
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"

	"logprocessor/internal/config"
	"logprocessor/internal/messageq"
	"logprocessor/internal/processor"
)

func init() {
	flag.Parse()
	flag.Set("logtostderr", "true")
}

func main() {
	// Step1: The following block is needed for the logger package to work correctly. Assume
	// that this is boiler-plate code and no need to look into this.
	defer glog.Flush()

	// At this point, logger object is ready and we can start logging messages to stdout.
	glog.Infoln("Starting log-processor process")

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Step (2): Load the configuration.
	conf := config.LoadConfiguration()

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Step (3): Create the kafka topic if it does not exist. Also create the kafka producer.

	// Create the kafka topic if it does not exist.
	if err := messageq.MaybeCreateKafkaTopic(conf); err != nil {
		glog.Fatalf("Failed to create Kafka topic:", err)
	}

	// Create Kafka producer configuration
	// Create the Kafka producer
	producer, err := messageq.CreateKafkaProducer(conf)
	if err != nil {
		glog.Fatalf("Failed to create Kafka producer:", err)
	}
	defer producer.Close()

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Step (4): Create the processor object to process the logs.
	proc := processor.NewLogProcessor(conf, producer)
	proc.ProcessLogs()

	glog.Infoln("Completed processing all the files in the input logs directory")

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Step (5):
	// Wait for termination signal to gracefully shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals
}

//----------------------------------------------------------------------------------------------------------------------
