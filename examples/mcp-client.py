#!/usr/bin/env python3
"""
Scansca MCP Client Example - Python

This example demonstrates how to connect to the Scansca MCP server using Server-Sent Events (SSE)
and execute database operations.

Requirements:
  - requests
  - sseclient

Installation:
  pip install requests sseclient-py

Usage:
  1. Start the Scansca MCP server
  2. Run this script: python mcp-client.py
"""

import json
import sys
import requests
import sseclient

# Server URL
MCP_SERVER_URL = 'http://localhost:8080/sse'

def connect_to_mcp(url):
    """Connect to MCP server and return SSE client"""
    print(f"Connecting to MCP server at {url}...")
    headers = {'Accept': 'text/event-stream'}
    response = requests.get(url, stream=True, headers=headers)
    if response.status_code != 200:
        print(f"Error connecting to server: {response.status_code}")
        sys.exit(1)
    return sseclient.SSEClient(response)

def send_request(url, request):
    """Send a request to the MCP server"""
    headers = {'Content-Type': 'application/json'}
    try:
        response = requests.post(url, json=request, headers=headers)
        if response.status_code != 200:
            print(f"Error sending request: {response.status_code}")
            return None
        return response.json()
    except Exception as e:
        print(f"Request error: {e}")
        return None

def execute_query(database, query, params=None):
    """Execute a database query"""
    if params is None:
        params = []
    
    print(f"Executing query on {database}: {query}")
    
    request = {
        'type': 'call',
        'tool': 'execute_query',
        'params': {
            'database': database,
            'query': query,
            'params': params
        }
    }
    
    return send_request(MCP_SERVER_URL, request)

def list_schemas(database):
    """List database schemas"""
    print(f"Listing schemas in {database}")
    
    request = {
        'type': 'call',
        'tool': 'list_schemas',
        'params': {
            'database': database
        }
    }
    
    return send_request(MCP_SERVER_URL, request)

def list_tables(database, schema):
    """List tables in a schema"""
    print(f"Listing tables in {database}.{schema}")
    
    request = {
        'type': 'call',
        'tool': 'list_tables',
        'params': {
            'database': database,
            'schema': schema
        }
    }
    
    return send_request(MCP_SERVER_URL, request)

def main():
    try:
        # Connect to SSE endpoint
        client = connect_to_mcp(MCP_SERVER_URL)
        
        # Send discovery request to get available tools
        print("Sending discovery request...")
        send_request(MCP_SERVER_URL, {'type': 'discover'})
        
        # Listen for events
        for event in client.events():
            data = json.loads(event.data)
            
            if event.event == 'init':
                print(f"Connected to MCP server: {data}")
            
            elif event.event == 'discover_response':
                print(f"Available tools:")
                for tool in data['tools']:
                    print(f"  - {tool['name']}: {tool['description']}")
                
                # Example: Execute a query (uncomment if you have a database configured)
                # database_name = 'postgres-main'
                # execute_query(database_name, 'SELECT * FROM users LIMIT 10')
            
            elif event.event == 'stream':
                if data['status'] == 'started':
                    print("Query execution started")
                elif data['status'] == 'columns':
                    print(f"Columns: {data['columns']}")
                elif data['status'] == 'data':
                    print(f"Received {len(data['data'])} rows")
                elif data['status'] == 'complete':
                    print(f"Query complete. Total rows: {data['total_rows']}, Duration: {data['duration']}")
                    if data.get('error'):
                        print(f"Query error: {data['error']}")
                elif data['status'] == 'error':
                    print(f"Error: {data['message']}")
            
            elif event.event == 'error':
                print(f"Error: {data}")
    
    except KeyboardInterrupt:
        print("\nExiting...")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    main()