package client

import (
	"context"
	"encoding/json"
	"net/url"
)

type CreateWorkflowRequest struct {
	ID                   *string         `json:"id,omitempty"`
	Name                 string          `json:"name"`
	Mode                 *string         `json:"mode,omitempty"`
	Description          string          `json:"description"`
	StartHint            *string         `json:"start_hint,omitempty"`
	Project              string          `json:"project"`
	IconURL              *string         `json:"icon_url,omitempty"`
	YAMLConfig           *string         `json:"yaml_config,omitempty"`
	Shared               bool            `json:"shared"`
	Assistants           json.RawMessage `json:"assistants,omitempty"`
	Tools                json.RawMessage `json:"tools,omitempty"`
	States               json.RawMessage `json:"states,omitempty"`
	SupervisorPrompt     *string         `json:"supervisor_prompt,omitempty"`
	MetaConfig           *string         `json:"meta_config,omitempty"`
	GuardrailAssignments json.RawMessage `json:"guardrail_assignments,omitempty"`
}

type UpdateWorkflowRequest struct {
	ID                   *string         `json:"id,omitempty"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	StartHint            *string         `json:"start_hint,omitempty"`
	Project              string          `json:"project"`
	Mode                 string          `json:"mode"`
	IconURL              *string         `json:"icon_url,omitempty"`
	Shared               bool            `json:"shared"`
	YAMLConfig           *string         `json:"yaml_config,omitempty"`
	SupervisorPrompt     *string         `json:"supervisor_prompt,omitempty"`
	MetaConfig           *string         `json:"meta_config,omitempty"`
	GuardrailAssignments json.RawMessage `json:"guardrail_assignments,omitempty"`
}

type Workflow struct {
	ID                   *string         `json:"id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	StartHint            *string         `json:"start_hint"`
	Project              string          `json:"project"`
	Mode                 string          `json:"mode"`
	IconURL              *string         `json:"icon_url"`
	Shared               bool            `json:"shared"`
	YAMLConfig           *string         `json:"yaml_config"`
	Assistants           json.RawMessage `json:"assistants"`
	Tools                json.RawMessage `json:"tools"`
	States               json.RawMessage `json:"states"`
	SupervisorPrompt     *string         `json:"supervisor_prompt"`
	MetaConfig           *string         `json:"meta_config"`
	GuardrailAssignments json.RawMessage `json:"guardrail_assignments"`
}

const workflowsPath = "/v1/workflows"

func (c *Client) CreateWorkflow(ctx context.Context, req *CreateWorkflowRequest) error {
	return c.doJSON(ctx, "POST", workflowsPath, req, nil)
}
func (c *Client) GetWorkflow(ctx context.Context, id string) (*Workflow, error) {
	var out Workflow
	if err := c.doJSON(ctx, "GET", workflowsPath+"/id/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
func (c *Client) UpdateWorkflow(ctx context.Context, id string, req *UpdateWorkflowRequest) error {
	return c.doJSON(ctx, "PUT", workflowsPath+"/"+url.PathEscape(id), req, nil)
}
func (c *Client) DeleteWorkflow(ctx context.Context, id string) error {
	err := c.doJSON(ctx, "DELETE", workflowsPath+"/"+url.PathEscape(id), nil, nil)
	if err != nil && IsNotFound(err) {
		return nil
	}
	return err
}
