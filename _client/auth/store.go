// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InMemoryCredentialStore implements CredentialService with in-memory storage.
type InMemoryCredentialStore struct {
	mu          sync.RWMutex
	credentials map[string]*Credentials
	ttl         time.Duration
	timestamps  map[string]time.Time
}

var _ CredentialService = (*InMemoryCredentialStore)(nil)

// NewInMemoryCredentialStore creates a new in-memory credential store.
func NewInMemoryCredentialStore(ttl time.Duration) *InMemoryCredentialStore {
	store := &InMemoryCredentialStore{
		credentials: make(map[string]*Credentials),
		timestamps:  make(map[string]time.Time),
		ttl:         ttl,
	}

	// Start cleanup goroutine if TTL is set
	if ttl > 0 {
		go store.cleanupExpired()
	}

	return store
}

// GetCredentials retrieves credentials for a given context.
func (s *InMemoryCredentialStore) GetCredentials(ctx context.Context, contextID string) (*Credentials, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	creds, exists := s.credentials[contextID]
	if !exists {
		return nil, fmt.Errorf("credentials not found for context: %s", contextID)
	}

	// Check if credentials are expired
	if creds.IsExpired() {
		return nil, fmt.Errorf("credentials expired for context: %s", contextID)
	}

	// Check TTL expiration
	if s.ttl > 0 {
		if timestamp, exists := s.timestamps[contextID]; exists {
			if time.Since(timestamp) > s.ttl {
				return nil, fmt.Errorf("credentials TTL expired for context: %s", contextID)
			}
		}
	}

	return creds, nil
}

// StoreCredentials stores credentials for a given context.
func (s *InMemoryCredentialStore) StoreCredentials(ctx context.Context, contextID string, credentials *Credentials) error {
	if credentials == nil {
		return fmt.Errorf("credentials cannot be nil")
	}

	if contextID == "" {
		return fmt.Errorf("context ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.credentials[contextID] = credentials
	s.timestamps[contextID] = time.Now()

	return nil
}

// DeleteCredentials deletes credentials for a given context.
func (s *InMemoryCredentialStore) DeleteCredentials(ctx context.Context, contextID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.credentials, contextID)
	delete(s.timestamps, contextID)

	return nil
}

// ValidateCredentials validates if credentials are still valid.
func (s *InMemoryCredentialStore) ValidateCredentials(ctx context.Context, credentials *Credentials) error {
	if credentials == nil {
		return fmt.Errorf("credentials cannot be nil")
	}

	if !credentials.IsValid() {
		return fmt.Errorf("credentials are not valid")
	}

	return nil
}

// ListContexts returns all context IDs in the store.
func (s *InMemoryCredentialStore) ListContexts() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	contexts := make([]string, 0, len(s.credentials))
	for contextID := range s.credentials {
		contexts = append(contexts, contextID)
	}

	return contexts
}

// Size returns the number of credentials in the store.
func (s *InMemoryCredentialStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.credentials)
}

// Clear removes all credentials from the store.
func (s *InMemoryCredentialStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.credentials = make(map[string]*Credentials)
	s.timestamps = make(map[string]time.Time)
}

// cleanupExpired removes expired credentials from the store.
func (s *InMemoryCredentialStore) cleanupExpired() {
	ticker := time.NewTicker(s.ttl / 2) // Cleanup every half TTL
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()

		now := time.Now()
		for contextID, timestamp := range s.timestamps {
			if now.Sub(timestamp) > s.ttl {
				delete(s.credentials, contextID)
				delete(s.timestamps, contextID)
			}
		}

		// Also clean up expired credentials
		for contextID, creds := range s.credentials {
			if creds.IsExpired() {
				delete(s.credentials, contextID)
				delete(s.timestamps, contextID)
			}
		}

		s.mu.Unlock()
	}
}

// FileCredentialStore implements CredentialService with file-based storage.
// This is a placeholder implementation for future file-based storage.
type FileCredentialStore struct {
	filePath string
}

var _ CredentialService = (*FileCredentialStore)(nil)

// NewFileCredentialStore creates a new file-based credential store.
func NewFileCredentialStore(filePath string) *FileCredentialStore {
	return &FileCredentialStore{
		filePath: filePath,
	}
}

// GetCredentials retrieves credentials for a given context.
func (s *FileCredentialStore) GetCredentials(ctx context.Context, contextID string) (*Credentials, error) {
	// TODO: Implement file-based credential retrieval
	return nil, fmt.Errorf("file-based credential store not implemented")
}

// StoreCredentials stores credentials for a given context.
func (s *FileCredentialStore) StoreCredentials(ctx context.Context, contextID string, credentials *Credentials) error {
	// TODO: Implement file-based credential storage
	return fmt.Errorf("file-based credential store not implemented")
}

// DeleteCredentials deletes credentials for a given context.
func (s *FileCredentialStore) DeleteCredentials(ctx context.Context, contextID string) error {
	// TODO: Implement file-based credential deletion
	return fmt.Errorf("file-based credential store not implemented")
}

// ValidateCredentials validates if credentials are still valid.
func (s *FileCredentialStore) ValidateCredentials(ctx context.Context, credentials *Credentials) error {
	if credentials == nil {
		return fmt.Errorf("credentials cannot be nil")
	}

	if !credentials.IsValid() {
		return fmt.Errorf("credentials are not valid")
	}

	return nil
}

// NoOpCredentialStore implements CredentialService with no-op behavior.
type NoOpCredentialStore struct{}

var _ CredentialService = (*NoOpCredentialStore)(nil)

// NewNoOpCredentialStore creates a new no-op credential store.
func NewNoOpCredentialStore() *NoOpCredentialStore {
	return &NoOpCredentialStore{}
}

// GetCredentials always returns an error indicating no credentials.
func (s *NoOpCredentialStore) GetCredentials(ctx context.Context, contextID string) (*Credentials, error) {
	return nil, fmt.Errorf("no credentials available (no-op store)")
}

// StoreCredentials does nothing and returns nil.
func (s *NoOpCredentialStore) StoreCredentials(ctx context.Context, contextID string, credentials *Credentials) error {
	return nil
}

// DeleteCredentials does nothing and returns nil.
func (s *NoOpCredentialStore) DeleteCredentials(ctx context.Context, contextID string) error {
	return nil
}

// ValidateCredentials validates if credentials are still valid.
func (s *NoOpCredentialStore) ValidateCredentials(ctx context.Context, credentials *Credentials) error {
	if credentials == nil {
		return fmt.Errorf("credentials cannot be nil")
	}

	if !credentials.IsValid() {
		return fmt.Errorf("credentials are not valid")
	}

	return nil
}
