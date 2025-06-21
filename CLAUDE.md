# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ToDo Info** is a Go web application that helps users analyze old tasks from Microsoft ToDo. It connects to Microsoft Graph API to fetch tasks, calculates task ages, and categorizes them by "rottenness" levels:
- 😊 Fresh (0-2 days)
- 😏 Ripe (3-6 days) 
- 🥱 Tired (7-13 days)
- 🤢 Zombie (14+ days)

The app displays metrics like total task ages, oldest tasks per list, and filtered views of old tasks.

## Development Commands

### Build and Run
```bash
go run cmd/web/main.go        # Run web server
go run cmd/example-page/main.go  # Run example page with sample data
go build -o todoinfo cmd/web/main.go  # Build binary
```

### Testing
```bash
go test ./...                 # Run all tests
go test ./internal/todometrics/  # Run specific package tests
go test -v ./internal/todometrics/  # Verbose test output
```

### Dependencies
```bash
go mod tidy                   # Clean up dependencies
go mod download               # Download dependencies
```

## Architecture

### Core Components

- **cmd/web/main.go**: Main web application entry point
- **cmd/example-page/main.go**: Example/demo application with sample data
- **internal/servers/**: HTTP server setup with Chi router, handles authentication middleware and route definitions
- **internal/todometrics/**: Core business logic for task age calculation and rottenness categorization
- **internal/todoclient/**: Microsoft Graph API client for fetching ToDo tasks
- **internal/login/**: OAuth2 authentication flow with Microsoft
- **internal/config/**: Environment-based configuration using .env files
- **internal/templates/**: HTML template system with embedded static assets

### Key Dependencies

- **Chi v5**: HTTP router and middleware
- **Gorilla Sessions**: Session management for OAuth tokens
- **Zerolog**: Structured logging
- **Microsoft Graph API**: Task data source via OAuth2

### Authentication Flow

1. User accesses protected route → redirected to `/auth`
2. OAuth2 flow to Microsoft with `User.Read Tasks.ReadWrite` scopes
3. Token stored in encrypted session cookie
4. Middleware validates token expiration on each request
5. Tasks fetched from Microsoft Graph API using stored token

### Required Environment Variables

Create a `.env` file with:
```
CLIENT_ID=your_microsoft_app_client_id
CLIENT_SECRET=your_microsoft_app_client_secret
HOST_URL=http://localhost:8080/
ADDR=:8080
SESSION_KEY=your_session_encryption_key
LOG_FOLDER=./logs
```

### Task Age Calculation

Task age is calculated using the most recent of `CreatedDateTime` or `LastModifiedDateTime`. Tasks can be skipped from age calculations by adding `#todo-info-skip` to the task note.