# Web Page Analyzer

A Go web application that analyzes web pages and provides information about HTML version, headings, links, and login forms.

## Features

- Analyzes web pages from URLs provided by users
- Detects HTML version
- Extracts page title
- Counts headings by level (h1, h2, etc.)
- Counts internal and external links
- Identifies inaccessible links
- Detects login forms
- Provides proper error handling for unreachable URLs

## Requirements

- Go 1.21 or later
- Docker (optional, for containerized deployment)

## Getting Started

### Local Development

1. Clone the repository
   ```
   git clone https://github.com/yourusername/web-page-analyzer.git
   cd web-page-analyzer
   ```

2. Run the application
   ```
   make run
   ```

3. Access the application in your browser at `http://localhost:8080`

### Using Docker

1. Build the Docker image
   ```
   make docker-build
   ```

2. Run the Docker container
   ```
   make docker-run
   ```

3. Access the application in your browser at `http://localhost:8080`

## Building

To build the binary:

```
make build
```

This will create a binary named `webanalyzer` in the project root.

## Testing

To run tests:

```
make test
```

## Design Decisions and Assumptions

1. **URL Validation**: The application validates that the URL starts with http:// or https:// to ensure proper analysis.

2. **HTML Version Detection**: HTML version is determined by analyzing the doctype declaration. For HTML5, both `<!DOCTYPE html>` and `<!DOCTYPE html 5>` are recognized.

3. **Link Classification**: Links are classified as internal if they point to the same host as the URL being analyzed, and external otherwise.

4. **Login Form Detection**: Login forms are detected by:
   - Looking for forms with "login" or "signin" in the action attribute
   - Checking for password input fields within forms

5. **Concurrency**: The application uses Go's concurrency features for link checking to improve performance.

6. **Error Handling**: The application provides meaningful error messages including HTTP status codes when URLs are unreachable.

## Future Improvements

1. **Enhanced Link Checking**: Implement actual HTTP requests to check link accessibility instead of estimation.

2. **Caching**: Add caching for previously analyzed URLs to improve performance.

3. **Metrics and Monitoring**: Integrate with Prometheus for metrics collection and monitoring.

4. **User Authentication**: Add user accounts to save analysis history.

5. **API Endpoint**: Create a REST API endpoint for programmatic access.

6. **Advanced Analysis**: Add more analysis features like SEO metrics, performance indicators, etc.

7. **Batch Processing**: Allow analysis of multiple URLs at once.

## License

This project is licensed under the MIT License - see the LICENSE file for details.