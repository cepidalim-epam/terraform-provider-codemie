package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cepidalim-epam/terraform-provider-codemie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = (*integrationResource)(nil)
var _ resource.ResourceWithImportState = (*integrationResource)(nil)

var supportedIntegrationTypes = []string{"Webhook", "MCP", "Scheduler", "Jira", "OpenAPI", "Git"}

type integrationResource struct{ client *client.Client }

type integrationModel struct {
	ID               types.String         `tfsdk:"id"`
	Scope            types.String         `tfsdk:"scope"`
	ProjectName      types.String         `tfsdk:"project_name"`
	Alias            types.String         `tfsdk:"alias"`
	CredentialType   types.String         `tfsdk:"credential_type"`
	CredentialValues jsontypes.Normalized `tfsdk:"credential_values"`
	IsGlobal         types.Bool           `tfsdk:"is_global"`
}

func NewIntegrationResource() resource.Resource { return &integrationResource{} }

func (r *integrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration"
}

func (r *integrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *integrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CodeMie user or project integration. Credential values are sensitive and remain in Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id":                schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"scope":             schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("project"), Validators: []validator.String{stringvalidator.OneOf("user", "project")}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"project_name":      schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"alias":             schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"credential_type":   schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{stringvalidator.OneOf(supportedIntegrationTypes...)}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"credential_values": schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}, Sensitive: true},
			"is_global":         schema.BoolAttribute{Optional: true, Computed: true},
		},
	}
}

func (m integrationModel) request(_ context.Context) (*client.IntegrationRequest, error) {
	if m.Alias.IsNull() || m.Alias.IsUnknown() || strings.TrimSpace(m.Alias.ValueString()) == "" {
		return nil, fmt.Errorf("alias must be set")
	}
	scope := m.Scope.ValueString()
	if scope != "user" && scope != "project" {
		return nil, fmt.Errorf("scope must be user or project")
	}
	if scope == "project" && (m.ProjectName.IsNull() || m.ProjectName.IsUnknown() || strings.TrimSpace(m.ProjectName.ValueString()) == "") {
		return nil, fmt.Errorf("project_name must be set")
	}
	if m.CredentialType.IsNull() || m.CredentialType.IsUnknown() || !contains(supportedIntegrationTypes, m.CredentialType.ValueString()) {
		return nil, fmt.Errorf("credential_type must be one of %s", strings.Join(supportedIntegrationTypes, ", "))
	}
	if m.CredentialValues.IsNull() || m.CredentialValues.IsUnknown() {
		return nil, fmt.Errorf("credential_values must be configured before creating or updating an integration")
	}
	var values []client.CredentialValue
	if err := json.Unmarshal([]byte(m.CredentialValues.ValueString()), &values); err != nil {
		return nil, fmt.Errorf("credential_values must be a JSON array of key/value objects")
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("credential_values must contain at least one credential")
	}
	aliasIndex := -1
	seen := make(map[string]struct{}, len(values))
	for i, value := range values {
		if strings.TrimSpace(value.Key) == "" {
			return nil, fmt.Errorf("credential_values entries must have a non-empty key")
		}
		if _, ok := seen[value.Key]; ok {
			return nil, fmt.Errorf("credential_values contains duplicate key %q", value.Key)
		}
		seen[value.Key] = struct{}{}
		if value.Key == "alias" {
			aliasIndex = i
			var configuredAlias string
			if err := json.Unmarshal(value.Value, &configuredAlias); err != nil || configuredAlias != m.Alias.ValueString() {
				return nil, fmt.Errorf("credential_values alias must match the alias attribute")
			}
		}
	}
	aliasValue, err := json.Marshal(m.Alias.ValueString())
	if err != nil {
		return nil, fmt.Errorf("encode alias credential: %w", err)
	}
	if aliasIndex == -1 {
		values = append(values, client.CredentialValue{Key: "alias", Value: aliasValue})
	} else {
		values[aliasIndex].Value = aliasValue
	}

	var projectName *string
	if scope == "project" {
		value := m.ProjectName.ValueString()
		projectName = &value
	}
	return &client.IntegrationRequest{
		ProjectName:      projectName,
		Alias:            m.Alias.ValueString(),
		CredentialType:   m.CredentialType.ValueString(),
		CredentialValues: values,
		IsGlobal:         !m.IsGlobal.IsNull() && !m.IsGlobal.IsUnknown() && m.IsGlobal.ValueBool(),
	}, nil
}

func (m integrationModel) validateCreateConfiguration() error {
	scope := m.Scope.ValueString()
	if scope != "user" && scope != "project" {
		return fmt.Errorf("scope must be user or project")
	}
	if scope == "project" && (m.ProjectName.IsNull() || m.ProjectName.IsUnknown() || strings.TrimSpace(m.ProjectName.ValueString()) == "") {
		return fmt.Errorf("project_name must be configured for project-scoped integrations")
	}
	if m.Alias.IsNull() || m.Alias.IsUnknown() || strings.TrimSpace(m.Alias.ValueString()) == "" {
		return fmt.Errorf("alias must be configured to create an integration")
	}
	if m.CredentialType.IsNull() || m.CredentialType.IsUnknown() || !contains(supportedIntegrationTypes, m.CredentialType.ValueString()) {
		return fmt.Errorf("credential_type must be configured as one of %s", strings.Join(supportedIntegrationTypes, ", "))
	}
	if m.CredentialValues.IsNull() || m.CredentialValues.IsUnknown() {
		return fmt.Errorf("credential_values must be configured to create an integration")
	}
	return nil
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func (m *integrationModel) hydrate(integration *client.Integration) {
	m.ID = stringFromPtr(integration.ID)
	m.ProjectName = types.StringValue(integration.ProjectName)
	m.Alias = stringFromPtr(integration.Alias)
	m.CredentialType = types.StringValue(integration.CredentialType)
	if integration.IsGlobal != nil {
		m.IsGlobal = types.BoolValue(*integration.IsGlobal)
	} else {
		m.IsGlobal = types.BoolValue(false)
	}
	if m.CredentialValues.IsNull() || m.CredentialValues.IsUnknown() {
		m.CredentialValues = jsontypes.NewNormalizedNull()
	}
}

func stringFromPtr(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func (r *integrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model integrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := model.validateCreateConfiguration(); err != nil {
		resp.Diagnostics.AddError("Invalid integration configuration", err.Error())
		return
	}
	request, err := model.request(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Invalid integration configuration", err.Error())
		return
	}
	integration, err := r.client.CreateIntegration(ctx, model.Scope.ValueString(), request)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create integration", "The CodeMie API rejected the request or the created integration could not be safely identified. Check the configured values and CodeMie settings before retrying.")
		return
	}
	model.hydrate(integration)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *integrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model integrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	integration, err := r.client.GetIntegration(ctx, model.Scope.ValueString(), model.ID.ValueString())
	if err != nil && client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read integration", "The CodeMie settings API request failed or returned an invalid response. Verify access and retry.")
		return
	}
	priorCredentials := model.CredentialValues
	model.hydrate(integration)
	model.CredentialValues = priorCredentials
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *integrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model integrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := model.validateCreateConfiguration(); err != nil {
		resp.Diagnostics.AddError("Invalid integration configuration", err.Error())
		return
	}
	request, err := model.request(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Invalid integration configuration", err.Error())
		return
	}
	if err := r.client.UpdateIntegration(ctx, model.Scope.ValueString(), model.ID.ValueString(), request); err != nil {
		resp.Diagnostics.AddError("Unable to update integration", "The CodeMie API rejected the request. Check the configured values and retry.")
		return
	}
	integration, err := r.client.GetIntegration(ctx, model.Scope.ValueString(), model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read integration after update", "The CodeMie settings API request failed or returned an invalid response.")
		return
	}
	model.hydrate(integration)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *integrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model integrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteIntegration(ctx, model.Scope.ValueString(), model.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete integration", "The CodeMie API rejected the request. Verify access and retry.")
	}
}

func (r *integrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	scope, id, err := integrationImportParts(req.ID)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Invalid integration import ID", "Use `user/<integration-id>` or `project/<integration-id>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scope"), scope)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func integrationImportParts(importID string) (string, string, error) {
	parts := strings.SplitN(importID, "/", 2)
	if len(parts) != 2 || (parts[0] != "user" && parts[0] != "project") || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("use user/<integration-id> or project/<integration-id>")
	}
	return parts[0], parts[1], nil
}

func (r *integrationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan integrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !plan.ID.IsUnknown() {
		return
	}
	if err := plan.validateCreateConfiguration(); err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("credential_values"), "Invalid integration configuration", err.Error())
	}
}
