// Copyright 2023
//
// Author: Suresh Bysani
//
// This file contains the main file for the end-to-end test.
//
// The end-to-end test is responsible for verifying the functionality of the log sanitization system.
// It performs the following steps:
// 1. Checks if Docker and Docker Compose are installed.
// 2. Clean up if any previous installation of the setup is running. We do this by checking any containers running with
//    eightfold-assignment prefix.
// 3. Wait all the containers from previous installation are successfuly killed.
// 4. Starts the log sanitization system using Docker Compose.
// 5. Waits for the system to initialize.
// 6. Check for the existence of sanitized directory and make sure required number of files are created.
// 7. CheckPostgresData executes a psql command to check if there is non-zero data in the log_lines table.
// 8. Makes API calls to the API Server and checks their responses.
// 9. Stops the log sanitization system using Docker Compose.
// 10. Ensures that no containers with the "eightfold-assignment" prefix are running.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	// Step 1: Check if Docker and Docker Compose are installed
	checkDockerInstallation()
	checkDockerComposeInstallation()

	// Step 2: Bring down the log sanitization system
	stopLogSanitizationSystem()

	// Step 3: Wait until all containers with prefix "eightfold-assignment" are down
	waitForContainersToStop()

	// Step 4: Start the log sanitization system using Docker Compose
	startLogSanitizationSystem()

	// Step 5: Wait for the system to initialize
	waitForSystemInitialization()

	// Step 6: Check for the existence of sanitized directory and make sure required number of files are created.
	checkSanitizedDirectory()

	// Step 7: Make sure postgres has all the tables created.
	// checkPostgresData()

	// Step 6: Make API calls to the API Server and check their responses
	testAPIs()

	// Step 7: Stop the log sanitization system
	stopLogSanitizationSystem()

	// Step 8: Check if the containers are stopped and clean up
	checkContainerStatus()
}

//----------------------------------------------------------------------------------------------------------------------

// checkDockerInstallation makes sure if docker is installed on the host machine on which the test is running.
func checkDockerInstallation() {
	cmd := exec.Command("docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Docker is not installed:", err)
	}

	fmt.Println("Docker version:", string(output))
}

//----------------------------------------------------------------------------------------------------------------------

// checkDockerComposeInstallation makes sure if the docker-compose is installed on the host machine on which the test is
// running.
func checkDockerComposeInstallation() {
	cmd := exec.Command("docker-compose", "--version")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Docker Compose is not installed:", err)
	}

	fmt.Println("Docker Compose version:", string(output))
}

//----------------------------------------------------------------------------------------------------------------------

// stopLogSanitizationSystem is a helper function to bring down all the microservices related to assignment.
func stopLogSanitizationSystem() {
	cmd := exec.Command("docker-compose", "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal("Failed to stop the log sanitization system:", err)
	}

	fmt.Println("Log sanitization system stopped successfully")
}

//----------------------------------------------------------------------------------------------------------------------

// waitForContainersToStop is a helper function to make sure we wait until all the required containers are stopped.
func waitForContainersToStop() {
	for {
		cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
		output, err := cmd.Output()
		if err != nil {
			log.Fatal("Failed to retrieve container status:", err)
		}

		containers := strings.Split(string(output), "\n")
		allStopped := true
		for _, container := range containers {
			if strings.HasPrefix(container, "eightfold-assignment") {
				allStopped = false
				break
			}
		}

		if allStopped {
			break
		}

		fmt.Println("Waiting for containers to stop...")
		time.Sleep(5 * time.Second) // Wait for 5 seconds before checking again
	}
}

//----------------------------------------------------------------------------------------------------------------------

// startLogSanitizationSystem to start the project again.
func startLogSanitizationSystem() {
	cmd := exec.Command("docker-compose", "up", "--build", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal("Failed to start the log sanitization system:", err)
	}

	fmt.Println("Log sanitization system started successfully")
}

//----------------------------------------------------------------------------------------------------------------------

// checkSanitizedDirectory checks if the "../data/sanitized" directory is created and contains 250 files.
func checkSanitizedDirectory() {
	dirPath := "../data/sanitized"
	expectedFileCount := 250

	_, err := os.Stat(dirPath)
	if err != nil {
		log.Fatal("Sanitized directory does not exist:", err)
	}

	fileInfos, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatal("Failed to read directory:", err)
	}

	if len(fileInfos) != expectedFileCount {
		log.Fatalf("Unexpected file count in sanitized directory. Expected: %d, Actual: %d", expectedFileCount,
			len(fileInfos))
	}

	fmt.Println("Sanitized directory contains the expected number of files.")
}

//----------------------------------------------------------------------------------------------------------------------

// checkPostgresData executes a psql command to check if there is non-zero data in the log_lines table.
func checkPostgresData() {
	cmd := exec.Command("docker", "exec", "-it", "eightfold-assignment-postgres-1", "psql", "-U", "suresh",
		"-d", "olap", "-c", "select count(*) from log_lines")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Failed to execute psql command:", err)
	}

	count := strings.TrimSpace(string(output))
	fmt.Println("Number of records in log_lines table:", count)
}

//----------------------------------------------------------------------------------------------------------------------

// checkContainerStatus to make sure all containers are bought up.
func checkContainerStatus() {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Failed to retrieve container status:", err)
	}

	containers := strings.Split(string(output), "\n")
	for _, container := range containers {
		if strings.HasPrefix(container, "eightfold-assignment") {
			log.Fatalf("Test failed: Container %s is still running", container)
		}
	}

	fmt.Println("All containers have been stopped and cleaned up")
}

//----------------------------------------------------------------------------------------------------------------------

func waitForSystemInitialization() {
	// Add code here to wait for the system to initialize
	time.Sleep(10 * time.Second) // Wait for 10 seconds as an example
}

//----------------------------------------------------------------------------------------------------------------------

func testAPIs() {
	apiURLs := []string{
		"http://localhost:8080/basicStats",
		"http://localhost:8080/maxConcurrentThreads",
		"http://localhost:8080/threadLifetimeStats",
	}

	for _, url := range apiURLs {
		switch url {
		case "http://localhost:8080/basicStats":
			testBasicStatsAPI(url)
		case "http://localhost:8080/maxConcurrentThreads":
			testMaxConcurrentThreadsAPI(url)
		case "http://localhost:8080/threadLifetimeStats":
			testThreadLifetimeStatsAPI(url)
		}
	}
}

func testBasicStatsAPI(url string) {
	// Construct the API URL with query parameters
	apiURL := fmt.Sprintf("%s?start_time_seconds=%d&end_time_seconds=%d", url, 1496999565, 1696999565)

	// Send the API request
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Fatal("API request failed:", err)
	}
	defer resp.Body.Close()

	// Read and parse the API response
	var response BasicLogStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Fatal("Failed to decode API response:", err)
	}

	// Validate the response
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Basic Stats API test passed: %s\n", url)
	} else {
		fmt.Printf("Basic Stats API test failed: %s\n", url)
	}
}

func testMaxConcurrentThreadsAPI(url string) {
	// Send the API request
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("API request failed:", err)
	}
	defer resp.Body.Close()

	// Read and parse the API response
	var response MaxConcurrentThreadsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Fatal("Failed to decode API response:", err)
	}

	// Validate the response
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Max Concurrent Threads API test passed: %s\n", url)
		// Add additional validation logic based on the response data model
	} else {
		fmt.Printf("Max Concurrent Threads API test failed: %s\n", url)
	}
}

func testThreadLifetimeStatsAPI(url string) {
	// Send the API request
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("API request failed:", err)
	}
	defer resp.Body.Close()

	// Read and parse the API response
	var response ThreadLifetimeStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Fatal("Failed to decode API response:", err)
	}

	// Validate the response
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Thread Lifetime Stats API test passed: %s\n", url)
		// Add additional validation logic based on the response data model
	} else {
		fmt.Printf("Thread Lifetime Stats API test failed: %s\n", url)
	}
}

func encodeRequestBody(data interface{}) io.Reader {
	body, err := json.Marshal(data)
	if err != nil {
		log.Fatal("Failed to encode request body:", err)
	}
	return bytes.NewReader(body)
}

//----------------------------------------------------------------------------------------------------------------------
// Data model for HTTP requests
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
