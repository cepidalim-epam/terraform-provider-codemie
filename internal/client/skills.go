package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// SkillCreateRequest mirrors the OpenAPI SkillCreateRequest schema used by
// POST /v1/skills.
type SkillCreateRequest struct {
	Name                    string          `json:"name"`
	Description             string          `json:"description"`
	Content                 string          `json:"content"`
	Project                 string          `json:"project"`
	Visibility              *string         `json:"visibility,omitempty"`
	Categories              []string        `json:"categories,omitempty"`
	Toolkits                json.RawMessage `json:"toolkits,omitempty"`
	MCPServers              json.RawMessage `json:"mcp_servers,omitempty"`
	CompanionFiles          json.RawMessage `json:"companion_files,omitempty"`
	EnabledBuiltinSubagents []string        `json:"enabled_builtin_subagents,omitempty"`
}

// SkillUpdateRequest mirrors the OpenAPI SkillUpdateRequest schema used by
// PUT /v1/skills/{id}. All fields are optional/nullable on the API side.
type SkillUpdateRequest struct {
	Name                    *string         `json:"name,omitempty"`
	Description             *string         `json:"description,omitempty"`
	Content                 *string         `json:"content,omitempty"`
	Project                 *string         `json:"project,omitempty"`
	Visibility              *string         `json:"visibility,omitempty"`
	Categories              []string        `json:"categories,omitempty"`
	Toolkits                json.RawMessage `json:"toolkits,omitempty"`
	MCPServers              json.RawMessage `json:"mcp_servers,omitempty"`
	CompanionFiles          json.RawMessage `json:"companion_files,omitempty"`
	EnabledBuiltinSubagents []string        `json:"enabled_builtin_subagents,omitempty"`
}

// SkillDetailResponse mirrors the OpenAPI SkillDetailResponse schema
// returned by create/get/update skill endpoints.
type SkillDetailResponse struct {
	ID                      string          `json:"id"`
	Name                    string          `json:"name"`
	Description             string          `json:"description"`
	Content                 string          `json:"content"`
	Project                 string          `json:"project"`
	DisplayName             *string         `json:"display_name"`
	Visibility              string          `json:"visibility"`
	Categories              CategoryList    `json:"categories"`
	Toolkits                json.RawMessage `json:"toolkits,omitempty"`
	MCPServers              json.RawMessage `json:"mcp_servers,omitempty"`
	CompanionFiles          json.RawMessage `json:"companion_files,omitempty"`
	EnabledBuiltinSubagents []string        `json:"enabled_builtin_subagents"`
}

type CompanionFileContentResponse struct {
	Content  string `json:"content"`
	Path     string `json:"path"`
	Size     int    `json:"size"`
	MimeType string `json:"mime_type"`
	Encoding string `json:"encoding"`
}

const skillsPath = "/v1/skills"

// CreateSkill calls POST /v1/skills.
func (c *Client) CreateSkill(ctx context.Context, req *SkillCreateRequest) (*SkillDetailResponse, error) {
	var out SkillDetailResponse
	if err := c.doJSON(ctx, "POST", skillsPath, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSkill calls GET /v1/skills/{id}.
func (c *Client) GetSkill(ctx context.Context, id string) (*SkillDetailResponse, error) {
	var out SkillDetailResponse
	if err := c.doJSON(ctx, "GET", skillsPath+"/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}

	if len(out.CompanionFiles) > 0 && string(out.CompanionFiles) != "null" {
		var files []CompanionFileContentResponse
		if err := json.Unmarshal(out.CompanionFiles, &files); err == nil && len(files) > 0 {
			detailedFiles := make([]CompanionFileContentResponse, len(files))
			for i, f := range files {
				if f.Path != "" {
					fileContent, err := c.GetSkillCompanionFileContent(ctx, id, f.Path)
					if err != nil {
						return nil, err
					}
					detailedFiles[i] = *fileContent
				} else {
					detailedFiles[i] = f
				}
			}
			raw, err := json.Marshal(detailedFiles)
			if err != nil {
				return nil, err
			}
			out.CompanionFiles = raw
		}
	}

	return &out, nil
}

// UpdateSkill calls PUT /v1/skills/{id}.
func (c *Client) UpdateSkill(ctx context.Context, id string, req *SkillUpdateRequest) (*SkillDetailResponse, error) {
	var out SkillDetailResponse
	if err := c.doJSON(ctx, "PUT", skillsPath+"/"+url.PathEscape(id), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSkill calls DELETE /v1/skills/{id}. The API returns 204 on success.
func (c *Client) DeleteSkill(ctx context.Context, id string) error {
	err := c.doJSON(ctx, "DELETE", skillsPath+"/"+url.PathEscape(id), nil, nil)
	if err != nil && IsNotFound(err) {
		return nil
	}
	return err
}

// GetSkillCompanionFileContent calls GET /v1/skills/{id}/companion_files?path={path}.
func (c *Client) GetSkillCompanionFileContent(ctx context.Context, id string, path string) (*CompanionFileContentResponse, error) {
	var out CompanionFileContentResponse
	endpoint := fmt.Sprintf("%s/%s/companion_files?path=%s", skillsPath, url.PathEscape(id), url.QueryEscape(path))
	if err := c.doJSON(ctx, "GET", endpoint, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
