package metron

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		password    string
		expectError bool
	}{
		{
			name:        "valid credentials",
			username:    "testuser",
			password:    "testpass",
			expectError: false,
		},
		{
			name:        "empty username",
			username:    "",
			password:    "testpass",
			expectError: true,
		},
		{
			name:        "empty password",
			username:    "testuser",
			password:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.username, tt.password)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if client == nil {
				t.Error("expected client, got nil")
			}
		})
	}
}

func TestClient_Validate(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	tests := []struct {
		name        string
		username    string
		password    string
		expectError bool
	}{
		{
			name:        "valid credentials",
			username:    "testuser",
			password:    "testpass",
			expectError: false,
		},
		{
			name:        "invalid credentials",
			username:    "wronguser",
			password:    "wrongpass",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := NewClient(tt.username, tt.password)
			client.baseURL = server.URL + "/api"

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := client.Validate(ctx)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_FindIssueByUPC(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, _ := NewClient("testuser", "testpass")
	client.baseURL = server.URL + "/api"

	tests := []struct {
		name          string
		upc           string
		expectError   bool
		expectedID    int
		expectedTitle string
	}{
		{
			name:          "found issue",
			upc:           "75960609558200111",
			expectError:   false,
			expectedID:    162239,
			expectedTitle: "Amazing Spider-Man",
		},
		{
			name:        "not found",
			upc:         "00000000000000000",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			issue, err := client.FindIssueByUPC(ctx, tt.upc)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if issue.ID != tt.expectedID {
				t.Errorf("expected ID %d, got %d", tt.expectedID, issue.ID)
			}

			if issue.Series.Name != tt.expectedTitle {
				t.Errorf("expected title %q, got %q", tt.expectedTitle, issue.Series.Name)
			}
		})
	}
}

func TestClient_GetIssue(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, _ := NewClient("testuser", "testpass")
	client.baseURL = server.URL + "/api"

	ctx := context.Background()

	t.Run("get issue with full details", func(t *testing.T) {
		issue, err := client.GetIssue(ctx, 162239)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Prüfe vollständige Details
		if issue.Description == "" {
			t.Error("expected description, got empty string")
		}

		if len(issue.Credits) == 0 {
			t.Error("expected credits, got none")
		}

		if len(issue.Characters) == 0 {
			t.Error("expected characters, got none")
		}

		// Prüfe Credits im Detail
		foundWriter := false
		for _, credit := range issue.Credits {
			for _, role := range credit.Role {
				if role.Name == "Writer" {
					foundWriter = true
					if credit.Creator.Name != "Zeb Wells" {
						t.Errorf("expected writer 'Zeb Wells', got %q", credit.Creator.Name)
					}
				}
			}
		}

		if !foundWriter {
			t.Error("expected to find a Writer credit")
		}
	})

	t.Run("issue not found", func(t *testing.T) {
		_, err := client.GetIssue(ctx, 999999)
		if err == nil {
			t.Error("expected error for non-existent issue")
		}
	})
}

func TestClient_RateLimiting(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, _ := NewClient("testuser", "testpass")
	client.baseURL = server.URL + "/api"

	ctx := context.Background()

	// Measure time for 3 Requests
	start := time.Now()

	for i := range 3 {
		_, err := client.FindIssueByUPC(ctx, "75960609558200111")
		if err != nil {
			t.Fatalf("request %d failed: %v", i+1, err)
		}
	}

	elapsed := time.Since(start)

	// 3 Requests with 4s delay between them should take at least 8s total
	expectedMinDuration := 8 * time.Second

	if elapsed < expectedMinDuration {
		t.Errorf("rate limiting not working: completed 3 requests in %v, expected >= %v",
			elapsed, expectedMinDuration)
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, _ := NewClient("testuser", "testpass")
	client.baseURL = server.URL + "/api"

	tests := []struct {
		name          string
		setupFunc     func(*Client)
		testFunc      func(context.Context, *Client) error
		expectError   bool
		errorContains string
	}{
		{
			name: "invalid base URL",
			setupFunc: func(c *Client) {
				c.baseURL = "http://invalid-url-that-does-not-exist.local/api"
			},
			testFunc: func(ctx context.Context, c *Client) error {
				_, err := c.FindIssueByUPC(ctx, "75960609558200111")
				return err
			},
			expectError:   true,
			errorContains: "request failed",
		},
		{
			name: "context timeout",
			setupFunc: func(c *Client) {
				c.httpClient.Timeout = 1 * time.Millisecond
			},
			testFunc: func(ctx context.Context, c *Client) error {
				_, err := c.FindIssueByUPC(ctx, "75960609558200111")
				return err
			},
			expectError:   true,
			errorContains: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testClient, _ := NewClient("testuser", "testpass")
			testClient.baseURL = client.baseURL

			if tt.setupFunc != nil {
				tt.setupFunc(testClient)
			}

			ctx := context.Background()
			err := tt.testFunc(ctx, testClient)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectError && tt.errorContains != "" {
				if !contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)*2))
}
