# Docker file for postgres 13.
# Author: Suresh Bysani

# Use the official PostgreSQL image as the base image
FROM postgres:13

# Set environment variables
ENV POSTGRES_USER=suresh
ENV POSTGRES_PASSWORD=suresh
ENV POSTGRES_DB=olap

# Copy custom initialization scripts to the container
COPY init.sql /docker-entrypoint-initdb.d/

# Expose the PostgreSQL default port (5432)
EXPOSE 5432