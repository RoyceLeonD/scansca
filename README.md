# Antska - Database Monitoring and Management System

Antska is a database monitoring and management system designed to help track, analyze, and maintain your database infrastructure. The system provides automated scanning, reporting, and management capabilities for PostgreSQL databases.

## Features

- **Database Scanning**: Automatically scan databases to gather schema information, table statistics, and column details
- **Scheduled Jobs**: Schedule recurring scan and report jobs using cron expressions
- **Service Management**: Register and manage database services in a centralized system
- **Monitoring Dashboard**: (Planned) Web interface for viewing database metrics and reports

## Technical Stack

- **Backend**: Go (Golang)
- **Database**: PostgreSQL with pg_cron extension
- **Containerization**: Docker and Docker Compose

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose

### Setup

1. Clone the repository:
   ```
   git clone https://github.com/royceleond/antska.git
   cd antska
   ```

2. Start the PostgreSQL database:
   ```
   docker-compose -f docker/docker-compose.yaml up -d
   ```

3. Run the application:
   ```
   go run main.go
   ```

## Project Structure

- `main.go`: Application entry point
- `src/agent`: Core agent components
  - `dbManager`: Database connection handling
  - `scanner`: Database scanning functionality
  - `serviceManager`: Service and job management
- `src/dashboard`: (Planned) Web dashboard for monitoring
- `docker`: Docker configuration and SQL migrations

## License

[MIT License](LICENSE)

## Contact

For questions or feedback, please contact the project maintainer.