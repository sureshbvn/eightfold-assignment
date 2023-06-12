// Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains stats services. This contains the business logic of all the apis defined by the api server.

package services

import (
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/golang/glog"
	"github.com/spf13/viper"

	"apiserver/internal/models"
)

type StatsServicer interface {
	// GetBasicStats contains the business logic to get the basic stats.
	GetBasicStats(request *models.BasicLogStatsRequest) (*models.BasicLogStatsResponse, error)

	// GetMaxConcurrentThreads retrieves the highest count of concurrent threads and the corresponding timestamp.
	GetMaxConcurrentThreads() (*models.MaxConcurrentThreadsResponse, error)

	// GetThreadLifetimeStats retrieves the average and standard deviation of thread lifetimes.
	GetThreadLifetimeStats() (*models.ThreadLifetimeStatsResponse, error)
}

// StatsService provides the business logic for retrieving log statistics
type StatsService struct {
	// The go-pg object.
	DB *pg.DB

	// The viper configuration object.
	conf *viper.Viper
}

// NewStatsService creates a new instance of StatsService
func NewStatsService(db *pg.DB, conf *viper.Viper) *StatsService {
	return &StatsService{
		DB:   db,
		conf: conf,
	}
}

//----------------------------------------------------------------------------------------------------------------------

// GetBasicStats retrieves basic log statistics within the specified time range.
func (s *StatsService) GetBasicStats(request *models.BasicLogStatsRequest) (*models.BasicLogStatsResponse, error) {

	glog.Infoln("fetching basic stats from log_lines table")

	var result models.BasicLogStatsResponse

	query := s.DB.Model((*models.LogLines)(nil)).
		ColumnExpr("COUNT(DISTINCT thread_id) AS active_threads_count").
		ColumnExpr("ARRAY_AGG(DISTINCT thread_id) AS active_thread_ids").
		ColumnExpr("ARRAY_AGG(DISTINCT process_id) AS active_process_ids").
		Where("timestamp_seconds >= ?", request.StartTimeSeconds).
		Where("timestamp_seconds <= ?", request.EndTimeSeconds)

	err := query.Select(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve basic stats: %v", err)
	}

	glog.Infoln(result)
	return &result, nil
}

//----------------------------------------------------------------------------------------------------------------------

// GetMaxConcurrentThreads retrieves the highest count of concurrent threads and the corresponding timestamp.
func (s *StatsService) GetMaxConcurrentThreads() (*models.MaxConcurrentThreadsResponse, error) {
	glog.Infoln("Fetching max concurrent threads from log_lines table")

	var result models.MaxConcurrentThreadsResponse

	err := s.DB.Model((*models.LogLines)(nil)).
		ColumnExpr("COUNT(DISTINCT thread_id) AS concurrent_threads").
		Column("timestamp_seconds").
		Group("timestamp_seconds").
		Order("concurrent_threads DESC").
		Limit(1).
		Select(&result)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve max concurrent threads: %v", err)
	}

	glog.Infoln(result)
	return &result, nil
}

//----------------------------------------------------------------------------------------------------------------------

// GetThreadLifetimeStats retrieves the average and standard deviation of thread lifetimes.
func (s *StatsService) GetThreadLifetimeStats() (*models.ThreadLifetimeStatsResponse, error) {
	glog.Infoln("Fetching thread lifetime stats from log_lines table")

	var result models.ThreadLifetimeStatsResponse

	query := fmt.Sprintf(`
        WITH thread_lifetimes AS (
            SELECT MAX(timestamp_seconds) - MIN(timestamp_seconds) AS lifetime
            FROM log_lines
            GROUP BY thread_id
        )
        SELECT AVG(lifetime) AS average_lifetime, STDDEV(lifetime) AS stdev_lifetime
        FROM thread_lifetimes
    `)

	_, err := s.DB.QueryOne(&result, query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve thread lifetime stats: %v", err)
	}

	glog.Infoln(result)
	return &result, nil
}

//----------------------------------------------------------------------------------------------------------------------
