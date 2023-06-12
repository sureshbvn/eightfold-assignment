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

	// KGroupKeyApiServer is group key for apiserver block in defaults.yaml. This is the parent key. All the
	// nested children in this group can be referenced with this group key. For example defaults.yaml has something
	// like this.
	// db:
	//  host : "postgres"
	//  port : 5432
	KGroupKeyApiServer = "apiserver"

	// KWebServerPort is a nested key under the group key KGroupKeyApiServer to obtain the port for webserver.
	KWebServerPort = KGroupKeyApiServer + ".port"

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Database related configuration

	// KGroupDatabase is group key for db block in defaults.yaml. This is the parent key. All the nested
	// children in this group can be referenced with this group key. For example defaults.yaml has something like this.
	// db:
	//   host: postgres
	//   port: 5432
	//   username: suresh
	//   password: suresh
	//   database: olap
	KGroupDatabase = "db"

	// KHost is a nested key under the group key KGroupDatabase to obtain the hostname for the postgres database.
	KHost = KGroupDatabase + ".host"

	// KPort is a nested key under the group key KGroupDatabase to obtain the portname for the postgres database.
	KPort = KGroupDatabase + ".port"

	// KUsername is a nested key under the group key KGroupDatabase to obtain the username to connect to the postgres
	// database.
	KUsername = KGroupDatabase + ".username"

	// KPassword is a nested key under the group key KGroupDatabase to obtain the password to connect to the postgres
	// database.
	KPassword = KGroupDatabase + ".password"

	// KDatabaseName is a nested key under the group key KGroupDatabase to obtain the database name to connect to the postgres
	// database.
	KDatabaseName = KGroupDatabase + ".database"
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
