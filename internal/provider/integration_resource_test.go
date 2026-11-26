package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cepidalim-epam/terraform-provider-codemie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIntegrationSchemaMarksCredentialsSensitive(t *testing.T) {
	var response resource.SchemaResponse
	(&integrationResource{}).Schema(context.Background(), resource.SchemaRequest{}, &response)

	credentials, ok := response.Schema.Attributes["credential_values"].(schema.StringAttribute)
	if !ok || !credentials.Sensitive {
		t.Fatal("credential_values must be marked sensitive")
	}
}

func TestIntegrationSchemaRequiredAndOptionalAttributes(t *testing.T) {
	var response resource.SchemaResponse
	(&integrationResource{}).Schema(context.Background(), resource.SchemaRequest{}, &response)
	if err := response.Schema.ValidateImplementation(context.Background()); err != nil {
		t.Fatalf("resource schema is invalid: %v", err)
	}
	for _, name := range []string{"project_name", "alias", "credential_type", "credential_values"} {
		if response.Schema.Attributes[name].IsRequired() {
			t.Fatalf("%s must support user-scoped integrations or imports", name)
		}
	}
}

func TestIntegrationSupportedTypeAllowlist(t *testing.T) {
	want := []string{"Webhook", "MCP", "Scheduler", "Jira", "OpenAPI", "Git"}
	if len(supportedIntegrationTypes) != len(want) {
		t.Fatalf("supported types = %v", supportedIntegrationTypes)
	}
	for i := range want {
		if supportedIntegrationTypes[i] != want[i] {
			t.Fatalf("supported types = %v, want %v", supportedIntegrationTypes, want)
		}
	}
}

func TestIntegrationCreatePreflightRejectsUnresolvedFields(t *testing.T) {
	for _, credentialType := range supportedIntegrationTypes {
		t.Run(credentialType, func(t *testing.T) {
			model := integrationModel{
				Scope: types.StringValue("project"), ProjectName: types.StringValue("example-project"),
				Alias: types.StringValue("test"), CredentialType: types.StringValue(credentialType),
				CredentialValues: jsontypes.NewNormalizedValue(`[{"key":"token","value":"synthetic"}]`),
			}
			if err := model.validateCreateConfiguration(); err != nil {
				t.Fatalf("supported type rejected: %v", err)
			}
		})
	}
	missingCredentials := integrationModel{
		Scope: types.StringValue("project"), ProjectName: types.StringValue("example-project"),
		Alias: types.StringValue("test"), CredentialType: types.StringValue("Git"),
		CredentialValues: jsontypes.NewNormalizedNull(),
	}
	if err := missingCredentials.validateCreateConfiguration(); err == nil {
		t.Fatal("create preflight accepted missing credential_values")
	}
}

func TestIntegrationCreatePreservesSecretsAcrossRefresh(t *testing.T) {
	var record map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"token","token_type":"Bearer","expires_in":3600}`))
			return
		}
		if r.URL.Path == "/v1/settings/project" && r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
				t.Errorf("decode request: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message":"created"}`))
			return
		}
		if r.URL.Path == "/v1/settings" && r.Method == http.MethodGet {
			query, _ := url.ParseQuery(r.URL.RawQuery)
			if query.Get("scope") != "project" {
				t.Errorf("scope query = %q", query.Get("scope"))
			}
			items := []map[string]any{}
			if record != nil {
				items = append(items, map[string]any{
					"id": "integration-id", "project_name": "example-project", "alias": record["alias"],
					"credential_type": record["credential_type"], "setting_type": "project", "is_global": false,
					"credential_values": []map[string]any{{"key": "token", "value": "**********"}},
				})
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"data": items, "pagination": map[string]int{"page": 0, "per_page": 100, "total": len(items), "pages": len(items)}})
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)

	apiClient, err := client.New(client.Config{Host: server.URL, TokenURL: server.URL + "/token", ClientID: "client", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	resourceInstance := &integrationResource{client: apiClient}
	var schemaResponse resource.SchemaResponse
	resourceInstance.Schema(context.Background(), resource.SchemaRequest{}, &schemaResponse)
	credentialValue := `[ {"key":"token","value":"terraform-secret"} ]`
	plan := tfsdk.Plan{Schema: schemaResponse.Schema}
	if diags := plan.Set(context.Background(), integrationModel{
		ID: types.StringUnknown(), Scope: types.StringValue("project"),
		ProjectName: types.StringValue("example-project"), Alias: types.StringValue("example-alias"),
		CredentialType: types.StringValue("Git"), CredentialValues: jsontypes.NewNormalizedValue(credentialValue),
		IsGlobal: types.BoolValue(false),
	}); diags.HasError() {
		t.Fatalf("build plan: %v", diags)
	}
	var stateResponse resource.CreateResponse
	stateResponse.State = tfsdk.State{Schema: &schemaResponse.Schema}
	resourceInstance.Create(context.Background(), resource.CreateRequest{Plan: plan}, &stateResponse)
	if stateResponse.Diagnostics.HasError() {
		t.Fatalf("create: %v", stateResponse.Diagnostics)
	}
	var state integrationModel
	stateResponse.Diagnostics.Append(stateResponse.State.Get(context.Background(), &state)...)
	if stateResponse.Diagnostics.HasError() {
		t.Fatalf("read created state: %v", stateResponse.Diagnostics)
	}
	if !strings.Contains(state.CredentialValues.ValueString(), "terraform-secret") {
		t.Fatalf("create lost configured secret: %s", state.CredentialValues.ValueString())
	}

	var readResponse resource.ReadResponse
	readResponse.State = tfsdk.State{Schema: &schemaResponse.Schema}
	resourceInstance.Read(context.Background(), resource.ReadRequest{State: stateResponse.State}, &readResponse)
	if readResponse.Diagnostics.HasError() {
		t.Fatalf("refresh: %v", readResponse.Diagnostics)
	}
	var refreshed integrationModel
	readResponse.Diagnostics.Append(readResponse.State.Get(context.Background(), &refreshed)...)
	if readResponse.Diagnostics.HasError() {
		t.Fatalf("read refreshed state: %v", readResponse.Diagnostics)
	}
	if !strings.Contains(refreshed.CredentialValues.ValueString(), "terraform-secret") {
		t.Fatalf("refresh lost configured secret: %s", refreshed.CredentialValues.ValueString())
	}

	values, ok := record["credential_values"].([]any)
	if !ok || len(values) != 2 {
		t.Fatalf("create payload did not contain credentials and alias: %#v", record)
	}
	if values[1].(map[string]any)["key"] != "alias" {
		t.Fatalf("create payload missing injected alias: %#v", values)
	}
}

func TestUserIntegrationCreateOmitsProjectName(t *testing.T) {
	var record map[string]json.RawMessage
	var created bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"token","token_type":"Bearer","expires_in":3600}`))
			return
		}
		if r.URL.Path == "/v1/settings/user" && r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
				t.Errorf("decode request: %v", err)
			}
			created = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"user-integration-id"}`))
			return
		}
		if r.URL.Path == "/v1/settings" && r.Method == http.MethodGet {
			if got := r.URL.Query().Get("scope"); got != "user" {
				t.Errorf("scope query = %q, want user", got)
			}
			items := []map[string]any{}
			if created {
				items = append(items, map[string]any{"id": "user-integration-id", "alias": "personal-git", "credential_type": "Git", "setting_type": "user", "project_name": "", "credential_values": []any{}})
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"data": items, "pagination": map[string]int{"page": 0, "per_page": 100, "total": len(items), "pages": len(items)}})
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)

	apiClient, err := client.New(client.Config{Host: server.URL, TokenURL: server.URL + "/token", ClientID: "client", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	model := integrationModel{
		ID: types.StringUnknown(), Scope: types.StringValue("user"), ProjectName: types.StringNull(),
		Alias: types.StringValue("personal-git"), CredentialType: types.StringValue("Git"),
		CredentialValues: jsontypes.NewNormalizedValue(`[{"key":"token","value":"synthetic"}]`),
	}
	request, err := model.request(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	integration, err := apiClient.CreateIntegration(context.Background(), "user", request)
	if err != nil {
		t.Fatal(err)
	}
	if integration.ID == nil || *integration.ID != "user-integration-id" {
		t.Fatalf("integration ID = %v", integration.ID)
	}
	if _, ok := record["project_name"]; !ok {
		t.Fatal("user integration request must include project_name as null")
	}
	if string(record["project_name"]) != "null" {
		t.Fatalf("user integration project_name = %s, want null", record["project_name"])
	}
}

func TestIntegrationRequestInjectsAliasAndPreservesJSONTypes(t *testing.T) {
	model := integrationModel{
		Scope:            types.StringValue("project"),
		ProjectName:      types.StringValue("example-project"),
		Alias:            types.StringValue("jira-main"),
		CredentialType:   types.StringValue("Jira"),
		CredentialValues: jsontypes.NewNormalizedValue(`[{"key":"is_cloud","value":true},{"key":"timeout","value":30},{"key":"options","value":[1,"two",false]},{"key":"optional","value":null}]`),
		IsGlobal:         types.BoolValue(false),
	}
	request, err := model.request(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if request.ProjectName == nil || *request.ProjectName != "example-project" {
		t.Fatalf("project name = %v, want example-project", request.ProjectName)
	}
	if len(request.CredentialValues) != 5 {
		t.Fatalf("credential values length = %d, want 5", len(request.CredentialValues))
	}
	var alias string
	if err := json.Unmarshal(request.CredentialValues[4].Value, &alias); err != nil {
		t.Fatal(err)
	}
	if request.CredentialValues[4].Key != "alias" || alias != "jira-main" {
		t.Fatalf("alias credential = %#v", request.CredentialValues[4])
	}
	var isCloud bool
	var timeout int
	if err := json.Unmarshal(request.CredentialValues[0].Value, &isCloud); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(request.CredentialValues[1].Value, &timeout); err != nil {
		t.Fatal(err)
	}
	if !isCloud || timeout != 30 {
		t.Fatalf("JSON credential value types were not preserved: %#v", request.CredentialValues)
	}
	if got := string(request.CredentialValues[2].Value); got != `[1,"two",false]` {
		t.Fatalf("array credential value = %s, want [1,\"two\",false]", got)
	}
	if got := string(request.CredentialValues[3].Value); got != "null" {
		t.Fatalf("explicit null credential value = %s, want null", got)
	}
}

func TestIntegrationRequestSynchronizesExistingAlias(t *testing.T) {
	model := integrationModel{
		Scope:            types.StringValue("project"),
		ProjectName:      types.StringValue("example-project"),
		Alias:            types.StringValue("jira-main"),
		CredentialType:   types.StringValue("Jira"),
		CredentialValues: jsontypes.NewNormalizedValue(`[{"key":"alias","value":"jira-main"},{"key":"token","value":"secret"}]`),
	}
	request, err := model.request(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(request.CredentialValues) != 2 {
		t.Fatalf("request did not preserve credential values: %#v", request.CredentialValues)
	}
	var alias string
	if err := json.Unmarshal(request.CredentialValues[0].Value, &alias); err != nil {
		t.Fatal(err)
	}
	if alias != "jira-main" {
		t.Fatalf("request did not synchronize the alias: %#v", request.CredentialValues)
	}
}

func TestIntegrationRequestRejectsDuplicateCredentialKeys(t *testing.T) {
	model := integrationModel{
		Scope:            types.StringValue("project"),
		ProjectName:      types.StringValue("example-project"),
		Alias:            types.StringValue("jira-main"),
		CredentialType:   types.StringValue("Jira"),
		CredentialValues: jsontypes.NewNormalizedValue(`[{"key":"token","value":"one"},{"key":"token","value":"two"}]`),
	}
	_, err := model.request(context.Background())
	if err == nil || !strings.Contains(err.Error(), "duplicate key") {
		t.Fatalf("expected duplicate key error, got %v", err)
	}
}

func TestIntegrationRequestRejectsUnsupportedTypesAndScopes(t *testing.T) {
	for _, test := range []struct {
		name     string
		scope    string
		typeName string
	}{
		{name: "unsupported type", scope: "project", typeName: "Confluence"},
		{name: "unsupported scope", scope: "global", typeName: "Git"},
		{name: "missing project", scope: "project", typeName: "Git"},
	} {
		t.Run(test.name, func(t *testing.T) {
			projectName := types.StringValue("example-project")
			if test.name == "missing project" {
				projectName = types.StringNull()
			}
			model := integrationModel{
				Scope:            types.StringValue(test.scope),
				ProjectName:      projectName,
				Alias:            types.StringValue("integration"),
				CredentialType:   types.StringValue(test.typeName),
				CredentialValues: jsontypes.NewNormalizedValue(`[{"key":"token","value":"secret"}]`),
			}
			if _, err := model.request(context.Background()); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestIntegrationRequestOmitsProjectNameForUserScope(t *testing.T) {
	model := integrationModel{
		Scope:            types.StringValue("user"),
		ProjectName:      types.StringNull(),
		Alias:            types.StringValue("jira-personal"),
		CredentialType:   types.StringValue("Jira"),
		CredentialValues: jsontypes.NewNormalizedValue(`[{"key":"token","value":"secret"}]`),
	}
	request, err := model.request(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if request.ProjectName != nil {
		t.Fatalf("user integration project_name = %q, want nil", *request.ProjectName)
	}
}

func TestIntegrationImportStateRequiresScopeAndID(t *testing.T) {
	for _, test := range []struct {
		name      string
		id        string
		wantError bool
	}{
		{name: "project", id: "project/abc"},
		{name: "user", id: "user/abc"},
		{name: "missing scope", id: "abc", wantError: true},
		{name: "unknown scope", id: "global/abc", wantError: true},
		{name: "empty ID", id: "project/", wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			scope, id, err := integrationImportParts(test.id)
			if (err != nil) != test.wantError {
				t.Fatalf("parse error = %v, want error %t", err, test.wantError)
			}
			if !test.wantError && (scope != strings.SplitN(test.id, "/", 2)[0] || id == "") {
				t.Fatalf("unexpected import parts: scope=%q id=%q", scope, id)
			}
		})
	}
}

func TestIntegrationReadPreservesConfiguredSecrets(t *testing.T) {
	previous := jsontypes.NewNormalizedValue(`[{"key":"token","value":"configured-secret"}]`)
	model := integrationModel{CredentialValues: previous}
	integration := &client.Integration{
		ID:             stringPointer("integration-id"),
		ProjectName:    "example-project",
		Alias:          stringPointer("example"),
		CredentialType: "Git",
		SettingType:    stringPointer("project"),
		CredentialValues: []client.CredentialValue{{
			Key:   "token",
			Value: json.RawMessage(`"**********"`),
		}},
	}
	model.hydrate(integration)
	model.CredentialValues = previous
	if model.CredentialValues.ValueString() != previous.ValueString() {
		t.Fatal("read replaced configured secret values")
	}
}

func TestIntegrationReadDoesNotAdoptAPISecretsWithoutState(t *testing.T) {
	model := integrationModel{CredentialValues: jsontypes.NewNormalizedNull()}
	model.hydrate(&client.Integration{
		ID:               stringPointer("integration-id"),
		ProjectName:      "example-project",
		Alias:            stringPointer("example"),
		CredentialType:   "Git",
		SettingType:      stringPointer("project"),
		CredentialValues: []client.CredentialValue{{Key: "token", Value: json.RawMessage(`"unmasked-api-secret"`)}},
	})
	if !model.CredentialValues.IsNull() {
		t.Fatalf("refresh adopted API credentials: %s", model.CredentialValues.ValueString())
	}
}

func TestIntegrationRequestRejectsEmptyValuesAndAliasMismatch(t *testing.T) {
	for _, raw := range []string{
		`[]`,
		`[{"key":"alias","value":"wrong"}]`,
		`[{"key":"","value":"x"}]`,
	} {
		model := integrationModel{
			Scope: types.StringValue("project"), ProjectName: types.StringValue("example-project"),
			Alias: types.StringValue("expected"), CredentialType: types.StringValue("Git"),
			CredentialValues: jsontypes.NewNormalizedValue(raw),
		}
		if _, err := model.request(context.Background()); err == nil {
			t.Fatalf("accepted invalid credential values: %s", raw)
		}
	}
}

func stringPointer(value string) *string { return &value }
