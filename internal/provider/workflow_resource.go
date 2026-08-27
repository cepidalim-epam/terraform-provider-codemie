package provider

import (
	"context"
	"fmt"

	"github.com/cepidalim-epam/terraform-provider-codemie/internal/client"
	"github.com/cepidalim-epam/terraform-provider-codemie/internal/provider/modelutil"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = (*workflowResource)(nil)
var _ resource.ResourceWithImportState = (*workflowResource)(nil)

type workflowResource struct{ client *client.Client }

func NewWorkflowResource() resource.Resource { return &workflowResource{} }

type workflowModel struct {
	ID                   types.String         `tfsdk:"id"`
	Name                 types.String         `tfsdk:"name"`
	Description          types.String         `tfsdk:"description"`
	StartHint            types.String         `tfsdk:"start_hint"`
	Project              types.String         `tfsdk:"project"`
	Mode                 types.String         `tfsdk:"mode"`
	IconURL              types.String         `tfsdk:"icon_url"`
	Shared               types.Bool           `tfsdk:"shared"`
	YAMLConfig           types.String         `tfsdk:"yaml_config"`
	Assistants           jsontypes.Normalized `tfsdk:"assistants"`
	Tools                jsontypes.Normalized `tfsdk:"tools"`
	States               jsontypes.Normalized `tfsdk:"states"`
	SupervisorPrompt     types.String         `tfsdk:"supervisor_prompt"`
	MetaConfig           types.String         `tfsdk:"meta_config"`
	GuardrailAssignments jsontypes.Normalized `tfsdk:"guardrail_assignments"`
}

func (r *workflowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}
func (r *workflowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *workflowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages a CodeMie workflow. yaml_config is the recommended graph representation.", Attributes: map[string]schema.Attribute{
		"id":                    schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name":                  schema.StringAttribute{Required: true},
		"description":           schema.StringAttribute{Required: true},
		"start_hint":            schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"project":               schema.StringAttribute{Required: true},
		"mode":                  schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"icon_url":              schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"shared":                schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"yaml_config":           schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"assistants":            schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}, Description: "JSON workflow assistant array.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"tools":                 schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}, Description: "JSON workflow tool array.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"states":                schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}, Description: "JSON workflow state array.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"supervisor_prompt":     schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"meta_config":           schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"guardrail_assignments": schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}, Description: "JSON guardrail assignment array.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func (m *workflowModel) createRequest(id string) *client.CreateWorkflowRequest {
	return &client.CreateWorkflowRequest{ID: &id, Name: m.Name.ValueString(), Mode: modelutil.StringPointer(m.Mode), Description: m.Description.ValueString(), StartHint: modelutil.StringPointer(m.StartHint), Project: m.Project.ValueString(), IconURL: modelutil.StringPointer(m.IconURL), YAMLConfig: modelutil.StringPointer(m.YAMLConfig), Shared: valueBool(m.Shared, true), Assistants: modelutil.RawJSON(m.Assistants), Tools: modelutil.RawJSON(m.Tools), States: modelutil.RawJSON(m.States), SupervisorPrompt: modelutil.StringPointer(m.SupervisorPrompt), MetaConfig: modelutil.StringPointer(m.MetaConfig), GuardrailAssignments: modelutil.RawJSON(m.GuardrailAssignments)}
}
func (m *workflowModel) updateRequest() *client.UpdateWorkflowRequest {
	id := m.ID.ValueString()
	return &client.UpdateWorkflowRequest{ID: &id, Name: m.Name.ValueString(), Description: m.Description.ValueString(), StartHint: modelutil.StringPointer(m.StartHint), Project: m.Project.ValueString(), Mode: valueString(m.Mode, "Sequential"), IconURL: modelutil.StringPointer(m.IconURL), Shared: valueBool(m.Shared, true), YAMLConfig: modelutil.StringPointer(m.YAMLConfig), SupervisorPrompt: modelutil.StringPointer(m.SupervisorPrompt), MetaConfig: modelutil.StringPointer(m.MetaConfig), GuardrailAssignments: modelutil.RawJSON(m.GuardrailAssignments)}
}
func valueString(v types.String, def string) string {
	if v.IsNull() || v.IsUnknown() {
		return def
	}
	return v.ValueString()
}
func valueBool(v types.Bool, def bool) bool {
	if v.IsNull() || v.IsUnknown() {
		return def
	}
	return v.ValueBool()
}
func (m *workflowModel) hydrate(w *client.Workflow) {
	if w.ID != nil {
		m.ID = types.StringValue(*w.ID)
	}
	m.Name = types.StringValue(w.Name)
	m.Description = types.StringValue(w.Description)
	m.StartHint = modelutil.StringFromPtr(w.StartHint)
	m.Project = types.StringValue(w.Project)
	m.Mode = types.StringValue(w.Mode)
	m.IconURL = modelutil.StringFromPtr(w.IconURL)
	m.Shared = types.BoolValue(w.Shared)
	m.YAMLConfig = modelutil.StringFromPtr(w.YAMLConfig)
	m.Assistants = modelutil.NormalizedJSON(w.Assistants, "[]")
	m.Tools = modelutil.NormalizedJSON(w.Tools, "[]")
	m.States = modelutil.NormalizedJSON(w.States, "[]")
	m.SupervisorPrompt = modelutil.StringFromPtr(w.SupervisorPrompt)
	m.MetaConfig = modelutil.StringFromPtr(w.MetaConfig)
	m.GuardrailAssignments = modelutil.NormalizedJSON(w.GuardrailAssignments, "[]")
}
func (r *workflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m workflowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := uuid.NewString()
	if err := r.client.CreateWorkflow(ctx, m.createRequest(id)); err != nil {
		resp.Diagnostics.AddError("Unable to create workflow", err.Error())
		return
	}
	m.ID = types.StringValue(id)
	w, err := r.client.GetWorkflow(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workflow after create", err.Error())
		return
	}
	m.hydrate(w)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
func (r *workflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m workflowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	w, err := r.client.GetWorkflow(ctx, m.ID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workflow", err.Error())
		return
	}
	m.hydrate(w)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
func (r *workflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m workflowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateWorkflow(ctx, m.ID.ValueString(), m.updateRequest()); err != nil {
		resp.Diagnostics.AddError("Unable to update workflow", err.Error())
		return
	}
	w, err := r.client.GetWorkflow(ctx, m.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workflow after update", err.Error())
		return
	}
	m.hydrate(w)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
func (r *workflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m workflowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteWorkflow(ctx, m.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete workflow", err.Error())
	}
}
func (r *workflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
