// Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains main file for api server.
//
// The api server provides api to answer the apis required in the problem statement.
//
// This services starts a web server(using echo framework).
//
// It offers APIs that are required for the assignment.

package main

import (
	"flag"

	"github.com/golang/glog"

	"apiserver/internal/config"
	"apiserver/internal/db"
	services "apiserver/internal/services"
	"apiserver/internal/web"
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
	glog.Infoln("Starting API server...")

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Step (2): Load the configuration.
	conf := config.LoadConfiguration()

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Step (3): Create the database object and stats services.
	statsService := services.NewStatsService(db.NewDB(conf), conf)

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Step (4): Create the web server.
	// This will be a blocking call.
	web.StartServer(conf, statsService)
}

//----------------------------------------------------------------------------------------------------------------------
