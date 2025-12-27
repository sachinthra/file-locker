#!/bin/bash

CONFIG_FILE="configs/config.yaml"
ENV_FILE=".env"

echo "Syncing configuration from $CONFIG_FILE to $ENV_FILE..."

# Extract values from YAML
# Note: simple grep/awk extraction. For complex YAML, 'yq' is better but this works for standard format.

SERVER_PORT=$(grep "port:" $CONFIG_FILE | grep -v "grpc" | head -1 | awk '{print $2}')
GRPC_PORT=$(grep "grpc_port:" $CONFIG_FILE | head -1 | awk '{print $2}')

# MinIO Ports
MINIO_PORT_API=$(grep "port_api:" $CONFIG_FILE | head -1 | awk '{print $2}')
MINIO_PORT_CONSOLE=$(grep "port_console:" $CONFIG_FILE | head -1 | awk '{print $2}')

# Redis Port
REDIS_PORT=$(grep "port:" $CONFIG_FILE | grep -v "server" | grep -v "grpc" | tail -1 | awk '{print $2}')

# Write to .env
echo "# Auto-generated from configs/config.yaml" > $ENV_FILE
echo "SERVER_PORT=$SERVER_PORT" >> $ENV_FILE
echo "GRPC_PORT=$GRPC_PORT" >> $ENV_FILE
echo "MINIO_PORT_API=$MINIO_PORT_API" >> $ENV_FILE
echo "MINIO_PORT_CONSOLE=$MINIO_PORT_CONSOLE" >> $ENV_FILE
echo "REDIS_PORT=$REDIS_PORT" >> $ENV_FILE

# Frontend Vars
echo "VITE_API_URL=http://localhost:$SERVER_PORT" >> $ENV_FILE
echo "VITE_GRPC_URL=http://localhost:$GRPC_PORT" >> $ENV_FILE

echo "✓ Sync Complete."
