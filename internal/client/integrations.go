package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

const settingsPath = "/v1/settings"

type CredentialValue struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value,omitempty"`
}

func (v *CredentialValue) UnmarshalJSON(data []byte) error {
	var value struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	v.Key = value.Key
	v.Value = value.Value
	if len(v.Value) == 0 {
		v.Value = json.RawMessage("null")
	}
	return nil
}

func (v CredentialValue) MarshalJSON() ([]byte, error) {
	value := v.Value
	if len(value) == 0 {
		value = json.RawMessage("null")
	}
	return json.Marshal(struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	}{Key: v.Key, Value: value})
}

type IntegrationRequest struct {
	ProjectName      *string           `json:"project_name"`
	Alias            string            `json:"alias"`
	CredentialType   string            `json:"credential_type"`
	CredentialValues []CredentialValue `json:"credential_values"`
	IsGlobal         bool              `json:"is_global"`
}

type Integration struct {
	ID               *string           `json:"id"`
	ProjectName      string            `json:"project_name"`
	Alias            *string           `json:"alias"`
	CredentialType   string            `json:"credential_type"`
	CredentialValues []CredentialValue `json:"credential_values"`
	SettingType      *string           `json:"setting_type"`
	IsGlobal         *bool             `json:"is_global"`
	Default          *bool             `json:"default"`
}

type settingsPage struct {
	Data       []Integration `json:"data"`
	Pagination *struct {
		Page  int `json:"page"`
		Pages int `json:"pages"`
		Total int `json:"total"`
	} `json:"pagination"`
}

type integrationCreateResponse struct {
	ID string `json:"id"`
}

func integrationFilters(scope string) string {
	query := url.Values{}
	query.Set("scope", scope)
	return settingsPath + "?" + query.Encode()
}

func settingsPathForScope(scope string) (string, error) {
	switch scope {
	case "user", "project":
		return settingsPath + "/" + scope, nil
	default:
		return "", fmt.Errorf("unsupported integration scope %q", scope)
	}
}

func (c *Client) ListIntegrations(ctx context.Context, scope string) ([]Integration, error) {
	if scope != "user" && scope != "project" {
		return nil, fmt.Errorf("unsupported integration scope %q", scope)
	}
	var integrations []Integration
	for page := 0; ; page++ {
		endpoint := fmt.Sprintf("%s&page=%d&per_page=100", integrationFilters(scope), page)
		var result settingsPage
		if err := c.doJSON(ctx, "GET", endpoint, nil, &result); err != nil {
			return nil, err
		}
		if result.Pagination == nil || result.Pagination.Page != page || result.Pagination.Pages < 0 || result.Pagination.Total < len(result.Data) {
			return nil, fmt.Errorf("client: integration settings response has incomplete pagination metadata")
		}
		if result.Pagination.Total > 0 && result.Pagination.Pages == 0 {
			return nil, fmt.Errorf("client: integration settings response has inconsistent pagination metadata")
		}
		integrations = append(integrations, result.Data...)
		if result.Pagination.Pages == 0 {
			if result.Pagination.Total != 0 || len(result.Data) != 0 {
				return nil, fmt.Errorf("client: integration settings response has inconsistent pagination metadata")
			}
			return integrations, nil
		}
		if page+1 >= result.Pagination.Pages {
			if len(integrations) < result.Pagination.Total {
				return nil, fmt.Errorf("client: integration settings response is missing results")
			}
			return integrations, nil
		}
	}
}

func findIntegrationByID(integrations []Integration, id string) (*Integration, error) {
	var found *Integration
	for i := range integrations {
		if integrations[i].ID == nil || *integrations[i].ID != id {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("integration ID %q appears more than once in the settings response", id)
		}
		found = &integrations[i]
	}
	return found, nil
}

func (c *Client) GetIntegration(ctx context.Context, scope, id string) (*Integration, error) {
	integrations, err := c.ListIntegrations(ctx, scope)
	if err != nil {
		return nil, err
	}
	integration, err := findIntegrationByID(integrations, id)
	if err != nil {
		return nil, err
	}
	if integration == nil {
		return nil, &APIError{StatusCode: 404, Body: "integration not found"}
	}
	if integration.SettingType != nil && *integration.SettingType != scope {
		return nil, &APIError{StatusCode: 404, Body: "integration not found in requested scope"}
	}
	if integration.SettingType == nil {
		integration.SettingType = &scope
	}
	return integration, nil
}

func (c *Client) CreateIntegration(ctx context.Context, scope string, request *IntegrationRequest) (*Integration, error) {
	endpoint, err := settingsPathForScope(scope)
	if err != nil {
		return nil, err
	}
	before, err := c.ListIntegrations(ctx, scope)
	if err != nil {
		return nil, err
	}
	requestedProjectName := ""
	if request.ProjectName != nil {
		requestedProjectName = *request.ProjectName
	}
	for _, current := range before {
		sameScopeIdentity := current.ProjectName == requestedProjectName
		if scope == "user" {
			sameScopeIdentity = true
		}
		if current.Alias != nil && *current.Alias == request.Alias && sameScopeIdentity && current.CredentialType == request.CredentialType && (current.SettingType == nil || *current.SettingType == scope) {
			return nil, fmt.Errorf("an integration with alias %q already exists in this scope and project; import it instead", request.Alias)
		}
	}
	var created integrationCreateResponse
	if err := c.doJSON(ctx, "POST", endpoint, request, &created); err != nil {
		return nil, err
	}
	if created.ID != "" {
		return c.GetIntegration(ctx, scope, created.ID)
	}
	after, err := c.ListIntegrations(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("integration was created, but its ID could not be read safely: %w", err)
	}
	previousIDs := make(map[string]struct{}, len(before))
	for _, current := range before {
		if current.ID != nil {
			previousIDs[*current.ID] = struct{}{}
		}
	}
	var matches []Integration
	for _, current := range after {
		if current.ID == nil || current.Alias == nil || *current.Alias != request.Alias || current.ProjectName != requestedProjectName || current.CredentialType != request.CredentialType || current.SettingType == nil || *current.SettingType != scope {
			continue
		}
		if _, existed := previousIDs[*current.ID]; !existed {
			matches = append(matches, current)
		}
	}
	if len(matches) != 1 {
		return nil, fmt.Errorf("integration was created, but its ID could not be determined unambiguously (matching new records: %d); inspect CodeMie settings for alias %q before retrying", len(matches), request.Alias)
	}
	return &matches[0], nil
}

func (c *Client) CreateProjectIntegration(ctx context.Context, request *IntegrationRequest) (*Integration, error) {
	return c.CreateIntegration(ctx, "project", request)
}

func (c *Client) CreateUserIntegration(ctx context.Context, request *IntegrationRequest) (*Integration, error) {
	return c.CreateIntegration(ctx, "user", request)
}

func (c *Client) UpdateIntegration(ctx context.Context, scope, id string, request *IntegrationRequest) error {
	base, err := settingsPathForScope(scope)
	if err != nil {
		return err
	}
	return c.doJSON(ctx, "PUT", base+"/"+url.PathEscape(id), request, nil)
}

func (c *Client) DeleteIntegration(ctx context.Context, scope, id string) error {
	base, err := settingsPathForScope(scope)
	if err != nil {
		return err
	}
	err = c.doJSON(ctx, "DELETE", base+"/"+url.PathEscape(id), nil, nil)
	if err != nil && IsNotFound(err) {
		return nil
	}
	return err
}
