# Syntax version
# syntax=docker/dockerfile:1

# Build stage
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies first (for better caching)
COPY package*.json ./

# Clean install with overrides support
RUN rm -f package-lock.json && \
    npm install --legacy-peer-deps --no-audit && \
    npm cache clean --force

# Copy source code
COPY . .

# Build the application with production optimizations
ENV NODE_ENV=production
RUN npm run build

# Production stage with nginx
FROM nginx:1.25-alpine

# Install curl for healthcheck
RUN apk add --no-cache curl

# Copy custom nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Copy built files from builder
COPY --from=builder /app/build /usr/share/nginx/html

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /usr/share/nginx/html && \
    chown -R appuser:appuser /var/cache/nginx && \
    chown -R appuser:appuser /var/log/nginx && \
    touch /var/run/nginx.pid && \
    chown -R appuser:appuser /var/run/nginx.pid

# Update nginx.conf to run as non-root (if needed)
RUN sed -i 's/user  nginx;/user  appuser;/' /etc/nginx/nginx.conf || true

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:80/ || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
