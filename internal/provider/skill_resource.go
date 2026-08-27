package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cepidalim-epam/terraform-provider-codemie/internal/client"
	"github.com/cepidalim-epam/terraform-provider-codemie/internal/provider/modelutil"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = (*skillResource)(nil)
var _ resource.ResourceWithImportState = (*skillResource)(nil)

type skillResource struct{ client *client.Client }

func NewSkillResource() resource.Resource { return &skillResource{} }

type skillModel struct {
	ID                      types.String         `tfsdk:"id"`
	Name                    types.String         `tfsdk:"name"`
	Description             types.String         `tfsdk:"description"`
	Content                 types.String         `tfsdk:"content"`
	Project                 types.String         `tfsdk:"project"`
	Visibility              types.String         `tfsdk:"visibility"`
	Categories              types.List           `tfsdk:"categories"`
	Toolkits                jsontypes.Normalized `tfsdk:"toolkits"`
	MCPServers              jsontypes.Normalized `tfsdk:"mcp_servers"`
	CompanionFiles          jsontypes.Normalized `tfsdk:"companion_files"`
	EnabledBuiltinSubagents types.List           `tfsdk:"enabled_builtin_subagents"`
}

func (r *skillResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_skill"
}
func (r *skillResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *skillResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages a CodeMie skill.", Attributes: map[string]schema.Attribute{
		"id":                        schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name":                      schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(3, 64), stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`), "must be kebab-case")}},
		"description":               schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(10, 1000)}},
		"content":                   schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthAtLeast(100)}},
		"project":                   schema.StringAttribute{Required: true},
		"visibility":                schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Validators: []validator.String{stringvalidator.OneOf("private", "project", "public")}},
		"categories":                schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}, Validators: []validator.List{listvalidator.SizeAtMost(3)}},
		"toolkits":                  schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"mcp_servers":               schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"companion_files":           schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"enabled_builtin_subagents": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
	}}
}
func (m *skillModel) values(ctx context.Context) ([]string, []string, error) {
	categories, d := modelutil.StringSlice(ctx, m.Categories)
	if d.HasError() {
		return nil, nil, fmt.Errorf("convert categories: %s", d.Errors()[0].Summary())
	}
	subs, d := modelutil.StringSlice(ctx, m.EnabledBuiltinSubagents)
	if d.HasError() {
		return nil, nil, fmt.Errorf("convert enabled_builtin_subagents: %s", d.Errors()[0].Summary())
	}
	return categories, subs, nil
}
func (m *skillModel) createRequest(ctx context.Context) (*client.SkillCreateRequest, error) {
	cats, subs, err := m.values(ctx)
	if err != nil {
		return nil, err
	}
	return &client.SkillCreateRequest{Name: m.Name.ValueString(), Description: m.Description.ValueString(), Content: m.Content.ValueString(), Project: m.Project.ValueString(), Visibility: stringPointerDefault(m.Visibility, "private"), Categories: cats, Toolkits: modelutil.RawJSON(m.Toolkits), MCPServers: modelutil.RawJSON(m.MCPServers), CompanionFiles: modelutil.RawJSON(m.CompanionFiles), EnabledBuiltinSubagents: subs}, nil
}
func (m *skillModel) updateRequest(ctx context.Context) (*client.SkillUpdateRequest, error) {
	cats, subs, err := m.values(ctx)
	if err != nil {
		return nil, err
	}
	return &client.SkillUpdateRequest{Name: modelutil.StringPointer(m.Name), Description: modelutil.StringPointer(m.Description), Content: modelutil.StringPointer(m.Content), Project: modelutil.StringPointer(m.Project), Visibility: stringPointerDefault(m.Visibility, "private"), Categories: cats, Toolkits: modelutil.RawJSON(m.Toolkits), MCPServers: modelutil.RawJSON(m.MCPServers), CompanionFiles: modelutil.RawJSON(m.CompanionFiles), EnabledBuiltinSubagents: subs}, nil
}
func stringPointerDefault(v types.String, def string) *string { s := valueString(v, def); return &s }
func (m *skillModel) hydrate(ctx context.Context, s *client.SkillDetailResponse) {
	m.ID = types.StringValue(s.ID)
	m.Name = types.StringValue(s.Name)
	m.Description = types.StringValue(s.Description)
	m.Content = types.StringValue(s.Content)
	m.Project = types.StringValue(s.Project)
	m.Visibility = types.StringValue(s.Visibility)
	m.Categories, _ = modelutil.StringList(ctx, []string(s.Categories))
	m.Toolkits = modelutil.NormalizedJSON(s.Toolkits, "[]")
	m.MCPServers = modelutil.NormalizedJSON(s.MCPServers, "[]")
	m.CompanionFiles = modelutil.NormalizedJSON(s.CompanionFiles, "[]")
	m.EnabledBuiltinSubagents, _ = modelutil.StringList(ctx, s.EnabledBuiltinSubagents)
}
func (r *skillResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m skillModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := m.createRequest(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Invalid skill configuration", err.Error())
		return
	}
	created, err := r.client.CreateSkill(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create skill", err.Error())
		return
	}
	s, err := r.client.GetSkill(ctx, created.ID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read skill after create", err.Error())
		return
	}
	m.hydrate(ctx, s)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
func (r *skillResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m skillModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	s, err := r.client.GetSkill(ctx, m.ID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read skill", err.Error())
		return
	}
	m.hydrate(ctx, s)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
func (r *skillResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m skillModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := m.updateRequest(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Invalid skill configuration", err.Error())
		return
	}
	if _, err = r.client.UpdateSkill(ctx, m.ID.ValueString(), body); err != nil {
		resp.Diagnostics.AddError("Unable to update skill", err.Error())
		return
	}
	s, err := r.client.GetSkill(ctx, m.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read skill after update", err.Error())
		return
	}
	m.hydrate(ctx, s)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
func (r *skillResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m skillModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSkill(ctx, m.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete skill", err.Error())
	}
}
func (r *skillResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
