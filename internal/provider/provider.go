package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/cepidalim-epam/terraform-provider-codemie/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*CodemieProvider)(nil)

type CodemieProvider struct{ version string }

type providerModel struct {
	Host         types.String `tfsdk:"host"`
	TokenURL     types.String `tfsdk:"token_url"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &CodemieProvider{version: version} }
}

func (p *CodemieProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "codemie"
	resp.Version = p.version
}

func (p *CodemieProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	nonEmpty := []validator.String{stringvalidator.LengthAtLeast(1)}
	resp.Schema = schema.Schema{
		Description: "Manage CodeMie assistants, workflows, and skills.",
		Attributes: map[string]schema.Attribute{
			"host":          schema.StringAttribute{Optional: true, Description: "CodeMie API base URL. May be set with CODEMIE_HOST.", Validators: nonEmpty},
			"token_url":     schema.StringAttribute{Optional: true, Description: "OAuth2 token endpoint. May be set with CODEMIE_TOKEN_URL.", Validators: nonEmpty},
			"client_id":     schema.StringAttribute{Optional: true, Description: "OAuth2 client ID. May be set with CODEMIE_CLIENT_ID.", Validators: nonEmpty},
			"client_secret": schema.StringAttribute{Optional: true, Sensitive: true, Description: "OAuth2 client secret. May be set with CODEMIE_CLIENT_SECRET.", Validators: nonEmpty},
		},
	}
}

func configValue(value types.String, env string) string {
	if !value.IsNull() && !value.IsUnknown() {
		return value.ValueString()
	}
	return os.Getenv(env)
}

func (p *CodemieProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := map[string]string{
		"host": configValue(config.Host, "CODEMIE_HOST"), "token_url": configValue(config.TokenURL, "CODEMIE_TOKEN_URL"),
		"client_id": configValue(config.ClientID, "CODEMIE_CLIENT_ID"), "client_secret": configValue(config.ClientSecret, "CODEMIE_CLIENT_SECRET"),
	}
	var missing diag.Diagnostics
	for name, value := range values {
		if value == "" {
			missing.AddError("Missing CodeMie configuration", fmt.Sprintf("Set provider attribute %q or its CODEMIE_%s environment variable.", name, envSuffix(name)))
		}
	}
	resp.Diagnostics.Append(missing...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, err := client.New(client.Config{Host: values["host"], TokenURL: values["token_url"], ClientID: values["client_id"], ClientSecret: values["client_secret"]})
	if err != nil {
		resp.Diagnostics.AddError("Unable to configure CodeMie client", err.Error())
		return
	}
	resp.ResourceData = apiClient
}

func envSuffix(name string) string {
	if name == "token_url" {
		return "TOKEN_URL"
	}
	if name == "client_id" {
		return "CLIENT_ID"
	}
	if name == "client_secret" {
		return "CLIENT_SECRET"
	}
	return "HOST"
}

func (p *CodemieProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewAssistantResource, NewWorkflowResource, NewSkillResource}
}

func (p *CodemieProvider) DataSources(context.Context) []func() datasource.DataSource { return nil }
