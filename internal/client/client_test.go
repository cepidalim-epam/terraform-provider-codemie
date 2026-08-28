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

func TestGetSkillCompanionFilesHydrated(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "token", "token_type": "Bearer", "expires_in": 3600})
	})
	mux.HandleFunc("/v1/skills/skill-with-files", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "skill-with-files",
			"name":        "test-skill",
			"description": "a valid description",
			"content":     "content",
			"project":     "demo",
			"visibility":  "private",
			"categories":  []string{},
			"companion_files": []map[string]any{
				{"path": "subdir/helper.py", "size": 42},
			},
		})
	})
	mux.HandleFunc("/v1/skills/skill-with-files/companion_files", func(w http.ResponseWriter, r *http.Request) {
		if path := r.URL.Query().Get("path"); path != "subdir/helper.py" {
			t.Errorf("unexpected path query param: %q", path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CompanionFileContentResponse{
			Path:     "subdir/helper.py",
			Content:  "print('hello companion')",
			Size:     42,
			MimeType: "text/x-python",
			Encoding: "utf-8",
		})
	})

	c, _ := New(Config{Host: server.URL, TokenURL: server.URL + "/token", ClientID: "id", ClientSecret: "secret"})
	skill, err := c.GetSkill(context.Background(), "skill-with-files")
	if err != nil {
		t.Fatalf("GetSkill failed: %v", err)
	}

	var files []CompanionFileContentResponse
	if err := json.Unmarshal(skill.CompanionFiles, &files); err != nil {
		t.Fatalf("unmarshal companion files failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 companion file, got %d", len(files))
	}
	if files[0].Content != "print('hello companion')" {
		t.Errorf("expected Content 'print(\\'hello companion\\')', got %q", files[0].Content)
	}
	if files[0].Path != "subdir/helper.py" {
		t.Errorf("expected Path 'subdir/helper.py', got %q", files[0].Path)
	}
}
