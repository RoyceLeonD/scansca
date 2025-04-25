package mcp

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"encoding/json"
	"sync"
	"time"
)

// ConnectorRegistry defines the interface for a connector registry
type ConnectorRegistry interface {
	ListDatabases() []DatabaseConfig
	Create(dbType string) (Connector, error)
}

// MCPServer represents the MCP server implementation
type MCPServer struct {
	addr          string
	dbRegistry    ConnectorRegistry
	clients       sync.Map
	clientCounter int
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(host string, port int, registry ConnectorRegistry) *MCPServer {
	return &MCPServer{
		addr:         fmt.Sprintf("%s:%d", host, port),
		dbRegistry:   registry,
		clients:      sync.Map{},
	}
}

// RegisterTools registers all tools with the MCP server
func (s *MCPServer) RegisterTools() error {
	// This is a simplified version - no actual registration needed
	// since we're using direct HTTP handlers
	return nil
}

// Start starts the MCP server
func (s *MCPServer) Start() error {
	mux := http.NewServeMux()
	
	// SSE endpoint for tool events
	mux.HandleFunc("/sse", s.handleSSE)
	
	// HTTP endpoints for tool calls
	mux.HandleFunc("/api/v1/tools/execute_query", s.handleExecuteQueryHTTP)
	mux.HandleFunc("/api/v1/tools/list_schemas", s.handleListSchemasHTTP)
	mux.HandleFunc("/api/v1/tools/list_tables", s.handleListTablesHTTP)
	mux.HandleFunc("/api/v1/tools/get_table_info", s.handleGetTableInfoHTTP)
	
	// Tool discovery endpoint
	mux.HandleFunc("/api/v1/tools", s.handleToolDiscoveryHTTP)
	
	log.Printf("Starting MCP server on %s\n", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

// Shutdown stops the MCP server
func (s *MCPServer) Shutdown() error {
	// Close all SSE connections
	s.clients.Range(func(key, value interface{}) bool {
		if flusher, ok := value.(http.Flusher); ok {
			flusher.Flush()
		}
		return true
	})
	return nil
}

// SSE Connection handling

// handleSSE handles SSE connections
func (s *MCPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	// Check if the response writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}
	
	// Assign a client ID
	clientID := s.nextClientID()
	s.clients.Store(clientID, flusher)
	defer s.clients.Delete(clientID)
	
	// Send init event
	initEvent := map[string]interface{}{
		"clientId": clientID,
		"server":   "Scansca MCP Server",
		"version":  "1.0.0",
		"time":     time.Now().Format(time.RFC3339),
	}
	
	s.sendSSEEvent(w, flusher, "init", initEvent)
	
	// Keep the connection open
	closeNotify := r.Context().Done()
	for {
		select {
		case <-closeNotify:
			// Client disconnected
			return
		case <-time.After(30 * time.Second):
			// Send heartbeat every 30 seconds
			s.sendSSEEvent(w, flusher, "heartbeat", map[string]interface{}{
				"time": time.Now().Format(time.RFC3339),
			})
		}
	}
}

// nextClientID generates a unique client ID
func (s *MCPServer) nextClientID() int {
	s.clientCounter++
	return s.clientCounter
}

// sendSSEEvent sends an SSE event to the client
func (s *MCPServer) sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling SSE event data: %v", err)
		return
	}
	
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, dataBytes)
	flusher.Flush()
}

// HTTP handlers for tool calls

// handleExecuteQueryHTTP handles execute_query tool calls
func (s *MCPServer) handleExecuteQueryHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Read and parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}
	
	var params ExecuteQueryRequest
	if err := json.Unmarshal(body, &params); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request format: %v", err), http.StatusBadRequest)
		return
	}
	
	// Execute query
	result, err := s.handleExecuteQuery(r.Context(), body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing query: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"result": result,
	})
}

// handleListSchemasHTTP handles list_schemas tool calls
func (s *MCPServer) handleListSchemasHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Read and parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}
	
	var params ListSchemasRequest
	if err := json.Unmarshal(body, &params); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request format: %v", err), http.StatusBadRequest)
		return
	}
	
	// List schemas
	result, err := s.handleListSchemas(r.Context(), body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing schemas: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"result": result,
	})
}

// handleListTablesHTTP handles list_tables tool calls
func (s *MCPServer) handleListTablesHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Read and parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}
	
	var params ListTablesRequest
	if err := json.Unmarshal(body, &params); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request format: %v", err), http.StatusBadRequest)
		return
	}
	
	// List tables
	result, err := s.handleListTables(r.Context(), body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing tables: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"result": result,
	})
}

// handleGetTableInfoHTTP handles get_table_info tool calls
func (s *MCPServer) handleGetTableInfoHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Read and parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}
	
	var params TableInfoRequest
	if err := json.Unmarshal(body, &params); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request format: %v", err), http.StatusBadRequest)
		return
	}
	
	// Get table info
	result, err := s.handleGetTableInfo(r.Context(), body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting table info: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"result": result,
	})
}

// handleToolDiscoveryHTTP returns information about available tools
func (s *MCPServer) handleToolDiscoveryHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Define available tools
	tools := []map[string]interface{}{
		{
			"name":        "execute_query",
			"description": "Execute a SQL query on a registered database",
			"parameters": map[string]interface{}{
				"database": map[string]string{
					"type":        "string",
					"description": "Database name",
				},
				"query": map[string]string{
					"type":        "string",
					"description": "SQL query to execute",
				},
				"params": map[string]string{
					"type":        "array",
					"description": "Query parameters (optional)",
				},
			},
		},
		{
			"name":        "list_schemas",
			"description": "List all schemas in a database",
			"parameters": map[string]interface{}{
				"database": map[string]string{
					"type":        "string",
					"description": "Database name",
				},
			},
		},
		{
			"name":        "list_tables",
			"description": "List all tables in a database schema",
			"parameters": map[string]interface{}{
				"database": map[string]string{
					"type":        "string",
					"description": "Database name",
				},
				"schema": map[string]string{
					"type":        "string",
					"description": "Schema name",
				},
			},
		},
		{
			"name":        "get_table_info",
			"description": "Get detailed information about a database table",
			"parameters": map[string]interface{}{
				"database": map[string]string{
					"type":        "string",
					"description": "Database name",
				},
				"schema": map[string]string{
					"type":        "string",
					"description": "Schema name",
				},
				"table": map[string]string{
					"type":        "string",
					"description": "Table name",
				},
			},
		},
	}
	
	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"tools":  tools,
	})
}