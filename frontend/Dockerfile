# ================================================================
# Multi-Stage Dockerfile for File Locker Frontend
# Stage 1: Build Preact application
# Stage 2: Serve with Nginx + Reverse Proxy
# Optimized for Raspberry Pi ARM64
# ================================================================

# ----------------------------------------------------------------
# STAGE 1: Builder
# ----------------------------------------------------------------
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files for dependency installation
COPY package.json package-lock.json ./

# Install all dependencies (including dev dependencies needed for build)
RUN npm ci && \
    npm cache clean --force

# Copy source code
COPY . .

# Build the Preact application
RUN npm run build

# ----------------------------------------------------------------
# STAGE 2: Nginx Runtime
# ----------------------------------------------------------------
FROM nginx:alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose port 80
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# Start nginx in foreground
CMD ["nginx", "-g", "daemon off;"]
