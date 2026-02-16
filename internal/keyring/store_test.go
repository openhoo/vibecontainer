package keyring

import (
	"os"
	"testing"

	"github.com/openhoo/vibecontainer/internal/domain"
	"github.com/zalando/go-keyring"
)

func TestMain(m *testing.M) {
	keyring.MockInit()
	os.Exit(m.Run())
}

func TestStoreSaveAndLoad(t *testing.T) {
	// Use a test service name to avoid interfering with production credentials
	store := &Store{service: "vibecontainer-test"}

	// Clean up before test
	_ = store.Clear()
	defer func() {
		_ = store.Clear()
	}()

	// Test saving auth
	testAuth := domain.Auth{
		ClaudeOAuthToken: "test-oauth-token",
		AnthropicAPIKey:  "test-api-key",
		TunnelToken:      "test-tunnel-token",
	}

	err := store.SaveAuth(testAuth)
	if err != nil {
		t.Fatalf("SaveAuth failed: %v", err)
	}

	// Test loading auth
	loaded := store.LoadAuth()

	if loaded.ClaudeOAuthToken != testAuth.ClaudeOAuthToken {
		t.Errorf("ClaudeOAuthToken mismatch: got %q, want %q", loaded.ClaudeOAuthToken, testAuth.ClaudeOAuthToken)
	}
	if loaded.AnthropicAPIKey != testAuth.AnthropicAPIKey {
		t.Errorf("AnthropicAPIKey mismatch: got %q, want %q", loaded.AnthropicAPIKey, testAuth.AnthropicAPIKey)
	}
	if loaded.TunnelToken != testAuth.TunnelToken {
		t.Errorf("TunnelToken mismatch: got %q, want %q", loaded.TunnelToken, testAuth.TunnelToken)
	}

	// Test that empty values are not stored
	if loaded.OpenAIAPIKey != "" {
		t.Errorf("OpenAIAPIKey should be empty, got %q", loaded.OpenAIAPIKey)
	}
}

func TestStoreGetSet(t *testing.T) {
	store := &Store{service: "vibecontainer-test"}

	// Clean up before test
	_ = store.Clear()
	defer func() {
		_ = store.Clear()
	}()

	// Test Set and Get
	testKey := KeyOpenAIAPIKey
	testValue := "sk-test-1234567890"

	err := store.Set(testKey, testValue)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := store.Get(testKey)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got != testValue {
		t.Errorf("Get mismatch: got %q, want %q", got, testValue)
	}
}

func TestStoreDelete(t *testing.T) {
	store := &Store{service: "vibecontainer-test"}

	// Clean up before test
	_ = store.Clear()
	defer func() {
		_ = store.Clear()
	}()

	// Set a value
	testKey := KeyCodexAPIKey
	testValue := "test-codex-key"

	err := store.Set(testKey, testValue)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify it exists
	_, err = store.Get(testKey)
	if err != nil {
		t.Fatalf("Get failed after Set: %v", err)
	}

	// Delete it
	err = store.Delete(testKey)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err = store.Get(testKey)
	if err != keyring.ErrNotFound {
		t.Errorf("Get should return ErrNotFound after Delete, got %v", err)
	}
}

func TestStoreClear(t *testing.T) {
	store := &Store{service: "vibecontainer-test"}

	// Clean up before test
	_ = store.Clear()
	defer func() {
		_ = store.Clear()
	}()

	// Set multiple values
	testAuth := domain.Auth{
		ClaudeOAuthToken: "test-oauth",
		OpenAIAPIKey:     "test-openai",
		TunnelToken:      "test-tunnel",
	}

	err := store.SaveAuth(testAuth)
	if err != nil {
		t.Fatalf("SaveAuth failed: %v", err)
	}

	// Clear all
	err = store.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify all are gone
	loaded := store.LoadAuth()
	if loaded.ClaudeOAuthToken != "" || loaded.OpenAIAPIKey != "" || loaded.TunnelToken != "" {
		t.Errorf("Clear did not remove all credentials: %+v", loaded)
	}
}
