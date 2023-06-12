// Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains web server related methods.

package web

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"

	"apiserver/internal/config"
	"apiserver/internal/models"
	services "apiserver/internal/services"
)

// Server defines the struct that encapsulates all the necessary injections to start the web server.
type Server struct {
	// Echo instance.
	ec *echo.Echo

	// Interface object for stats services.
	statsService services.StatsServicer

	// The configuration object.
	conf *viper.Viper
}

// ---------------------------------------------------------------------------------------------------------------------

// NewWebServer returns new instance of WebServer.
func NewWebServer(ec *echo.Echo, statsService services.StatsServicer, conf *viper.Viper) *Server {
	ws := new(Server)
	ws.ec = ec
	ws.statsService = statsService
	ws.conf = conf
	return ws
}

//----------------------------------------------------------------------------------------------------------------------

// StartServer starts the Echo server.
func StartServer(conf *viper.Viper, statsService services.StatsServicer) {
	// Initialize Echo instance
	ec := echo.New()

	// Create the web server object.
	webServer := NewWebServer(ec, statsService, conf)

	// Middleware
	webServer.ec.Use(middleware.Logger())
	webServer.ec.Use(middleware.Recover())

	// Routes
	// Basic api requested in assignment.
	webServer.ec.GET("/basicStats", webServer.BasicStatsAPIHandler)

	// Bonus api 1 and 2 respectively.
	webServer.ec.GET("/maxConcurrentThreads", webServer.GetMaxConcurrentThreadsHandler)
	webServer.ec.GET("/threadLifetimeStats", webServer.GetThreadLifetimeStatsHandler)

	// Start web server.
	addr := fmt.Sprintf(":%d", conf.GetInt(config.KWebServerPort))
	glog.Infoln("Starting web server on port :", addr)
	webServer.ec.Logger.Fatal(webServer.ec.Start(addr))
}

//----------------------------------------------------------------------------------------------------------------------

// BasicStatsAPIHandler is a handler.
func (server *Server) BasicStatsAPIHandler(c echo.Context) error {

	// Parse the request body into the BasicLogStatsRequest struct
	req := new(models.BasicLogStatsRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}

	glog.Infoln("Received request for basic stats handler ", req)

	// Call the GetBasicStats method on the statsService
	resp, err := server.statsService.GetBasicStats(req)
	if err != nil {
		glog.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, "Failed to retrieve stats")
	}

	return c.JSON(http.StatusOK, resp)
}

//----------------------------------------------------------------------------------------------------------------------

// GetMaxConcurrentThreadsHandler handles the maxConcurrentThreads API.
func (server *Server) GetMaxConcurrentThreadsHandler(c echo.Context) error {
	// Call the GetMaxConcurrentThreads method on the statsService
	resp, err := server.statsService.GetMaxConcurrentThreads()
	if err != nil {
		glog.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, "Failed to retrieve max concurrent threads")
	}

	return c.JSON(http.StatusOK, resp)
}

//----------------------------------------------------------------------------------------------------------------------

// GetThreadLifetimeStatsHandler handles the threadLifetimeStats API.
func (server *Server) GetThreadLifetimeStatsHandler(c echo.Context) error {
	// Call the GetThreadLifetimeStats method on the statsService
	resp, err := server.statsService.GetThreadLifetimeStats()
	if err != nil {
		glog.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, "Failed to retrieve thread lifetime stats")
	}

	return c.JSON(http.StatusOK, resp)

}

//----------------------------------------------------------------------------------------------------------------------
