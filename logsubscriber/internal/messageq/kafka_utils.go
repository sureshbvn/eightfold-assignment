// Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains message queue related utils.

package messageq

import (
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"logworker/internal/config"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// CreateKafkaConsumer is a helper function to create kafka consumer.
//
// config: The configuration object which contains details kafka brokers.
// consuerGroupId: The consumer group id for establishing the kafka consumer. In the logsubscriber the plan is to create
//                 two consumer groups. One consumer group to read the log lines and write them to files for sanitizied
//                 log viewing. The second consumer group is to prepare stats related to log line and write them in
//                 OLAP databases.
func CreateKafkaConsumer(conf *viper.Viper, consumerGroupId string) *kafka.Consumer {

	// Read the broker config from the configuration.
	brokers := conf.GetString(config.KBootstrapServers)

	// Kafka consumer configuration
	consumerConfig := &kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          consumerGroupId,
		"auto.offset.reset": "earliest",
	}

	// Create Kafka consumer
	consumer, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		glog.Fatal("Failed to create Kafka consumer:", err)
	}

	return consumer
}

//----------------------------------------------------------------------------------------------------------------------
