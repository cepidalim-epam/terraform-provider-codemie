package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// AssistantRequest mirrors the OpenAPI AssistantRequest schema. It is used
// as the request body for both create (POST /v1/assistants) and update
// (PUT /v1/assistants/{id}) operations.
//
// Fields whose nested shape is not fully documented in the OpenAPI spec
// (Context, ToolKitDetails, MCPServerDetails, HedgingConfig,
// InteractiveFeaturesConfig, BedrockAgentData,
// BedrockAgentcoreRuntimeData, AgentCard, PromptVariable,
// GuardrailAssignmentItem, custom_metadata) are represented as raw
// json.RawMessage so that any shape accepted by the API can be configured.
type AssistantRequest struct {
	Name                      string          `json:"name"`
	Description               *string         `json:"description,omitempty"`
	SystemPrompt              *string         `json:"system_prompt,omitempty"`
	Project                   string          `json:"project"`
	Context                   json.RawMessage `json:"context,omitempty"`
	IconURL                   *string         `json:"icon_url,omitempty"`
	LLMModelType              *string         `json:"llm_model_type,omitempty"`
	EnableImageGeneration     *bool           `json:"enable_image_generation,omitempty"`
	ImageGenerationModel      *string         `json:"image_generation_model,omitempty"`
	Toolkits                  json.RawMessage `json:"toolkits,omitempty"`
	ConversationStarters      []string        `json:"conversation_starters,omitempty"`
	Shared                    *bool           `json:"shared,omitempty"`
	IsReact                   *bool           `json:"is_react,omitempty"`
	IsGlobal                  *bool           `json:"is_global,omitempty"`
	AgentMode                 *string         `json:"agent_mode,omitempty"`
	PlanPrompt                *string         `json:"plan_prompt,omitempty"`
	Slug                      *string         `json:"slug,omitempty"`
	Temperature               *float64        `json:"temperature,omitempty"`
	TopP                      *float64        `json:"top_p,omitempty"`
	ToolsTokensSizeLimit      *int64          `json:"tools_tokens_size_limit,omitempty"`
	SmartToolSelectionEnabled *bool           `json:"smart_tool_selection_enabled,omitempty"`
	HedgingConfig             json.RawMessage `json:"hedging_config,omitempty"`
	InteractiveFeatures       json.RawMessage `json:"interactive_features,omitempty"`
	MCPServers                json.RawMessage `json:"mcp_servers,omitempty"`
	AssistantIDs              []string        `json:"assistant_ids,omitempty"`
	EnabledBuiltinSubagents   []string        `json:"enabled_builtin_subagents,omitempty"`
	SkillIDs                  []string        `json:"skill_ids,omitempty"`
	Bedrock                   json.RawMessage `json:"bedrock,omitempty"`
	BedrockAgentcoreRuntime   json.RawMessage `json:"bedrock_agentcore_runtime,omitempty"`
	Type                      *string         `json:"type,omitempty"`
	AgentCard                 json.RawMessage `json:"agent_card,omitempty"`
	Categories                []string        `json:"categories,omitempty"`
	PromptVariables           json.RawMessage `json:"prompt_variables,omitempty"`
	CustomMetadata            json.RawMessage `json:"custom_metadata,omitempty"`
	SourceAssistantID         *string         `json:"source_assistant_id,omitempty"`
	GuardrailAssignments      json.RawMessage `json:"guardrail_assignments,omitempty"`
	SkipIntegrationValidation *bool           `json:"skip_integration_validation,omitempty"`
}

// Assistant mirrors the OpenAPI Assistant schema returned by
// GET /v1/assistants/id/{id}.
type Assistant struct {
	ID                        *string         `json:"id"`
	Name                      string          `json:"name"`
	Description               string          `json:"description"`
	SystemPrompt              string          `json:"system_prompt"`
	Project                   string          `json:"project"`
	Context                   json.RawMessage `json:"context,omitempty"`
	IconURL                   *string         `json:"icon_url"`
	LLMModelType              *string         `json:"llm_model_type"`
	EnableImageGeneration     *bool           `json:"enable_image_generation"`
	ImageGenerationModel      *string         `json:"image_generation_model"`
	Toolkits                  json.RawMessage `json:"toolkits,omitempty"`
	ConversationStarters      []string        `json:"conversation_starters"`
	Shared                    bool            `json:"shared"`
	IsReact                   bool            `json:"is_react"`
	IsGlobal                  *bool           `json:"is_global"`
	AgentMode                 *string         `json:"agent_mode"`
	PlanPrompt                *string         `json:"plan_prompt"`
	Slug                      *string         `json:"slug"`
	Temperature               *float64        `json:"temperature"`
	TopP                      *float64        `json:"top_p"`
	ToolsTokensSizeLimit      *int64          `json:"tools_tokens_size_limit"`
	SmartToolSelectionEnabled *bool           `json:"smart_tool_selection_enabled"`
	HedgingConfig             json.RawMessage `json:"hedging_config,omitempty"`
	InteractiveFeatures       json.RawMessage `json:"interactive_features,omitempty"`
	MCPServers                json.RawMessage `json:"mcp_servers,omitempty"`
	AssistantIDs              []string        `json:"assistant_ids"`
	EnabledBuiltinSubagents   []string        `json:"enabled_builtin_subagents"`
	SkillIDs                  []string        `json:"skill_ids"`
	Bedrock                   json.RawMessage `json:"bedrock,omitempty"`
	BedrockAgentcoreRuntime   json.RawMessage `json:"bedrock_agentcore_runtime,omitempty"`
	Type                      string          `json:"type"`
	AgentCard                 json.RawMessage `json:"agent_card,omitempty"`
	Categories                CategoryList    `json:"categories"`
	PromptVariables           json.RawMessage `json:"prompt_variables,omitempty"`
	CustomMetadata            json.RawMessage `json:"custom_metadata,omitempty"`
	GuardrailAssignments      json.RawMessage `json:"guardrail_assignments,omitempty"`
}

// AssistantCreateResponse mirrors the OpenAPI AssistantCreateResponse
// returned by POST /v1/assistants.
type AssistantCreateResponse struct {
	Message     string          `json:"message"`
	AssistantID *string         `json:"assistantId"`
	Validation  json.RawMessage `json:"validation,omitempty"`
}

// AssistantUpdateResponse mirrors the response of PUT /v1/assistants/{id}.
// Its exact schema was not fully documented in the OpenAPI spec excerpt
// available at implementation time; only the always-present message is
// modeled, and the provider always follows up with a GET to hydrate state.
type AssistantUpdateResponse struct {
	Message string `json:"message"`
}

// BaseResponse is a generic {message} response used by several delete
// endpoints.
type BaseResponse struct {
	Message string `json:"message"`
}

const assistantsPath = "/v1/assistants"

// CreateAssistant calls POST /v1/assistants.
func (c *Client) CreateAssistant(ctx context.Context, req *AssistantRequest) (*AssistantCreateResponse, error) {
	var out AssistantCreateResponse
	if err := c.doJSON(ctx, "POST", assistantsPath, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAssistant calls GET /v1/assistants/id/{id}.
func (c *Client) GetAssistant(ctx context.Context, id string) (*Assistant, error) {
	var out Assistant
	if err := c.doJSON(ctx, "GET", assistantsPath+"/id/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAssistant calls PUT /v1/assistants/{id}.
func (c *Client) UpdateAssistant(ctx context.Context, id string, req *AssistantRequest) error {
	return c.doJSON(ctx, "PUT", assistantsPath+"/"+url.PathEscape(id), req, nil)
}

// DeleteAssistant calls DELETE /v1/assistants/{id}.
func (c *Client) DeleteAssistant(ctx context.Context, id string) error {
	err := c.doJSON(ctx, "DELETE", assistantsPath+"/"+url.PathEscape(id), nil, nil)
	if err != nil && IsNotFound(err) {
		return nil
	}
	return err
}

// AttachSkillToAssistant calls POST /v1/assistants/{assistant_id}/skills.
func (c *Client) AttachSkillToAssistant(ctx context.Context, assistantID, skillID string) error {
	body := map[string]string{"skill_id": skillID}
	return c.doJSON(ctx, "POST", fmt.Sprintf("%s/%s/skills", assistantsPath, url.PathEscape(assistantID)), body, nil)
}
