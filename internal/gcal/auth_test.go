package gcal

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/oauth2"
)

func TestReadOauthClientID(t *testing.T) {
	// Test with non-existent file
	t.Run("non-existent file returns error", func(t *testing.T) {
		_, err := ReadOauthClientID("/nonexistent/path/credentials.json")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	// Test with invalid JSON
	t.Run("invalid JSON returns error", func(t *testing.T) {
		// Create a temporary file with invalid JSON
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "invalid.json")

		err := os.WriteFile(tmpFile, []byte("invalid json content"), 0600)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		_, err = ReadOauthClientID(tmpFile)
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})

	// Test with valid OAuth client JSON structure
	t.Run("valid OAuth JSON is parsed", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "credentials.json")

		// Minimal valid OAuth client JSON structure
		validJSON := `{
			"installed": {
				"client_id": "test-client-id",
				"client_secret": "test-client-secret",
				"redirect_uris": ["http://localhost"],
				"auth_uri": "https://accounts.google.com/o/oauth2/auth",
				"token_uri": "https://oauth2.googleapis.com/token"
			}
		}`

		err := os.WriteFile(tmpFile, []byte(validJSON), 0600)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		config, err := ReadOauthClientID(tmpFile)
		if err != nil {
			t.Errorf("Expected no error for valid JSON, got: %v", err)
		}

		if config == nil {
			t.Error("Expected non-nil config")
		}

		if config != nil && config.ClientID != "test-client-id" {
			t.Errorf("Expected client ID 'test-client-id', got '%s'", config.ClientID)
		}
	})
}

func TestTokenFromFile(t *testing.T) {
	t.Run("non-existent file returns error", func(t *testing.T) {
		_, err := tokenFromFile("/nonexistent/token.json")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "token.json")

		err := os.WriteFile(tmpFile, []byte("invalid json"), 0600)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		_, err = tokenFromFile(tmpFile)
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})

	t.Run("valid token JSON is parsed", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "token.json")

		validTokenJSON := `{
			"access_token": "test-access-token",
			"token_type": "Bearer",
			"refresh_token": "test-refresh-token",
			"expiry": "2026-12-31T23:59:59Z"
		}`

		err := os.WriteFile(tmpFile, []byte(validTokenJSON), 0600)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		token, err := tokenFromFile(tmpFile)
		if err != nil {
			t.Errorf("Expected no error for valid token JSON, got: %v", err)
		}

		if token == nil {
			t.Error("Expected non-nil token")
		}

		if token != nil && token.AccessToken != "test-access-token" {
			t.Errorf("Expected access token 'test-access-token', got '%s'", token.AccessToken)
		}
	})
}

func TestSaveToken(t *testing.T) {
	t.Run("saves token to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "token.json")

		token := &oauth2.Token{
			AccessToken:  "test-access-token",
			TokenType:    "Bearer",
			RefreshToken: "test-refresh-token",
		}

		err := saveToken(tmpFile, token)
		if err != nil {
			t.Errorf("Expected no error saving token, got: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
			t.Error("Expected token file to be created")
		}

		// Verify file permissions (should be 0600)
		info, err := os.Stat(tmpFile)
		if err != nil {
			t.Fatalf("Failed to stat token file: %v", err)
		}

		mode := info.Mode().Perm()
		if mode != 0600 {
			t.Errorf("Expected file permissions 0600, got %o", mode)
		}
	})

	t.Run("invalid path returns error", func(t *testing.T) {
		token := &oauth2.Token{
			AccessToken: "test-token",
		}

		err := saveToken("/invalid/nonexistent/path/token.json", token)
		if err == nil {
			t.Error("Expected error for invalid path, got nil")
		}
	})
}

func TestRefreshToken(t *testing.T) {
	t.Run("no refresh token returns error", func(t *testing.T) {
		config := &oauth2.Config{}
		token := &oauth2.Token{
			AccessToken: "test-token",
			// No RefreshToken
		}

		_, err := refreshToken(config, token)
		if err == nil {
			t.Error("Expected error when refresh token is missing, got nil")
		}
	})
}
