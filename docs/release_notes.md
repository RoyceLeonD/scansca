# Release Notes

## v0.0.1-alpha (April 2025)

This is the first alpha release of Antska, a Model Context Protocol (MCP) server for connecting LLMs to database systems.

### What's New

- **MCP Implementation**: Custom MCP server with Server-Sent Events (SSE) support for real-time data streaming
- **PostgreSQL Support**: Connect to PostgreSQL databases with full schema exploration and query capabilities
- **HTTP API**: RESTful endpoints for database operations and MCP tool discovery

### Known Limitations

- Authentication is not yet implemented - use in trusted environments only
- Only PostgreSQL is supported in this release
- Large result sets may cause performance issues - use query limits

### Coming Soon

- MySQL and SQLite support
- Authentication and authorization
- Query caching and optimization
- Dashboard for monitoring and administration