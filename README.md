# Assignment Structure

Author: Suresh Bysani

This comprehensive documentation provides an in-depth overview of the assignment, including the folder structure, component functionalities, and deployment configuration.

## Folder Structure

The assignment adheres to a well-organised folder structure to ensure clarity and maintainability. The table below outlines the key directories and files used in the assignment:

| Folder/File                | Description                                                       |
| -------------------------- | ----------------------------------------------------------------- |
| eightfold/                 | The main folder for the assignment, encompassing all project components |
| eightfold/data/            | The directory housing the assignment data                         |
| eightfold/data/input/      | Contains the input directory for raw log files                     |
| eightfold/data/sanitized/  | Contains the output directory for sanitized log files              |
| eightfold/logprocessor/    | Includes the code and Dockerfile for the Log Processor microservice |
| eightfold/logsubscriber/   | Contains the code and Dockerfile for the Log Subscriber microservice |
| eightfold/postgres/        | Encompasses the Dockerfile for the Postgres database and init.sql file |
| eightfold/apiserver/       | Houses the code for the API Server microservice                    |
| eightfold/docker-compose.yml | Provides the Docker Compose file for orchestrating the microservices and infrastructure |

## Components

The assignment comprises the following components, each serving a specific purpose within the system:

- **Log Processor**: Ingests log files, extracts lines, and writes them to Kafka.
- **Log Subscriber**: Consumes log lines from Kafka, sanitizes them, and persists them.
- **Postgres**: Stores log statistics for long-term storage.
- **API Server**: Exposes a RESTful API to access log statistics from Postgres.
- **Docker Compose**: Orchestrates the deployment of microservices and infrastructure.

## Appendix

### Testing Instructions

1. Download the project repository from the email.
2. Unzip the downloaded file to your desired location.
3. Open a terminal and navigate to the "eightfold" folder in the extracted project directory.
4. Ensure that Docker and Docker Compose are installed on your system. You can check their versions using the following commands:

   ```bash
   docker --version
   docker-compose --version
   ```
5. Make sure that the versions are compatible with the system requirements mentioned in the design document.

6. Run the following command to start the system:

   ```
   docker-compose up --build
   ```
   This command will build and start the microservices and infrastructure components defined in the Docker Compose file. Wait for the services to initialize and start running.

7. Once the services are up and running, you can use a tool like CURL or a web browser to make API calls to the API server and retrieve log statistics. Here are examples of the API endpoints:

  Basic Stats API:
  ```
  curl http://localhost:8080/basicStats?start=timestamp1&end=timestamp2
  ```
  
  Bonus API1:
  ```
  curl http://localhost:8080/maxConcurrentThreads
  ```
  
  Bonus API2:
  ```
  curl http://localhost:8080/threadLifetimeStats
  ```
  
### Development Environment

The system was developed using the following environment:

- Operating System: Ubuntu 22
- Go Version: 1.17
- Docker: 20.10.21, build 20.10.21-0ubuntu1~22.04.3
- Docker Compose: v2.18.1
