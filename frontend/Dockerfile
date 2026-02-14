# =============================================================================
# DataLens 2.0 — Frontend Monorepo Build
# =============================================================================
# Builds all 3 applications: Control Centre, Admin, Portal
# Serves static assets via Nginx
# =============================================================================

# --- Stage 1: Builder ---
FROM node:20-alpine AS builder

WORKDIR /app

# Enable pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Copy workspace configs
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/control-centre/package.json ./apps/control-centre/
COPY apps/admin/package.json ./apps/admin/
COPY apps/portal/package.json ./apps/portal/
COPY apps/shared/package.json ./apps/shared/

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy source code
COPY . .

# Build all applications
RUN pnpm build -w @datalens/shared
RUN pnpm build -w @datalens/control-centre
RUN pnpm build -w @datalens/admin
RUN pnpm build -w @datalens/portal

# --- Stage 2: Production Server (Nginx) ---
FROM nginx:alpine

# Remove default Nginx website
RUN rm -rf /usr/share/nginx/html/*

# Copy build artifacts to separate directories
COPY --from=builder /app/apps/control-centre/dist /usr/share/nginx/cc
COPY --from=builder /app/apps/admin/dist /usr/share/nginx/admin
COPY --from=builder /app/apps/portal/dist /usr/share/nginx/portal

# Copy custom Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]
