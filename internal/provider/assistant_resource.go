package provider

import (
	"context"
	"fmt"

	"github.com/cepidalim-epam/terraform-provider-codemie/internal/client"
	"github.com/cepidalim-epam/terraform-provider-codemie/internal/provider/modelutil"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = (*assistantResource)(nil)
var _ resource.ResourceWithImportState = (*assistantResource)(nil)

type assistantResource struct{ client *client.Client }

func NewAssistantResource() resource.Resource { return &assistantResource{} }

type assistantModel struct {
	ID                        types.String         `tfsdk:"id"`
	Name                      types.String         `tfsdk:"name"`
	Description               types.String         `tfsdk:"description"`
	SystemPrompt              types.String         `tfsdk:"system_prompt"`
	Project                   types.String         `tfsdk:"project"`
	Context                   jsontypes.Normalized `tfsdk:"context"`
	IconURL                   types.String         `tfsdk:"icon_url"`
	LLMModelType              types.String         `tfsdk:"llm_model_type"`
	EnableImageGeneration     types.Bool           `tfsdk:"enable_image_generation"`
	ImageGenerationModel      types.String         `tfsdk:"image_generation_model"`
	Toolkits                  jsontypes.Normalized `tfsdk:"toolkits"`
	ConversationStarters      types.List           `tfsdk:"conversation_starters"`
	Shared                    types.Bool           `tfsdk:"shared"`
	IsReact                   types.Bool           `tfsdk:"is_react"`
	IsGlobal                  types.Bool           `tfsdk:"is_global"`
	AgentMode                 types.String         `tfsdk:"agent_mode"`
	PlanPrompt                types.String         `tfsdk:"plan_prompt"`
	Slug                      types.String         `tfsdk:"slug"`
	Temperature               types.Float64        `tfsdk:"temperature"`
	TopP                      types.Float64        `tfsdk:"top_p"`
	ToolsTokensSizeLimit      types.Int64          `tfsdk:"tools_tokens_size_limit"`
	SmartToolSelectionEnabled types.Bool           `tfsdk:"smart_tool_selection_enabled"`
	HedgingConfig             jsontypes.Normalized `tfsdk:"hedging_config"`
	InteractiveFeatures       jsontypes.Normalized `tfsdk:"interactive_features"`
	MCPServers                jsontypes.Normalized `tfsdk:"mcp_servers"`
	AssistantIDs              types.List           `tfsdk:"assistant_ids"`
	EnabledBuiltinSubagents   types.List           `tfsdk:"enabled_builtin_subagents"`
	SkillIDs                  types.List           `tfsdk:"skill_ids"`
	Bedrock                   jsontypes.Normalized `tfsdk:"bedrock"`
	BedrockAgentcoreRuntime   jsontypes.Normalized `tfsdk:"bedrock_agentcore_runtime"`
	Type                      types.String         `tfsdk:"type"`
	AgentCard                 jsontypes.Normalized `tfsdk:"agent_card"`
	Categories                types.List           `tfsdk:"categories"`
	PromptVariables           jsontypes.Normalized `tfsdk:"prompt_variables"`
	CustomMetadata            jsontypes.Normalized `tfsdk:"custom_metadata"`
	SourceAssistantID         types.String         `tfsdk:"source_assistant_id"`
	GuardrailAssignments      jsontypes.Normalized `tfsdk:"guardrail_assignments"`
	SkipIntegrationValidation types.Bool           `tfsdk:"skip_integration_validation"`
}

func (r *assistantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assistant"
}
func (r *assistantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configuration", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}
func jsonAttr(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:      true,
		Computed:      true,
		CustomType:    jsontypes.NormalizedType{},
		Description:   description,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}
func stringListAttr(description string, validators ...validator.List) schema.ListAttribute {
	return schema.ListAttribute{
		Optional:      true,
		Computed:      true,
		ElementType:   types.StringType,
		Description:   description,
		Validators:    validators,
		PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
	}
}
func (r *assistantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages a CodeMie assistant.", Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}, "name": schema.StringAttribute{Required: true}, "description": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}, "system_prompt": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}, "project": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("demo"), PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"context": jsonAttr("JSON array of assistant context objects."), "icon_url": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}, "llm_model_type": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}, "enable_image_generation": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}}, "image_generation_model": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}, "toolkits": jsonAttr("JSON array of toolkit configurations."),
		"conversation_starters": stringListAttr("Conversation starter strings."), "shared": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}}, "is_react": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}}, "is_global": schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}}, "agent_mode": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("general"), PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}, "plan_prompt": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}, "slug": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}, "temperature": schema.Float64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Float64{float64planmodifier.UseStateForUnknown()}}, "top_p": schema.Float64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Float64{float64planmodifier.UseStateForUnknown()}}, "tools_tokens_size_limit": schema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}}, "smart_tool_selection_enabled": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"hedging_config": jsonAttr("JSON hedging configuration."), "interactive_features": jsonAttr("JSON interactive features configuration."), "mcp_servers": jsonAttr("JSON array of MCP server configurations."), "assistant_ids": stringListAttr("Nested assistant IDs."), "enabled_builtin_subagents": stringListAttr("Enabled built-in subagents."), "skill_ids": stringListAttr("Attached skill IDs."), "bedrock": jsonAttr("JSON Bedrock agent configuration."), "bedrock_agentcore_runtime": jsonAttr("JSON Bedrock AgentCore runtime configuration."), "type": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("codemie"), PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}, "agent_card": jsonAttr("JSON agent card."), "categories": stringListAttr("Category IDs, at most three.", listvalidator.SizeAtMost(3)), "prompt_variables": jsonAttr("JSON prompt variable array."), "custom_metadata": jsonAttr("JSON custom metadata object."),
		"source_assistant_id": schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}}, "guardrail_assignments": jsonAttr("JSON guardrail assignment array."), "skip_integration_validation": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, Description: "Skip toolkit credential checks; request-only and retained from configuration."},
	}}
}

func (m *assistantModel) request(ctx context.Context) (*client.AssistantRequest, error) {
	conversation, d := modelutil.StringSlice(ctx, m.ConversationStarters)
	if d.HasError() {
		return nil, fmt.Errorf("convert conversation_starters: %s", d.Errors()[0].Summary())
	}
	assistantIDs, d := modelutil.StringSlice(ctx, m.AssistantIDs)
	if d.HasError() {
		return nil, fmt.Errorf("convert assistant_ids: %s", d.Errors()[0].Summary())
	}
	subagents, d := modelutil.StringSlice(ctx, m.EnabledBuiltinSubagents)
	if d.HasError() {
		return nil, fmt.Errorf("convert enabled_builtin_subagents: %s", d.Errors()[0].Summary())
	}
	skills, d := modelutil.StringSlice(ctx, m.SkillIDs)
	if d.HasError() {
		return nil, fmt.Errorf("convert skill_ids: %s", d.Errors()[0].Summary())
	}
	categories, d := modelutil.StringSlice(ctx, m.Categories)
	if d.HasError() {
		return nil, fmt.Errorf("convert categories: %s", d.Errors()[0].Summary())
	}
	return &client.AssistantRequest{Name: m.Name.ValueString(), Description: modelutil.StringPointer(m.Description), SystemPrompt: modelutil.StringPointer(m.SystemPrompt), Project: m.Project.ValueString(), Context: modelutil.RawJSON(m.Context), IconURL: modelutil.StringPointer(m.IconURL), LLMModelType: modelutil.StringPointer(m.LLMModelType), EnableImageGeneration: modelutil.BoolPointer(m.EnableImageGeneration), ImageGenerationModel: modelutil.StringPointer(m.ImageGenerationModel), Toolkits: modelutil.RawJSON(m.Toolkits), ConversationStarters: conversation, Shared: modelutil.BoolPointer(m.Shared), IsReact: modelutil.BoolPointer(m.IsReact), IsGlobal: modelutil.BoolPointer(m.IsGlobal), AgentMode: modelutil.StringPointer(m.AgentMode), PlanPrompt: modelutil.StringPointer(m.PlanPrompt), Slug: modelutil.StringPointer(m.Slug), Temperature: modelutil.Float64Pointer(m.Temperature), TopP: modelutil.Float64Pointer(m.TopP), ToolsTokensSizeLimit: modelutil.Int64Pointer(m.ToolsTokensSizeLimit), SmartToolSelectionEnabled: modelutil.BoolPointer(m.SmartToolSelectionEnabled), HedgingConfig: modelutil.RawJSON(m.HedgingConfig), InteractiveFeatures: modelutil.RawJSON(m.InteractiveFeatures), MCPServers: modelutil.RawJSON(m.MCPServers), AssistantIDs: assistantIDs, EnabledBuiltinSubagents: subagents, SkillIDs: skills, Bedrock: modelutil.RawJSON(m.Bedrock), BedrockAgentcoreRuntime: modelutil.RawJSON(m.BedrockAgentcoreRuntime), Type: modelutil.StringPointer(m.Type), AgentCard: modelutil.RawJSON(m.AgentCard), Categories: categories, PromptVariables: modelutil.RawJSON(m.PromptVariables), CustomMetadata: modelutil.RawJSON(m.CustomMetadata), SourceAssistantID: modelutil.StringPointer(m.SourceAssistantID), GuardrailAssignments: modelutil.RawJSON(m.GuardrailAssignments), SkipIntegrationValidation: modelutil.BoolPointer(m.SkipIntegrationValidation)}, nil
}

func (m *assistantModel) hydrate(ctx context.Context, a *client.Assistant) error {
	if a.ID != nil {
		m.ID = types.StringValue(*a.ID)
	}
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.SystemPrompt = types.StringValue(a.SystemPrompt)
	m.Project = types.StringValue(a.Project)
	m.Context = modelutil.NormalizedJSON(a.Context, "[]")
	m.IconURL = modelutil.StringFromPtr(a.IconURL)
	m.LLMModelType = modelutil.StringFromPtr(a.LLMModelType)
	m.EnableImageGeneration = modelutil.BoolFromPtr(a.EnableImageGeneration)
	m.ImageGenerationModel = modelutil.StringFromPtr(a.ImageGenerationModel)
	m.Toolkits = modelutil.NormalizedJSON(a.Toolkits, "[]")
	m.ConversationStarters, _ = modelutil.StringList(ctx, a.ConversationStarters)
	m.Shared = types.BoolValue(a.Shared)
	m.IsReact = types.BoolValue(a.IsReact)
	m.IsGlobal = modelutil.BoolFromPtr(a.IsGlobal)
	m.AgentMode = modelutil.StringFromPtr(a.AgentMode)
	m.PlanPrompt = modelutil.StringFromPtr(a.PlanPrompt)
	m.Slug = modelutil.StringFromPtr(a.Slug)
	m.Temperature = modelutil.Float64FromPtr(a.Temperature)
	m.TopP = modelutil.Float64FromPtr(a.TopP)
	m.ToolsTokensSizeLimit = modelutil.Int64FromPtr(a.ToolsTokensSizeLimit)
	m.SmartToolSelectionEnabled = modelutil.BoolFromPtr(a.SmartToolSelectionEnabled)
	m.HedgingConfig = modelutil.NormalizedJSON(a.HedgingConfig, "null")
	m.InteractiveFeatures = modelutil.NormalizedJSON(a.InteractiveFeatures, "null")
	m.MCPServers = modelutil.NormalizedJSON(a.MCPServers, "[]")
	m.AssistantIDs, _ = modelutil.StringList(ctx, a.AssistantIDs)
	m.EnabledBuiltinSubagents, _ = modelutil.StringList(ctx, a.EnabledBuiltinSubagents)
	m.SkillIDs, _ = modelutil.StringList(ctx, a.SkillIDs)
	m.Bedrock = modelutil.NormalizedJSON(a.Bedrock, "null")
	m.BedrockAgentcoreRuntime = modelutil.NormalizedJSON(a.BedrockAgentcoreRuntime, "null")
	m.Type = types.StringValue(a.Type)
	m.AgentCard = modelutil.NormalizedJSON(a.AgentCard, "null")
	m.Categories, _ = modelutil.StringList(ctx, []string(a.Categories))
	m.PromptVariables = modelutil.NormalizedJSON(a.PromptVariables, "[]")
	m.CustomMetadata = modelutil.NormalizedJSON(a.CustomMetadata, "null")
	m.GuardrailAssignments = modelutil.NormalizedJSON(a.GuardrailAssignments, "[]")
	return nil
}

func (r *assistantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m assistantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := m.request(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Invalid assistant configuration", err.Error())
		return
	}
	out, err := r.client.CreateAssistant(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create assistant", err.Error())
		return
	}
	if out.AssistantID == nil || *out.AssistantID == "" {
		resp.Diagnostics.AddError("CodeMie did not return an assistant ID", out.Message)
		return
	}
	m.ID = types.StringValue(*out.AssistantID)
	r.read(ctx, &m, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
	}
}
func (r *assistantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m assistantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	a, err := r.client.GetAssistant(ctx, m.ID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read assistant", err.Error())
		return
	}
	_ = m.hydrate(ctx, a)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
func (r *assistantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m assistantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := m.request(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Invalid assistant configuration", err.Error())
		return
	}
	if err = r.client.UpdateAssistant(ctx, m.ID.ValueString(), body); err != nil {
		resp.Diagnostics.AddError("Unable to update assistant", err.Error())
		return
	}
	r.read(ctx, &m, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
	}
}
func (r *assistantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m assistantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAssistant(ctx, m.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete assistant", err.Error())
	}
}
func (r *assistantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *assistantResource) read(ctx context.Context, m *assistantModel, diags interface{ AddError(string, string) }) {
	a, err := r.client.GetAssistant(ctx, m.ID.ValueString())
	if err != nil {
		diags.AddError("Unable to read assistant after write", err.Error())
		return
	}
	_ = m.hydrate(ctx, a)
}
