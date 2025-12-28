#!/bin/bash

CONFIG_FILE="configs/config.yaml"
ENV_FILE=".env"

echo "🔄 Syncing configuration from $CONFIG_FILE to $ENV_FILE..."
echo "   Making config.yaml the single source of truth..."

# Extract values from YAML
# Note: simple grep/awk extraction. For complex YAML, 'yq' is better but this works for standard format.

# Server Ports
SERVER_PORT=$(grep "port:" $CONFIG_FILE | grep -v "grpc" | grep -v "port_" | head -1 | awk '{print $2}')
GRPC_PORT=$(grep "grpc_port:" $CONFIG_FILE | head -1 | awk '{print $2}')

# Database Configuration
# Extract from the storage.database section
DB_USER=$(awk '/storage:/,/minio:/ {if (/user:/) print $2}' $CONFIG_FILE | tr -d '"')
DB_PASSWORD=$(awk '/storage:/,/minio:/ {if (/password:/) print $2}' $CONFIG_FILE | tr -d '"')
DB_NAME=$(awk '/storage:/,/minio:/ {if (/dbname:/) print $2}' $CONFIG_FILE | tr -d '"')
DB_PORT=$(awk '/storage:/,/minio:/ {if (/port:/ && !/port_/) print $2}' $CONFIG_FILE | head -1)

# MinIO Configuration
MINIO_PORT_API=$(grep "port_api:" $CONFIG_FILE | head -1 | awk '{print $2}')
MINIO_PORT_CONSOLE=$(grep "port_console:" $CONFIG_FILE | head -1 | awk '{print $2}')
MINIO_ACCESS_KEY=$(grep "access_key:" $CONFIG_FILE | awk '{print $2}' | tr -d '"')
MINIO_SECRET_KEY=$(grep "secret_key:" $CONFIG_FILE | awk '{print $2}' | tr -d '"')
MINIO_BUCKET=$(grep "bucket:" $CONFIG_FILE | awk '{print $2}' | tr -d '"')

# Redis Configuration
REDIS_PORT=$(grep "port:" $CONFIG_FILE | grep -v "server" | grep -v "grpc" | grep -v "port_" | tail -1 | awk '{print $2}')

# Security
JWT_SECRET=$(grep "jwt_secret:" $CONFIG_FILE | awk '{print $2}' | tr -d '"')

# Write to .env
cat > $ENV_FILE << EOF
# ================================================================
# Auto-generated from configs/config.yaml
# DO NOT EDIT MANUALLY - Run 'make sync-config' to regenerate
# ================================================================

# Server Configuration
SERVER_PORT=$SERVER_PORT
GRPC_PORT=$GRPC_PORT

# Database Configuration
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
DB_NAME=$DB_NAME
DB_PORT=$DB_PORT

# MinIO Configuration
MINIO_PORT_API=$MINIO_PORT_API
MINIO_PORT_CONSOLE=$MINIO_PORT_CONSOLE
MINIO_ACCESS_KEY=$MINIO_ACCESS_KEY
MINIO_SECRET_KEY=$MINIO_SECRET_KEY
MINIO_BUCKET=$MINIO_BUCKET

# Redis Configuration
REDIS_PORT=$REDIS_PORT

# Security
JWT_SECRET=$JWT_SECRET

# Frontend Configuration
VITE_API_URL=http://localhost:$SERVER_PORT
VITE_GRPC_URL=http://localhost:$GRPC_PORT
EOF

echo "✅ Sync Complete!"
echo ""
echo "📋 Summary:"
echo "   Database: $DB_USER @ $DB_NAME"
echo "   MinIO:    $MINIO_ACCESS_KEY (bucket: $MINIO_BUCKET)"
echo "   Ports:    HTTP=$SERVER_PORT, gRPC=$GRPC_PORT, MinIO=$MINIO_PORT_API, Redis=$REDIS_PORT"
echo ""
echo "💡 Next: Run 'docker compose up -d' to apply changes"
