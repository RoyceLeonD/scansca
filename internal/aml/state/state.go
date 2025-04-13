package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

// StateManager handles persistent state management
type StateManager struct {
	stateFile string
	stateLock sync.RWMutex
	state     map[string]interface{}
}

// NewStateManager creates a new state manager
func NewStateManager(configDir, stateFileName string) (*StateManager, error) {
	// Ensure the config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	stateFile := filepath.Join(configDir, stateFileName)

	// Initialize the state manager
	manager := &StateManager{
		stateFile: stateFile,
		state:     make(map[string]interface{}),
	}

	// Load existing state if available
	if err := manager.Load(); err != nil {
		// If the file doesn't exist yet, that's not an error
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load state: %w", err)
		}
	}

	return manager, nil
}

// Load loads the state from the state file
func (s *StateManager) Load() error {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	// Initialize viper
	v := viper.New()
	v.SetConfigFile(s.stateFile)

	// Try different formats
	if filepath.Ext(s.stateFile) == ".json" {
		v.SetConfigType("json")
	} else if filepath.Ext(s.stateFile) == ".yaml" || filepath.Ext(s.stateFile) == ".yml" {
		v.SetConfigType("yaml")
	} else {
		// Default to JSON
		v.SetConfigType("json")
	}

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("could not read config file: %w", err)
	}

	// Get all settings
	allSettings := v.AllSettings()
	s.state = allSettings

	return nil
}

// Save saves the current state to the state file
func (s *StateManager) Save() error {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()

	// Create the directory if it doesn't exist
	dir := filepath.Dir(s.stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal the state to JSON
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to the file
	if err := os.WriteFile(s.stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// Get retrieves a value from the state
func (s *StateManager) Get(key string) (interface{}, bool) {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()

	value, exists := s.state[key]
	return value, exists
}

// GetString retrieves a string value from the state
func (s *StateManager) GetString(key string) (string, bool) {
	value, exists := s.Get(key)
	if !exists {
		return "", false
	}

	strValue, ok := value.(string)
	return strValue, ok
}

// Set stores a value in the state
func (s *StateManager) Set(key string, value interface{}) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	s.state[key] = value
}

// Delete removes a key from the state
func (s *StateManager) Delete(key string) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	delete(s.state, key)
}

// GetAll returns a copy of the entire state
func (s *StateManager) GetAll() map[string]interface{} {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()

	// Create a copy of the state to avoid race conditions
	stateCopy := make(map[string]interface{})
	for k, v := range s.state {
		stateCopy[k] = v
	}

	return stateCopy
}

// Clear removes all keys from the state
func (s *StateManager) Clear() {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	s.state = make(map[string]interface{})
}