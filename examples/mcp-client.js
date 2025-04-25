/**
 * Scansca MCP Client Example - JavaScript
 * 
 * This example demonstrates how to connect to the Scansca MCP server using Server-Sent Events (SSE)
 * and execute database operations.
 * 
 * Usage:
 *   1. Start the Scansca MCP server
 *   2. Run this script in a browser or Node.js environment with EventSource support
 */

// Server URL
const MCP_SERVER_URL = 'http://localhost:8080/sse';

// Connect to MCP SSE endpoint
const eventSource = new EventSource(MCP_SERVER_URL);

// Track the connection state
let isConnected = false;

// Event handlers
eventSource.addEventListener('open', () => {
  console.log('Connected to MCP server');
});

eventSource.addEventListener('error', (error) => {
  console.error('SSE Connection error:', error);
  isConnected = false;
});

// Handle connection setup events
eventSource.addEventListener('init', (event) => {
  console.log('MCP server initialized');
  const data = JSON.parse(event.data);
  console.log('Server info:', data);
  
  isConnected = true;
  
  // Discover available tools
  sendRequest({
    type: 'discover'
  });
});

// Handle tool discovery response
eventSource.addEventListener('discover_response', (event) => {
  const data = JSON.parse(event.data);
  console.log('Available tools:', data.tools);
  
  // Example: Execute a query
  executeQuery('postgres-main', 'SELECT * FROM users LIMIT 10');
});

// Handle streaming tool results
eventSource.addEventListener('stream', (event) => {
  const data = JSON.parse(event.data);
  
  if (data.status === 'started') {
    console.log('Query execution started');
  }
  else if (data.status === 'columns') {
    console.log('Columns:', data.columns);
    setupResultTable(data.columns);
  } 
  else if (data.status === 'data') {
    console.log(`Received ${data.count} rows`);
    appendRows(data.data);
  }
  else if (data.status === 'complete') {
    console.log(`Query complete. Total rows: ${data.total_rows}, Duration: ${data.duration}`);
    if (data.error) {
      console.error('Query error:', data.error);
    }
  }
  else if (data.status === 'error') {
    console.error('Error:', data.message);
  }
});

// Function to send requests to the server
function sendRequest(request) {
  if (!isConnected) {
    console.error('Not connected to MCP server');
    return;
  }
  
  // Using POST for requests to the SSE connection
  fetch(MCP_SERVER_URL, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(request)
  }).catch(error => {
    console.error('Request error:', error);
  });
}

// Execute a database query
function executeQuery(database, query, params = []) {
  console.log(`Executing query on ${database}: ${query}`);
  
  sendRequest({
    type: 'call',
    tool: 'execute_query',
    params: {
      database: database,
      query: query,
      params: params
    }
  });
}

// List database schemas
function listSchemas(database) {
  console.log(`Listing schemas in ${database}`);
  
  sendRequest({
    type: 'call',
    tool: 'list_schemas',
    params: {
      database: database
    }
  });
}

// List tables in a schema
function listTables(database, schema) {
  console.log(`Listing tables in ${database}.${schema}`);
  
  sendRequest({
    type: 'call',
    tool: 'list_tables',
    params: {
      database: database,
      schema: schema
    }
  });
}

// Helper functions for UI (adapt to your needs)
function setupResultTable(columns) {
  console.log('Setting up table with columns:', columns);
  // In a real application, create a table with the column headers
}

function appendRows(rows) {
  console.log('Appending rows:', rows.length);
  // In a real application, append these rows to your results table
}

// Close connection on page unload
window.addEventListener('beforeunload', () => {
  eventSource.close();
});