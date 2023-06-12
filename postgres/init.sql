-- Connect to the olap database
\c olap

CREATE TABLE IF NOT EXISTS log_lines (
    process_id VARCHAR(255),
    thread_id VARCHAR(255),
    timestamp TIMESTAMPTZ,
    timestamp_seconds BIGINT,
    log_message TEXT,
    PRIMARY KEY (process_id, thread_id, timestamp, timestamp_seconds)
)