// Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains main file for log-subscriber.
//
// The log-subscriber is a microservice which establishes multiple consumers to kafka which contains log lines
// created from the log processor services.
//
// The log-subscriber has two workers.
//  1. File worker to prepare sanitized logs.
//  2. Stats worker to prepare stats to answer the apis.
//
// The services is completely stateless and a number of replicas of the log-subscriber.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/viper"

	"logworker/internal/config"
	"logworker/internal/messageq"
	"logworker/internal/workers"
)

func init() {
	flag.Parse()
	flag.Set("logtostderr", "true")
}

func main() {
	// Step (1): The following block is needed for the logger package to work correctly. Assume
	// that this is boiler-plate code and no need to look into this.
	defer glog.Flush()

	// At this point, logger object is ready and we can start logging messages to stdout.
	glog.Infoln("Starting log-subscriber process")

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Step (2): Load the configuration.
	conf := config.LoadConfiguration()

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Step (3): Clean up any old sanitized log files directory.
	if err := mayBeDeleteOldSanitizedDir(conf); err != nil {
		glog.Fatalf(err.Error())
	}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Step (4): Create all the kafka consumers.
	// Create Kafka consumer for file worker
	fileConsumer := messageq.CreateKafkaConsumer(conf, "file-consumer-group-id")
	defer fileConsumer.Close()

	// Create Kafka consumer for stats worker
	statsConsumer := messageq.CreateKafkaConsumer(conf, "stats-consumer-group-id")
	defer statsConsumer.Close()

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Step (4): Create all the workers which process log statements from kafka.

	// Create context for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())

	// Create file worker.
	fileWorker := workers.NewFileWorker(conf, fileConsumer)
	go func() {
		err := fileWorker.Start(ctx)
		if err != nil {
			glog.Fatalf("File worker error: %v", err)
		}
	}()

	// Create stats worker.
	statsWorker := workers.NewStatsWorker(conf, statsConsumer)
	go func() {
		err := statsWorker.Start(ctx)
		if err != nil {
			glog.Fatalf("Stats worker error: %v", err)
		}
	}()

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// Step (5):
	// Wait for termination signal to gracefully shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals

	// Cancel the context to signal workers to stop
	cancel()
}

//----------------------------------------------------------------------------------------------------------------------

// mayBeDeleteOldSanitizedDir is a helper function to delete the old sanitized directory if exists.
func mayBeDeleteOldSanitizedDir(conf *viper.Viper) error {
	// Retrieve the directory path from sanitized log directory. This can contain output from previous runs.
	dirPath := conf.GetString(config.KSanitizedLogsDirectory)

	// Check if the directory exists.
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		// Directory does not exist, no action needed.
		glog.Infoln("The sanitized directory does not exist")
		return nil
	}

	// Delete the directory and its contents.
	err = os.RemoveAll(dirPath)
	if err != nil {
		// Handle the error if deletion fails
		msg := fmt.Sprintf("Failed to delete directory: %v\n", err)
		return fmt.Errorf(msg)
	}

	// Directory successfully deleted.
	glog.Infoln("Directory deleted:", dirPath)
	return nil
}

//----------------------------------------------------------------------------------------------------------------------
