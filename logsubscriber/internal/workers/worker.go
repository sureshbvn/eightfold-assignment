// Copyright 2023
//
// Author: Suresh Bysani
//
// The file contains the interface for the workers.
//
// The log subscriber microservice has two main workers.
//      1. File worker.
//      2. Stats worker.
//
// The log-subscriber services in general is responsible for reading the log lines from kafka. The file worker will
// read the log messages from kafka and write them to respective sanitized files. The stats worker on other hand
// will prepare stats need to get the answers for apis(including bonus APIs)

package workers

import (
	"context"
)

// Worker defines the interface for a worker.
type Worker interface {
	// Start the worker.
	Start(ctx context.Context) error

	// Stop the worker.
	Stop() error
}
