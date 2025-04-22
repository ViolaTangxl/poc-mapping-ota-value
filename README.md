# Derby Mapping Project

This project is a Go-based web application for Derby mapping functionality.

## Project Structure

- `main.go`: Main application entry point
- `internal/`: Internal application code
- `templates/`: HTML templates
- `utils/`: Utility functions
- `uploads/`: Directory for uploaded files

## Getting Started

1. Clone this repository
2. Copy `.env.example` to `.env` and fill in your credentials
3. Run `go mod download` to install dependencies
4. Run `go run main.go` to start the application

## Environment Variables

The application requires the following environment variables:

- `AWS_ACCESS_KEY`: Your AWS access key
- `AWS_SECRET_KEY`: Your AWS secret key
- `FEISHU_WEB_HOOK`: Feishu webhook URL for notifications
- `BEDROCK_REGION`: AWS Bedrock region (default: us-east-1)
- `CLAUDE_MODEL`: Claude model identifier

## Requirements

- Go 1.16+
