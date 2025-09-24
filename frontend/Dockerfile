# Frontend Dockerfile
FROM node:18-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies with optimizations
RUN npm ci --silent --no-audit --no-fund

# Copy source code
COPY . .

# Build the app
RUN npm run build

# Clean up npm cache to reduce image size
RUN npm cache clean --force

# Serve the app using a lightweight web server
FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
