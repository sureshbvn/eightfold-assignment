// Package cmd Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains worker class for creating OLAP stats for APIs.
//
// The kafka queue contains messages related to the log lines. Each message
// in kafka is related to a single log line. The kafka partitioning scheme is using a hash function.
// The message key is (processId-threadId). Kafka is sequential in nature. This means all the messages
// in a given partition is read sequentially. A folder called data is mounted from hostpath to this microservice
// as /app/data.
//
// At a high-level stats-worker does the following.
//
// 1. Establish a infinite loop which acts as consumer. Read one message at a time from the kafka queue (From a given
//    partition).
// 2. Process each line.
//          a) Extract process_id, thread_id, timestamp, log message from the log.
//          b) Write them to postgres database.

package workers

import (
	"context"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-pg/pg/v10"
	"github.com/golang/glog"
	"github.com/spf13/viper"

	"logworker/internal/config"
	"logworker/internal/db"
)

// LogLine encapsulates the structure of postgres table. The table name is specified in the tableName field below.
// To be precise its called "log_lines".
//
// The primary key here is a composite key between process_id, thread_id, timestamp and timestamp millis.
// Note that timestamp is field in "timestampz" format. This is done to make sure that original timestamp is preserved
// with milliseconds precision.
// In addition to timestamp, there is an addition field called timestamp_seconds. This is because all the queries are at
// seconds precision.
// The timestamp_seconds is also interesting because it will allow us to partition the table based on the seconds.
// To create a partition the filed must be part of unique key or primary key. Hence the timestamp_seconds field is
// added.
type LogLine struct {
	tableName        struct{}  `pg:"log_lines"`
	ProcessID        string    `pg:"process_id,notnull,pk"`
	ThreadID         string    `pg:"thread_id,notnull,pk"`
	Timestamp        time.Time `pg:"timestamp,notnull,pk"`
	TimestampSeconds int64     `pg:"timestamp_seconds,notnull,pk"`
	LogMessage       string    `pg:"log_message"`
}

// StatsWorker implements the worker interface.
type StatsWorker struct {
	conf     *viper.Viper
	consumer *kafka.Consumer
	db       *pg.DB
}

// NewStatsWorker returns new instance of StatsWorker.
func NewStatsWorker(conf *viper.Viper, consumer *kafka.Consumer) *StatsWorker {
	return &StatsWorker{
		conf:     conf,
		consumer: consumer,
	}
}

//----------------------------------------------------------------------------------------------------------------------

func (worker *StatsWorker) Start(ctx context.Context) error {
	// Get the kafka topic name from the configuration object.
	topic := worker.conf.GetString(config.KTopic)

	// Subscribe to the log processor topic. Please note that this is just establishing the subscription. The messages
	// must be still read. It is read in an infinite for select below
	err := worker.consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("failed to subscribe to Kafka topic: %v", err)
	}

	// Create a new db object.
	worker.db = db.NewDB(worker.conf)

	glog.Infof("Stats consumer established for topic: %s", topic)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// This infinitely blocks until next message is available for the consumer group to consume.
			msg, err := worker.consumer.ReadMessage(-1)
			if err != nil {
				glog.Errorf("error while consuming message: %v", err)
				continue
			}

			// Process the log line that we just obtained from kafka.
			err = worker.processLogLine(string(msg.Value))
			if err != nil {
				glog.Errorf("error processing log line: %v", err)
				continue
			}
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------

// processLogLine is a helper function to process a single line. This involves obtaining some stats and writing the
// stats to the postgres database.
func (worker *StatsWorker) processLogLine(logLine string) error {

	// Define the regular expression pattern.
	pattern := `(\d+):(\d+)::([\w-]+) (\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2},\d{3}) - (.*(?:\n.*)*)`

	// Compile the regular expression pattern.
	regex := regexp.MustCompile(pattern)

	// Find submatches within the log line
	matches := regex.FindStringSubmatch(logLine)

	// Extract the captured groups.
	processID := matches[1]
	threadID := matches[2]
	threadName := matches[3]
	loggedTime := matches[4]
	logMessage := strings.TrimSpace(matches[5])

	// Print the extracted information.
	glog.Infoln("Process ID: ", processID)
	glog.Infoln("Thread ID: ", threadID)
	glog.Infoln("Thread Name: ", threadName)
	glog.Infoln("Logged Time: ", loggedTime)
	glog.Infoln("Log Message: ", logMessage)

	// Parse the timestamp string to a time.Time value.
	// Note that this is the timestamp format in the log message.
	timestamp, err := time.Parse("2006-01-02 15:04:05.999", loggedTime)
	if err != nil {
		log.Printf("Failed to parse timestamp: %v", err)
		return nil
	}

	// Create a new LogLine object.
	logLineObj := &LogLine{
		ProcessID:        processID,
		ThreadID:         threadID,
		Timestamp:        timestamp.UTC(),
		TimestampSeconds: timestamp.Unix(),
		LogMessage:       logMessage,
	}

	// Insert the log line into the database.
	_, err = worker.db.Model(logLineObj).Insert()
	if err != nil {
		log.Printf("failed to insert log line: %v", err)
		return nil
	}

	return nil
}

//----------------------------------------------------------------------------------------------------------------------
