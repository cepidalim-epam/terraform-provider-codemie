package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
	mux.HandleFunc("/v1/skills/skill-with-files/companion-files/content", func(w http.ResponseWriter, r *http.Request) {
		if path := r.URL.Query().Get("path"); path != "subdir/helper.py" {
			t.Errorf("unexpected path query param: %q", path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CompanionFileContentResponse{
			Path:     "subdir/helper.py",
			Content:  "print('hello companion')",
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

func TestListIntegrationsReadsAllPages(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "token", "token_type": "Bearer", "expires_in": 3600})
	})
	mux.HandleFunc("/v1/settings", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("scope"); got != "project" {
			t.Errorf("scope = %q, want project", got)
		}
		w.Header().Set("Content-Type", "application/json")
		page := r.URL.Query().Get("page")
		if page == "0" {
			_, _ = w.Write([]byte(`{"data":[{"id":"one","project_name":"demo","alias":"first","credential_type":"Git","setting_type":"project","credential_values":[]}],"pagination":{"page":0,"per_page":1,"total":2,"pages":2}}`))
			return
		}
		if page != "1" {
			t.Errorf("page = %q, want 1", page)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"two","project_name":"demo","alias":"second","credential_type":"Jira","setting_type":"project","credential_values":[]}],"pagination":{"page":1,"per_page":1,"total":2,"pages":2}}`))
	})
	c, err := New(Config{Host: server.URL, TokenURL: server.URL + "/token", ClientID: "id", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	integrations, err := c.ListIntegrations(context.Background(), "project")
	if err != nil {
		t.Fatal(err)
	}
	if len(integrations) != 2 || *integrations[1].ID != "two" {
		t.Fatalf("unexpected paginated result: %#v", integrations)
	}
}

func TestCreateProjectIntegrationFindsUniqueNewID(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "token", "token_type": "Bearer", "expires_in": 3600})
	})
	var listCalls int
	mux.HandleFunc("/v1/settings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method %s", r.Method)
			return
		}
		listCalls++
		if listCalls > 1 {
			_, _ = w.Write([]byte(`{"data":[{"id":"new-id","project_name":"demo","alias":"terraform-test","credential_type":"Git","setting_type":"project","credential_values":[{"key":"token","value":"**********"}]}],"pagination":{"page":0,"per_page":100,"total":1,"pages":1}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[],"pagination":{"page":0,"per_page":100,"total":0,"pages":0}}`))
	})
	mux.HandleFunc("/v1/settings/project", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			var body IntegrationRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode create request: %v", err)
			}
			if body.Alias != "terraform-test" || body.CredentialType != "Git" {
				t.Errorf("unexpected request: %#v", body)
			}
			_, _ = w.Write([]byte(`{"message":"created"}`))
			return
		}
		t.Errorf("unexpected method %s", r.Method)
	})
	c, err := New(Config{Host: server.URL, TokenURL: server.URL + "/token", ClientID: "id", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	projectName := "demo"
	integration, err := c.CreateIntegration(context.Background(), "project", &IntegrationRequest{Alias: "terraform-test", ProjectName: &projectName, CredentialType: "Git", CredentialValues: []CredentialValue{{Key: "alias", Value: json.RawMessage(`"terraform-test"`)}, {Key: "token", Value: json.RawMessage(`"synthetic-secret"`)}}})
	if err != nil {
		t.Fatal(err)
	}
	if integration.ID == nil || *integration.ID != "new-id" {
		t.Fatalf("ID = %v, want new-id", integration.ID)
	}
}

func TestIntegrationAPIErrorDoesNotExposeCredentialResponse(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "token", "token_type": "Bearer", "expires_in": 3600})
	})
	mux.HandleFunc("/v1/settings/project", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":"secret-must-not-leak"}`))
	})
	c, err := New(Config{Host: server.URL, TokenURL: server.URL + "/token", ClientID: "id", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.ListIntegrations(context.Background(), "project")
	if err == nil {
		t.Fatal("expected API error")
	}
	if strings.Contains(err.Error(), "secret-must-not-leak") {
		t.Fatalf("integration API error exposed response body: %v", err)
	}
}

func TestIntegrationListRejectsMissingPaginationMetadata(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "token", "token_type": "Bearer", "expires_in": 3600})
	})
	mux.HandleFunc("/v1/settings", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	c, err := New(Config{Host: server.URL, TokenURL: server.URL + "/token", ClientID: "id", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.ListIntegrations(context.Background(), "project"); err == nil {
		t.Fatal("expected missing pagination metadata to fail closed")
	}
}

func TestCreateIntegrationRefusesExistingAlias(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "token", "token_type": "Bearer", "expires_in": 3600})
	})
	var postCalls int
	mux.HandleFunc("/v1/settings", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"existing","project_name":"demo","alias":"terraform-test","credential_type":"Git","setting_type":"project"}],"pagination":{"page":0,"per_page":100,"total":1,"pages":1}}`))
	})
	mux.HandleFunc("/v1/settings/project", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			postCalls++
			w.WriteHeader(http.StatusCreated)
		}
	})
	c, err := New(Config{Host: server.URL, TokenURL: server.URL + "/token", ClientID: "id", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	projectName := "demo"
	_, err = c.CreateIntegration(context.Background(), "project", &IntegrationRequest{Alias: "terraform-test", ProjectName: &projectName, CredentialType: "Git"})
	if err == nil || !strings.Contains(err.Error(), "import it instead") {
		t.Fatalf("expected existing-alias error, got %v", err)
	}
	if postCalls != 0 {
		t.Fatalf("create POSTs = %d, want 0", postCalls)
	}
}

func TestCreateIntegrationRefusesAmbiguousProjectID(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "token", "token_type": "Bearer", "expires_in": 3600})
	})
	var listCalls int
	mux.HandleFunc("/v1/settings", func(w http.ResponseWriter, _ *http.Request) {
		listCalls++
		rows := []map[string]any{}
		if listCalls > 1 {
			rows = append(rows,
				map[string]any{"id": "one", "project_name": "demo", "alias": "terraform-test", "credential_type": "Git", "setting_type": "project"},
				map[string]any{"id": "two", "project_name": "demo", "alias": "terraform-test", "credential_type": "Git", "setting_type": "project"},
			)
		}
		w.Header().Set("Content-Type", "application/json")
		pages := 0
		if len(rows) > 0 {
			pages = 1
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": rows, "pagination": map[string]int{"page": 0, "per_page": 100, "total": len(rows), "pages": pages}})
	})
	mux.HandleFunc("/v1/settings/project", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_, _ = w.Write([]byte(`{"message":"created"}`))
			return
		}
		http.NotFound(w, r)
	})
	c, err := New(Config{Host: server.URL, TokenURL: server.URL + "/token", ClientID: "id", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	projectName := "demo"
	_, err = c.CreateIntegration(context.Background(), "project", &IntegrationRequest{Alias: "terraform-test", ProjectName: &projectName, CredentialType: "Git"})
	if err == nil || !strings.Contains(err.Error(), "could not be determined unambiguously") {
		t.Fatalf("expected ambiguous ID error, got %v", err)
	}
}

func TestGetIntegrationDoesNotTreatMissingScopeMetadataAsNotFound(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "token", "token_type": "Bearer", "expires_in": 3600})
	})
	mux.HandleFunc("/v1/settings", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"one","project_name":"demo","alias":"test","credential_type":"Git"}],"pagination":{"page":0,"per_page":100,"total":1,"pages":1}}`))
	})
	c, err := New(Config{Host: server.URL, TokenURL: server.URL + "/token", ClientID: "id", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	integration, err := c.GetIntegration(context.Background(), "project", "one")
	if err != nil {
		t.Fatal(err)
	}
	if integration.SettingType == nil || *integration.SettingType != "project" {
		t.Fatalf("setting_type = %v, want inferred project", integration.SettingType)
	}
}
