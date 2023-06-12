// Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains utils related to configuration management.
//
// The root directory of this microservice has a file called defaults.yaml. All the configuration is maintained in this
// configuration file. Using this design practice we have can modify the configuration directly from the helm without
// needing to recompile the code. We do not use helm in this excercise. But this pattern is extensible.
//
// The file contains a method called LoadConfiguration() which will load all the configuration represented in the
// defaults.yaml and creates a golang object for that. This will be passed into all the other functions/classes in this
// microservice and this object will be a single place where all the configuration is all maintained.
//
// The file also creates constants like "KTopic" which represents the key in defaults.yaml which represents the kafka
// topic name.
//
// Any other package in this microservice, to refer to this configuration, it will simply do the following.
//
// import "eightfold/internal/config"
//
// func CreateKafkaConsumer(conf *viper.Viper) error {
//   topicName = conf.GetString(config.GetString(KKafkaTopic)))
// }
//
// Here conf is the configuration object that is created once and injected as dependency.

package config

import (
	"fmt"

	"github.com/spf13/viper"
)

const (

	// KGroupKeyLogProcessor is group key for log-processor block in defaults.yaml. This is the parent key. All the
	// nested children in this group can be referenced with this group key. For example defaults.yaml has something
	// like this.
	// log-processor:
	//  logs_directory: "/app/data"
	KGroupKeyLogProcessor = "log_processor"

	// KLogsDirectory is a nested key under the group key KGroupKeyLogWorker to obtain the logs_directory.
	KLogsDirectory = KGroupKeyLogProcessor + ".logs_directory"

	// KMaxFilesPerBatch is a nested key under the group key KGroupKeyLogWorker to obtain the max files per batch
	// configuration.
	KMaxFilesPerBatch = KGroupKeyLogProcessor + ".max_files_per_batch"

	// KMaxParallelLines s a nested key under the group key KGroupKeyLogWorker to obtain the max parallel lines.
	KMaxParallelLines = KGroupKeyLogProcessor + ".max_parallel_lines"

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Kafka related configuration.

	// KGroupKafka is group key for kafka block in defaults.yaml. This is the parent key. All the nested
	// children in this group can be referenced with this group key. For example defaults.yaml has something like this.
	// kafka:
	//  topic: "processor-messages"
	// To access the topic key it can be accessed as
	KGroupKafka = "kafka"

	// KBootstrapServers is a nested key under the group key KGroupKafka to obtain the kafka boostrap servers.
	KBootstrapServers = KGroupKafka + ".bootstrap_servers"

	// KTopic is a nested key under the group key KTopic to obtain the kafka topic name.
	KTopic = KGroupKafka + ".topic"

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
)

// LoadConfiguration is a helper function to load the configuration present in defaults.yaml. This will be loaded
// to a config object which then can be passed around(injected) to all the structs/classes to read the global
// configuration.
func LoadConfiguration() *viper.Viper {

	// Create a new Viper instance.
	config := viper.New()

	// Initialize Viper config
	config.SetConfigName("defaults")
	config.SetConfigType("yaml")
	config.AddConfigPath(".")

	// Read the configuration file.
	if err := config.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("failed to read config file: %v", err))
	}

	// At this point all the configuration present in defaults.yaml will be loaded into the config object.
	return config
}

//----------------------------------------------------------------------------------------------------------------------
