// Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains message queue related utils.

package messageq

import (
	"context"
	"log"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/glog"
	"github.com/spf13/viper"

	"logprocessor/internal/config"
)

// MaybeCreateKafkaTopic is a helper function to create a topic in message cluster. The topic
// will be created only if the topic does not exist.
func MaybeCreateKafkaTopic(conf *viper.Viper) error {

	// Read the broker config from the configuration.
	brokers := conf.GetString(config.KBootstrapServers)
	topic := conf.GetString(config.KTopic)

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		return err
	}
	defer adminClient.Close()

	// Check if the topic already exists.
	exists, err := topicExists(topic, adminClient)
	if err != nil {
		glog.Errorln(err.Error())
		return err
	}

	// If we reach here, the topic does not exist. Create one.
	if !exists {
		err = createKafkaTopic(topic, adminClient)
		if err != nil {
			glog.Errorln(err.Error())
			return err
		}

		glog.Infoln("Kafka topic", topic, "created successfully")
	} else {
		glog.Infoln("Kafka topic", topic, "already exists")
	}

	return nil
}

//---------------------------------------------------------------------------------------------------------------------

// CreateKafkaProducer creates and returns a new Kafka producer instance.
func CreateKafkaProducer(conf *viper.Viper) (*kafka.Producer, error) {

	brokers := conf.GetString(config.KBootstrapServers)
	producerConfig := &kafka.ConfigMap{"bootstrap.servers": brokers}
	producer, err := kafka.NewProducer(producerConfig)
	if err != nil {
		glog.Errorln(err.Error())
		return nil, err
	}
	return producer, nil
}

//----------------------------------------------------------------------------------------------------------------------

// PublishToKafka is a helper function which is run as a go routine per batch of files being processed. The function
// takes the following parameters.
//
// logLines : a buffered channel which is populated various go routines that is processing the files in a given batch.
func PublishToKafka(logLines chan string, producer *kafka.Producer, topic string) {

	// Please note that we are iterating over a buffered channel here. This is a blocking call. The go routine will
	// infinitely block until the next message is available in the buffered channel.
	for logLine := range logLines {
		// Parse the log line and extract the required fields
		parts := strings.Split(logLine, " - ")
		if len(parts) < 2 {
			log.Printf("Invalid log line: %s", logLine)
			continue
		}

		messageKey := parts[0]
		messageValue := logLine

		// Publish the log message to Kafka.
		err := producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(messageKey),
			Value:          []byte(messageValue),
		}, nil)

		if err != nil {
			log.Printf("Failed to produce message: %s", err.Error())
		}

		glog.Infoln("The message key: ", messageKey)
		glog.Infoln("The message value: ", messageValue)
	}
}

//----------------------------------------------------------------------------------------------------------------------

// topicExists is a helper function to check if the topic exists in the given kafka broker.
func topicExists(topic string, adminClient *kafka.AdminClient) (bool, error) {
	topics, err := adminClient.GetMetadata(&topic, false, 5000)
	if err != nil {
		return false, err
	}

	// Iterate over all the topics that are present and check if the topic exists.
	for _, t := range topics.Topics {
		if t.Topic == topic {
			return true, nil
		}
	}

	// If we reach here the topic does not exist.
	return false, nil
}

//----------------------------------------------------------------------------------------------------------------------

// createKafkaTopic is a helper function to create the kafka topic.
func createKafkaTopic(topic string, adminClient *kafka.AdminClient) error {
	topicConfig := &kafka.TopicSpecification{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	topics := []kafka.TopicSpecification{*topicConfig}
	_, err := adminClient.CreateTopics(context.Background(), topics)
	if err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------------------------------------------------
