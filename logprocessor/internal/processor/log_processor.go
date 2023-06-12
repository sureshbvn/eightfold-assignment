// Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains processor related utils.

package processor

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/glog"
	"github.com/spf13/viper"

	"logprocessor/internal/config"
	"logprocessor/internal/messageq"
)

// Mutex to synchronize access to the logLine variable
var logLineMutex sync.Mutex

type LogProcessor struct {
	// The configuration object.
	conf *viper.Viper

	// The kafka producer for the log processor.
	producer *kafka.Producer
}

// NewLogProcessor creates a new instance of the LogProcessor.
func NewLogProcessor(conf *viper.Viper, producer *kafka.Producer) *LogProcessor {
	return &LogProcessor{
		conf:     conf,
		producer: producer,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

// ProcessLogs is a the struct method to process the logs.
func (processor *LogProcessor) ProcessLogs() {
	// Get the input logs directory from config.
	inputLogsDir := processor.conf.GetString(config.KLogsDirectory)
	glog.Infoln("The logs directory", inputLogsDir)

	// Get a list of files in the data directory.
	filePaths, err := filepath.Glob(inputLogsDir + "/*")
	if err != nil {
		glog.Fatalf("Failed to get list of log files: ", err)
	}

	glog.Infoln("Printing all the log files in the directory...")
	for _, fileName := range filePaths {
		glog.Infoln(fileName)
	}

	maxParallelLines := 100
	maxFilesPerBatch := 10
	logLines := make(chan string, maxParallelLines)
	var wg sync.WaitGroup

	topic := processor.conf.GetString(config.KTopic)

	// Process log files in batches
	for i := 0; i < len(filePaths); i += maxFilesPerBatch {
		glog.Infoln("Processing file with index: ", filePaths[i])
		// Determine the end index of the current batch
		end := i + maxFilesPerBatch
		if end > len(filePaths) {
			end = len(filePaths)
		}

		// Process log files in the current batch
		for _, filePath := range filePaths[i:end] {
			wg.Add(1)
			glog.Infoln("Starting a go routine for a file: ", filePath)
			go processor.ProcessLogFile(filePath, logLines, &wg)
		}

		// Start goroutine to publish log lines to Kafka
		go messageq.PublishToKafka(logLines, processor.producer, topic)

		// Wait for the current batch to finish processing
		wg.Wait()
	}

	// Close the logLines channel to signal the end of processing
	close(logLines)
}

//----------------------------------------------------------------------------------------------------------------------

// ProcessLogFile is a helper function to process the data from a given file path.
func (processor *LogProcessor) ProcessLogFile(filePath string, logLines chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open file: %s - %s", filePath, err.Error())
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Regex pattern to match the log line format
	pattern := `(\d+:\d+::[\w-]+ \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2},\d{3}) - (.*)`
	regex := regexp.MustCompile(pattern)

	// Variable to store the current log line
	var logLine string

	for scanner.Scan() {
		line := scanner.Text()

		// Check if the line matches the log line format
		match := regex.FindStringSubmatch(line)
		if len(match) == 3 {
			// The line matches the log line format
			logLineMutex.Lock()
			if logLine != "" {
				// If there is a previous log line, send it to the logLines channel
				logLines <- logLine
			}
			// Set the current log line to the matched line
			logLine = line
			logLineMutex.Unlock()
		} else {
			// The line does not match the log line format, append it to the current log line
			logLineMutex.Lock()
			logLine += "\n" + line
			logLineMutex.Unlock()
		}
	}

	if err := scanner.Err(); err != nil {
		glog.Infoln("Error reading file: %s - %s", filePath, err.Error())
	}

	// Send the last log line to the logLines channel
	logLineMutex.Lock()
	if logLine != "" {
		logLines <- logLine
	}
	logLineMutex.Unlock()
}

//----------------------------------------------------------------------------------------------------------------------
