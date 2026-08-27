package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientReusesOAuthToken(t *testing.T) {
	var tokens atomic.Int32
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		tokens.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600})
	})
	mux.HandleFunc("/v1/skills/skill-id", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "skill-id", "name": "test-skill", "description": "a valid description", "content": "content", "project": "demo", "visibility": "private", "categories": []string{}, "createdDate": "2026-01-01T00:00:00Z"})
	})
	c, err := New(Config{Host: server.URL, TokenURL: server.URL + "/token", ClientID: "id", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err = c.GetSkill(context.Background(), "skill-id"); err != nil {
			t.Fatal(err)
		}
	}
	if got := tokens.Load(); got != 1 {
		t.Fatalf("token requests = %d, want 1", got)
	}
}

func TestValidationError(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "token", "token_type": "Bearer", "expires_in": 3600})
	})
	mux.HandleFunc("/v1/skills", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"detail":[{"loc":["body","name"],"msg":"invalid name","type":"value_error"}]}`))
	})
	c, _ := New(Config{Host: server.URL, TokenURL: server.URL + "/token", ClientID: "id", ClientSecret: "secret"})
	_, err := c.CreateSkill(context.Background(), &SkillCreateRequest{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Validation == nil || len(apiErr.Validation.Detail) != 1 || apiErr.Validation.Detail[0].Message != "invalid name" {
		t.Fatalf("unexpected validation error: %#v", apiErr.Validation)
	}
}
