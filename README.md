# Web Page Analyzer

A Go application that analyzes web pages and provides detailed information about their structure, links, and forms.

## Features

- Analyzes HTML structure and version
- Counts headings and links
- Detects login forms
- Checks link accessibility
- Provides Prometheus metrics
- Beautiful web interface
- Docker support

## Prerequisites

- Go 1.22 or later
- Docker Desktop (for Docker deployment)
- Git (for dependency management)

## Installation

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/home24.git
   cd home24
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the application:
   ```bash
   make build
   ```

4. Run the application:
   ```bash
   make run
   ```

### Docker Deployment

1. Ensure Docker Desktop is running and you have sufficient disk space.

2. Build the Docker image:
   ```bash
   # Using make
   make docker-build
   
   # Or using docker directly
   docker build -t analyzer .
   ```

3. Run the container:
   ```bash
   # Using make
   make docker-run
   
   # Or using docker directly
   docker run -p 8080:8080 analyzer
   ```

4. Using docker-compose (recommended):
   ```bash
   # Build and start
   docker-compose up --build
   
   # Run in detached mode
   docker-compose up -d --build
   
   # Stop the container
   docker-compose down
   ```

#### Troubleshooting Docker Build

If you encounter issues during the Docker build process:

1. **go mod download fails**:
   ```bash
   # Clean Docker build cache
   docker builder prune
   
   # Ensure git is installed in the builder stage
   # Check your Dockerfile has: RUN apk add --no-cache git
   ```

2. **Permission issues**:
   ```bash
   # Fix permission issues
   sudo chown -R $USER:$USER .
   ```

3. **Network issues**:
   ```bash
   # Configure Docker to use a different DNS
   echo '{"dns": ["8.8.8.8", "8.8.4.4"]}' > /etc/docker/daemon.json
   sudo systemctl restart docker
   ```

4. **Build cache issues**:
   ```bash
   # Force a clean build
   docker build --no-cache -t analyzer .
   ```

## Configuration

The application can be configured through:

1. Environment variables:
   ```bash
   export CONFIG_PATH=/path/to/config.yaml
   ```

2. Configuration file (`config/application.yaml`):
   ```yaml
   server:
     port: "8080"
     readTimeout: "10s"
     writeTimeout: "30s"
     idleTimeout: "120s"

   analyzer:
     timeout: "30s"
     maxConcurrentLinks: 10
     userAgent: "WebPageAnalyzer/1.0"
     retryAttempts: 3
     maxLinksPerPage: 100
     maxDepth: 2
     enableMetrics: true
     metricsPrefix: "webpage_analyzer"
   ```

## Usage

1. Open your browser and navigate to `http://localhost:8080`
2. Enter a URL to analyze
3. View the analysis results, including:
   - HTML version
   - Page title
   - Heading structure
   - Link counts (internal/external)
   - Login form detection
   - Link accessibility

## Metrics

Prometheus metrics are available at `http://localhost:8080/metrics`:

- `webpage_analyzer_analysis_duration_seconds`: Analysis duration histogram
- `webpage_analyzer_requests_total`: Total analysis requests
- `webpage_analyzer_errors_total`: Total analysis errors
- `webpage_analyzer_link_counts`: Link counts by type
- `webpage_analyzer_heading_counts`: Heading counts by level
- `webpage_analyzer_login_forms_total`: Total login forms found
- `webpage_analyzer_html_versions_total`: HTML versions encountered

## Development

### Running Tests

```bash
make test
```

### Cleaning Build Artifacts

```bash
make clean
```

### Docker Development

1. Build the development image:
   ```bash
   docker-compose -f docker-compose.dev.yml build
   ```

2. Run the development container:
   ```bash
   docker-compose -f docker-compose.dev.yml up
   ```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.