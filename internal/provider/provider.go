package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/jkroepke/terraform-provider-azure-aks-command/internal/helpers"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure AzureAksCommandProvider satisfies various provider interfaces.
var _ provider.Provider = &AzureAksCommandProvider{}
var _ provider.ProviderWithMetadata = &AzureAksCommandProvider{}

// AzureAksCommandProvider defines the provider implementation.
type AzureAksCommandProvider struct {
	// the version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// AzureAksCommandProviderModel describes the provider data model.
type AzureAksCommandProviderModel struct {
	SubscriptionId            types.String `tfsdk:"subscription_id"`
	ClientId                  types.String `tfsdk:"client_id"`
	TenantId                  types.String `tfsdk:"tenant_id"`
	Environment               types.String `tfsdk:"environment"`
	MetadataHost              types.String `tfsdk:"metadata_host"`
	ClientCertificatePath     types.String `tfsdk:"client_certificate_path"`
	ClientCertificatePassword types.String `tfsdk:"client_certificate_password"`
	ClientSecret              types.String `tfsdk:"client_secret"`
	OidcRequestToken          types.String `tfsdk:"oidc_request_token"`
	OidcRequestUrl            types.String `tfsdk:"oidc_request_url"`
	OidcToken                 types.String `tfsdk:"oidc_token"`
	OidcTokenFilePath         types.String `tfsdk:"oidc_token_file_path"`
	UseOidc                   types.Bool   `tfsdk:"use_oidc"`
	UseMsi                    types.Bool   `tfsdk:"use_msi"`
	MsiEndpoint               types.String `tfsdk:"msi_endpoint"`
	PartnerId                 types.String `tfsdk:"partner_id"`
}

func (p *AzureAksCommandProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "azure-aks-command"
	resp.Version = p.version
}

func (p *AzureAksCommandProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"subscription_id": {
				Description: "The Subscription ID which should be used.",
				Optional:    true,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_SUBSCRIPTION_ID"}, DefaultValue: ""},
				},
			},
			"client_id": {
				Description: "The Client ID which should be used.",
				Optional:    true,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_CLIENT_ID"}, DefaultValue: ""},
				},
			},
			"tenant_id": {
				Description: "The Tenant ID which should be used.",
				Optional:    true,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_TENANT_ID"}, DefaultValue: ""},
				},
			},
			"environment": {
				Description: "The Cloud Environment which should be used. Possible values are public, usgovernment, and china. Defaults to public.",
				Optional:    true,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_ENVIRONMENT"}, DefaultValue: "public"},
				},
			},
			"metadata_host": {
				Description: "The Hostname which should be used for the Azure Metadata Service.",
				Optional:    true,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_METADATA_HOSTNAME"}, DefaultValue: ""},
				},
			},

			// Client Certificate specific fields
			"client_certificate_path": {
				Description: "The path to the Client Certificate associated with the Service Principal for use when authenticating as a Service Principal using a Client Certificate.",
				Optional:    true,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_CLIENT_CERTIFICATE_PATH"}, DefaultValue: ""},
				},
			},
			"client_certificate_password": {
				Description: "The password associated with the Client Certificate. For use when authenticating as a Service Principal using a Client Certificate",
				Optional:    true,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_CLIENT_CERTIFICATE_PASSWORD"}, DefaultValue: ""},
				},
			},

			// Client Secret specific fields
			"client_secret": {
				Description: "The Client Secret which should be used. For use When authenticating as a Service Principal using a Client Secret.",
				Optional:    true,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_CLIENT_SECRET"}, DefaultValue: ""},
				},
			},

			// OIDC specifc fields
			"oidc_request_token": {
				Description: "The bearer token for the request to the OIDC provider. For use when authenticating as a Service Principal using OpenID Connect.",
				Optional:    true,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_OIDC_REQUEST_TOKEN", "ACTIONS_ID_TOKEN_REQUEST_TOKEN"}, DefaultValue: ""},
				},
			},
			"oidc_request_url": {
				Description: "The URL for the OIDC provider from which to request an ID token. For use when authenticating as a Service Principal using OpenID Connect.",
				Optional:    true,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_OIDC_REQUEST_TOKEN", "ACTIONS_ID_TOKEN_REQUEST_TOKEN"}, DefaultValue: ""},
				},
			},
			"oidc_token": {
				Description: "The OIDC ID token for use when authenticating as a Service Principal using OpenID Connect.",
				Optional:    true,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_OIDC_TOKEN"}, DefaultValue: ""},
				},
			},
			"oidc_token_file_path": {
				Description: "The path to a file containing an OIDC ID token for use when authenticating as a Service Principal using OpenID Connect.",
				Optional:    true,
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_OIDC_TOKEN_FILE_PATH"}, DefaultValue: ""},
				},
			},
			"use_oidc": {
				Description: "Allow OpenID Connect to be used for authentication",
				Optional:    true,
				Type:        types.BoolType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_USE_OIDC"}, DefaultValue: false},
				},
			},

			// Managed Service Identity specific fields
			"use_msi": {
				Description: "Allowed Managed Service Identity be used for Authentication.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_USE_MSI"}, DefaultValue: false},
				},
			},
			"msi_endpoint": {
				Description: "The path to a custom endpoint for Managed Service Identity - in most circumstances this should be detected automatically. ",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_MSI_ENDPOINT"}, DefaultValue: false},
				},
			},

			// Managed Tracking GUID for User-agent
			"partner_id": {
				Description: "A GUID/UUID that is registered with Microsoft to facilitate partner resource usage attribution.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					helpers.EnvVarModifier{EnvVarNames: []string{"ARM_PARTNER_ID"}, DefaultValue: ""},
				},
			},
		},
	}, nil
}

func (p *AzureAksCommandProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AzureAksCommandProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	builder := &authentication.Builder{
		SubscriptionID:      data.SubscriptionId.ValueString(),
		Environment:         data.Environment.ValueString(),
		MetadataHost:        data.MetadataHost.ValueString(),
		MsiEndpoint:         data.MsiEndpoint.ValueString(),
		ClientCertPassword:  data.ClientCertificatePassword.ValueString(),
		ClientCertPath:      data.ClientCertificatePath.ValueString(),
		IDTokenRequestToken: data.OidcRequestToken.ValueString(),
		IDTokenRequestURL:   data.OidcRequestUrl.ValueString(),
		IDToken:             data.OidcToken.ValueString(),
		IDTokenFilePath:     data.OidcTokenFilePath.ValueString(),

		// Feature Toggles
		SupportsClientCertAuth:         true,
		SupportsClientSecretAuth:       true,
		SupportsOIDCAuth:               data.UseOidc.ValueBool(),
		SupportsManagedServiceIdentity: data.UseMsi.ValueBool(),
		SupportsAzureCliToken:          true,
		SupportsAuxiliaryTenants:       false,

		// Doc Links
		ClientSecretDocsLink: "https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/service_principal_client_secret",

		// Use MSAL
		UseMicrosoftGraph: true,
	}

	config, err := builder.Build()

	if err != nil {
		resp.Diagnostics.AddError("building AzureRM Client", err.Error())
		return
	}

	/*
		_ = os.Setenv("AZURE_TENANT_ID", data.TenantId.ValueString())
		_ = os.Setenv("AZURE_CLIENT_ID", data.ClientId.ValueString())
		_ = os.Setenv("AZURE_CLIENT_SECRET", data.ClientSecret.ValueString())
		_ = os.Setenv("AZURE_CERTIFICATE_PATH", data.ClientCertificatePath.ValueString())
		_ = os.Setenv("AZURE_CERTIFICATE_PASSWORD", data.ClientCertificatePassword.ValueString())
		_ = os.Setenv("AZURE_ENVIRONMENT", data.Environment.ValueString())

		_ = os.Setenv("AZURE_FEDERATED_TOKEN_FILE", data.OidcTokenFilePath.ValueString())

		cred, err := helpers.NewAzureCredential()
		resp.Diagnostics.AddError("Error while authenticate against azure", err.Error())

		cred.GetToken(ctx, policy.TokenRequestOptions{})

	*/

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *AzureAksCommandProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *AzureAksCommandProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AzureAksCommandProvider{
			version: version,
		}
	}
}

func buildUserAgent(terraformVersion string, providerVersion string, partnerID string) string {
	if terraformVersion == "" {
		// Terraform 0.12 introduced this field to the protocol
		// We can therefore assume that if it's missing it's 0.10 or 0.11
		terraformVersion = "0.11+compatible"
	}

	tfUserAgent := fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io) Terraform Plugin SDK/%s", terraformVersion, meta.SDKVersionString())
	providerUserAgent := fmt.Sprintf("terraform-provider-azure-ak-scommand/%s", providerVersion)
	userAgent := strings.TrimSpace(fmt.Sprintf("%s %s", tfUserAgent, providerUserAgent))

	// append the CloudShell version to the user agent if it exists
	if azureAgent := os.Getenv("AZURE_HTTP_USER_AGENT"); azureAgent != "" {
		userAgent = fmt.Sprintf("%s %s", userAgent, azureAgent)
	}

	if partnerID != "" {
		userAgent = fmt.Sprintf("%s pid-%s", userAgent, partnerID)
	}
	return userAgent
}
