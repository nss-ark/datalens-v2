# =============================================================================
# DataLens 2.0 — Frontend Production Dockerfile
# =============================================================================
# Multi-stage build: Build static assets → Serve with nginx
# =============================================================================

# --- Build Stage ---
FROM node:20-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY package.json package-lock.json ./
RUN npm ci

# Build
COPY . .
RUN npm run build

# --- Production Stage ---
FROM nginx:alpine

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Copy built static files
COPY --from=builder /app/dist /usr/share/nginx/html

# Create health check file
RUN echo '{"status":"healthy"}' > /usr/share/nginx/html/health.json

# Non-root user (nginx already runs as nginx user)
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
