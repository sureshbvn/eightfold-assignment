// Package cmd Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains worker class for sanitized file writing.
//
// The kafka queue contains messages related to the log lines. Each message
// in kafka is related to a single log line. The kafka partitioning scheme is using a hash function.
// The message key is (processId-threadId). Kafka is sequential in nature. This means all the messages
// in a given partition is read sequentially. A folder called data is mounted from hostpath to this microservice
// as /app/data.
//
// At a high-level file-worker does the following.
// 1. Create a folder called "sanitized" directory. This is where all the sanitized data is going to be written.
// 2. Establish a infinite loop which acts as consumer. Read one message at a time from the kafka queue (From a given
//    partition).
// 3. The strategy here is to write to one file per (process-id:thread-id). Keep the file descriptors open in the
//    memory for efficient catching.

package workers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/glog"
	"github.com/spf13/viper"

	"logworker/internal/config"
)

type FileWorker struct {
	// The configuration object.
	conf *viper.Viper

	// The kafka consumer established for the file worker.
	consumer *kafka.Consumer
}

// NewFileWorker creates a new instance of the FileWorker.
func NewFileWorker(conf *viper.Viper, consumer *kafka.Consumer) *FileWorker {
	return &FileWorker{
		conf:     conf,
		consumer: consumer,
	}
}

//----------------------------------------------------------------------------------------------------------------------

// Start starts the FileWorker and begins consuming messages from Kafka.
func (worker *FileWorker) Start(ctx context.Context) error {

	sanitizedDir := worker.conf.GetString(config.KSanitizedLogsDirectory)
	glog.Infoln("The sanitized directory", sanitizedDir)
	// Create the log directory if it doesn't exist.
	if err := os.MkdirAll(sanitizedDir, os.ModePerm); err != nil {
		glog.Fatal("failed to create log directory:", err)
	}

	// Get the kafka topic name from the configuration object.
	topic := worker.conf.GetString(config.KTopic)

	// Subscribe to the log processor topic. Please note that this is just establishing the subscription. The messages
	// must be still read. It is read in an infinite for select below
	err := worker.consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("failed to subscribe to Kafka topic: %v", err)
	}

	// If we are here the consumer is successfully established.
	glog.Infoln("The consumer established for file worker and  topic: ", topic)

	// Map to store the log files per process. This can be thought of as in-memory cache for all the opened file
	// descriptors.
	logFiles := make(map[string]*os.File)

	// Start consuming messages by establishing a for select.
	for {
		select {
		case <-ctx.Done():
			// Stop the worker gracefully. Close all the open file descriptors before closing the go routine.
			worker.closeLogFiles(logFiles)
			return nil
		default:
			// This infinitely blocks until next message is available for the consumer group to consume.
			msg, err := worker.consumer.ReadMessage(-1)
			if err != nil {
				glog.Error("error while consuming message: ", err)
				continue
			}

			// Extract process ID and thread ID from Kafka key
			key := string(msg.Key)
			parts := strings.Split(key, "-")
			if len(parts) < 2 {
				glog.Error("Invalid key format:", key)
				continue
			}
			processID := parts[0]
			//threadID := parts[1]

			// Check if the file descriptor is available in cache.
			logFile, ok := logFiles[processID]
			if !ok {
				// If we reach here, the file descriptor is not available and hence we are creating it.
				fileName := fmt.Sprintf("%s/%s.log", sanitizedDir, processID)
				logFile, err = os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					glog.Errorf("failed to create log file for process %s: %v",
						processID, err)
					continue
				}

				// Cache the file descriptor.
				logFiles[processID] = logFile
			}

			// Sanitize the log message.
			logMessage := string(msg.Value)

			// Write the log message to the process's log file.
			_, err = logFile.WriteString(logMessage + "\n")
			if err != nil {
				glog.Errorf("failed to write log message for process %s: %v", processID, err)
			}
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------

// closeLogFiles is a helper function to close all all the open file descriptors that are cached in memory.
func (worker *FileWorker) closeLogFiles(logFiles map[string]*os.File) {
	for _, logFile := range logFiles {
		logFile.Close()
	}
}

//----------------------------------------------------------------------------------------------------------------------
