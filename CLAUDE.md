# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ToDo Info** is a beautiful Go CLI application that helps users analyze old tasks from Microsoft ToDo. It connects to Microsoft Graph API to fetch tasks, calculates task ages, and categorizes them by "rottenness" levels:
- 😊 Fresh (0-2 days)
- 😏 Ripe (3-6 days) 
- 🥱 Tired (7-13 days)
- 🤢 Zombie (14+ days)

The CLI displays beautiful statistics with colored output, progress bars, charts, and comprehensive task analysis.

## Development Commands

### Build and Run
```bash
go run cmd/cli/main.go       # Run CLI application (shows help)
go build -o todoinfo cmd/cli/main.go # Build binary
./todoinfo --help           # Show CLI help
```

### CLI Usage
```bash
# Authentication
./todoinfo login
./todoinfo logout
./todoinfo status

# Task Analysis
./todoinfo stats                    # Live data (requires authentication)
./todoinfo stats --offline          # Offline mode using stored historical data
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

- **cmd/cli/main.go**: Main CLI application entry point
- **internal/cli/**: Modern CLI interface with Cobra framework and beautiful output
- **internal/auth/**: Azure AD authentication with token caching for CLI
- **internal/todometrics/**: Core business logic for task age calculation and rottenness categorization
- **internal/todoclient/**: Microsoft Graph SDK client for fetching ToDo tasks
- **internal/config/**: Environment-based configuration using .env files
- **internal/todo/**: Legacy task structures for compatibility

### Key Dependencies

- **Cobra**: Modern CLI framework with commands and flags
- **Lipgloss**: Beautiful terminal styling and formatting with colors, borders, and tables
- **NimbleMarkets ntcharts**: Terminal charts library for line charts and data visualization
- **Viper**: Configuration management with multiple sources
- **Azure SDK**: Official Azure authentication and Graph API client
- **Microsoft Graph SDK**: Official Microsoft Graph API SDK for Go
- **Zerolog**: Structured logging

### Authentication Flow

1. User runs `todoinfo login --client-id CLIENT_ID`
2. Browser-based OAuth2 flow to Microsoft with `User.Read Tasks.ReadWrite` scopes
3. Token cached locally in `.azure-cli-cache` directory
4. Subsequent commands use cached token automatically
5. Token refreshed automatically when expired
6. Use `todoinfo logout` to clear cached credentials

### Required Configuration

Option 1 - `.env` file (recommended):
```bash
# Create .env file in project root
echo "AZURE_CLIENT_ID=your_azure_client_id" > .env
./todoinfo stats
```

Option 2 - Command line flag:
```bash
./todoinfo stats --client-id YOUR_AZURE_CLIENT_ID
```

Option 3 - Environment variable:
```bash
export AZURE_CLIENT_ID=your_azure_client_id
./todoinfo stats
```

Option 4 - Config file `~/.todoinfo.yaml`:
```yaml
client-id: your_azure_client_id
```

### Setting up Azure App Registration

1. Go to Azure Portal → App Registrations
2. Create new registration with redirect URI: `http://localhost:8080`
3. Note the Application (client) ID
4. Grant `Tasks.ReadWrite` and `User.Read` permissions
5. Use the client ID with todoinfo CLI

### Task Age Calculation

Task age is calculated using the most recent of `CreatedDateTime` or `LastModifiedDateTime`. Tasks can be skipped from age calculations by adding `#todo-info-skip` to the task note.

### UI and Visualization

The CLI uses **Charmbracelet Lipgloss** for beautiful terminal styling including:
- Colored ASCII art banner with gradient effects
- Styled tables with rounded borders and proper alignment
- Consistent color scheme throughout the interface
- Boxed content with proper spacing and margins

Historical data visualization uses **NimbleMarkets ntcharts** for:
- Beautiful braille-based line charts showing task age trends over time
- Task count trends with linear scaling for accurate data interpretation
- Interactive charts with proper axes labels and data range information
- 90-day historical tracking stored in `~/.todoinfo/data/` as JSON snapshots

### Offline Mode

The CLI includes an `--offline` flag for using stored data without fetching new statistics:
```bash
./todoinfo stats --offline
```
This mode uses existing historical data from `~/.todoinfo/data/` without requiring authentication, perfect for:
- Viewing previously stored statistics without API calls
- Testing chart functionality with real data
- Development and debugging
- Quick access to task analysis when offline