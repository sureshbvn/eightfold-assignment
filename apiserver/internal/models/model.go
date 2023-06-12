// Copyright 2023
//
// Author: Suresh Bysani
//
// The file contains the data model for rest apis for all the apis needed for the assignment.

package models

import "time"

// LogLines encapsulates the data model for the schema in the database.
type LogLines struct {
	tableName        struct{}  `pg:"log_lines"`
	ProcessID        string    `pg:"process_id,notnull,pk"`
	ThreadID         string    `pg:"thread_id,notnull,pk"`
	Timestamp        time.Time `pg:"timestamp,notnull,pk"`
	TimestampSeconds int64     `pg:"timestamp_seconds,notnull,pk"`
	LogMessage       string    `pg:"log_message"`
}

//----------------------------------------------------------------------------------------------------------------------
// The data model for the basic log stats api. This contains request and response.

type BasicLogStatsRequest struct {
	StartTimeSeconds int64 `json:"start_time_seconds"`
	EndTimeSeconds   int64 `json:"end_time_seconds"`
}

type BasicLogStatsResponse struct {
	ActiveThreadsCount int   `pg:"active_threads_count"`
	ActiveThreadIDs    []int `pg:"active_thread_ids,array"`
	ActiveProcessIDs   []int `pg:"active_process_ids,array"`
}

//----------------------------------------------------------------------------------------------------------------------
// The data model for the bonus api 1.

// MaxConcurrentThreadsResponse represents the response structure for the maximum concurrent threads API.
type MaxConcurrentThreadsResponse struct {
	ConcurrentThreads int64 `json:"concurrent_threads"`
	TimestampSeconds  int64 `json:"timestamp_seconds"`
}

//----------------------------------------------------------------------------------------------------------------------
// The data model for the bonus api 2.

// ThreadLifetimeStatsResponse represents the response structure for the thread lifetime statistics API.
type ThreadLifetimeStatsResponse struct {
	AverageLifetime float64 `json:"average_lifetime"`
	StdevLifetime   float64 `json:"stdev_lifetime"`
}

//----------------------------------------------------------------------------------------------------------------------
