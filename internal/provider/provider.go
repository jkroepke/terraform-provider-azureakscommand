package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/jkroepke/terraform-provider-azure-aks-command/internal/clients"
	"github.com/jkroepke/terraform-provider-azure-aks-command/internal/helpers"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure AzureAksCommandProvider satisfies various provider interfaces.
var _ provider.Provider = &AzureAksCommandProvider{}

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
	DisableTerraformPartnerId types.Bool   `tfsdk:"disable_terraform_partner_id"`
}

type AzureAksCommandClient struct {
	cred   azcore.TokenCredential
	client *armcontainerservice.ManagedClustersClient
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AzureAksCommandProvider{
			version: version,
		}
	}
}
func (p *AzureAksCommandProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "azureakscommand"
	resp.Version = p.version
}

func (p *AzureAksCommandProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A terraform provider to run commands inside AKS through Azure API. It doesn't require any connection to the AKS." +
			"\n\n" +
			"See https://learn.microsoft.com/en-us/azure/aks/command-invoke for more information.",

		Attributes: map[string]schema.Attribute{
			"subscription_id": schema.StringAttribute{
				MarkdownDescription: "The Subscription ID which should be used. This can also be sourced from the `ARM_SUBSCRIPTION_ID` or `AZURE_SUBSCRIPTION_ID` Environment Variables.",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The Client ID which should be used. This can also be sourced from the `ARM_CLIENT_ID` or `AZURE_CLIENT_ID` Environment Variables.",
				Optional:            true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "The Tenant ID which should be used. This can also be sourced from the `ARM_TENANT_ID` or `AZURE_TENANT_ID` Environment Variables.",
				Optional:            true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "The Cloud Environment which should be used. Possible values are `public`, `usgovernment`, and `china`. Defaults to `public`. This can also be sourced from the `ARM_ENVIRONMENT` or `AZURE_ENVIRONMENT` Environment Variables.",
				Optional:            true,
			},

			// Client Certificate specific fields
			"client_certificate_path": schema.StringAttribute{
				MarkdownDescription: "The path to the Client Certificate associated with the Service Principal for use when authenticating as a Service Principal using a Client Certificate. This can also be sourced from the `ARM_CLIENT_CERTIFICATE_PATH` or `AZURE_CERTIFICATE_PATH` Environment Variables.",
				Optional:            true,
				Sensitive:           true,
			},
			"client_certificate_password": schema.StringAttribute{
				MarkdownDescription: "The password associated with the Client Certificate. For use when authenticating as a Service Principal using a Client Certificate. This can also be sourced from the `ARM_CLIENT_CERTIFICATE_PASSWORD` or `AZURE_CERTIFICATE_PASSWORD` Environment Variables.",
				Optional:            true,
				Sensitive:           true,
			},

			// Client Secret specific fields
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The Client Secret which should be used. For use When authenticating as a Service Principal using a Client Secret. This can also be sourced from the `ARM_CLIENT_SECRET` or `AZURE_CLIENT_SECRET` Environment Variables.",
				Optional:            true,
				Sensitive:           true,
			},

			// OIDC specific fields
			"oidc_request_token": schema.StringAttribute{
				MarkdownDescription: "The bearer token for the request to the OIDC provider. This can also be sourced from the `ARM_OIDC_REQUEST_TOKEN` or `ACTIONS_ID_TOKEN_REQUEST_TOKEN` Environment Variables.",
				Optional:            true,
			},
			"oidc_request_url": schema.StringAttribute{
				MarkdownDescription: "The URL for the OIDC provider from which to request an ID token. This can also be sourced from the `ARM_OIDC_REQUEST_URL` or `ACTIONS_ID_TOKEN_REQUEST_URL` Environment Variables.",
				Optional:            true,
			},
			"oidc_token": schema.StringAttribute{
				MarkdownDescription: "The ID token when authenticating using OpenID Connect (OIDC). This can also be sourced from the `ARM_OIDC_TOKEN` environment Variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"oidc_token_file_path": schema.StringAttribute{
				MarkdownDescription: "The path to a file containing an ID token when authenticating using OpenID Connect (OIDC). This can also be sourced from the `ARM_OIDC_TOKEN_FILE_PATH` or `AZURE_FEDERATED_TOKEN_FILE` environment Variable.",
				Optional:            true,
			},
			"use_oidc": schema.BoolAttribute{
				MarkdownDescription: "Should OIDC be used for Authentication? This can also be sourced from the `ARM_USE_OIDC` Environment Variable. Defaults to `false`.",
				Optional:            true,
			},

			// Managed Service Identity specific fields
			"use_msi": schema.BoolAttribute{
				MarkdownDescription: "Allowed Managed Service Identity be used for Authentication.",
				Optional:            true,
			},
			"msi_endpoint": schema.StringAttribute{
				MarkdownDescription: "The path to a custom endpoint for Managed Service Identity - in most circumstances, this should be detected automatically. This can also, be sourced from the `ARM_MSI_ENDPOINT` or `MSI_ENDPOINT` Environment Variable.",
				Optional:            true,
			},

			// Managed Tracking GUID for User-agent
			"partner_id": schema.StringAttribute{
				MarkdownDescription: "A GUID/UUID registered with Microsoft to facilitate partner resource usage attribution). This can also be sourced from the `ARM_PARTNER_ID` Environment Variable. Supported formats are `<guid>` / `pid-<guid>` (GUIDs registered in Partner Center) and `pid-<guid>-partnercenter` (for published [commercial marketplace Azure apps](https://docs.microsoft.com/azure/marketplace/azure-partner-customer-usage-attribution#commercial-marketplace-azure-apps)).",
				Optional:            true,
			},
			"disable_terraform_partner_id": schema.BoolAttribute{
				MarkdownDescription: "Disable sending the Terraform Partner ID if a custom partner_id isn't specified, which allows Microsoft to better understand the usage of Terraform. The Partner ID does not give the author any direct access to usage information. This can also be sourced from the `ARM_DISABLE_TERRAFORM_PARTNER_ID` environment variable. Defaults to `false`.",
				Optional:            true,
			},
		},
	}
}

func (p *AzureAksCommandProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AzureAksCommandProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	subscriptionId := getStringAttributeFromEnvironment(data.SubscriptionId, []string{"ARM_SUBSCRIPTION_ID", "AZURE_SUBSCRIPTION_ID"}, "")

	if subscriptionId == "" {
		resp.Diagnostics.AddError("Missing subscription_id", "Could not dectect subscription id through ARM_SUBSCRIPTION_ID, AZURE_SUBSCRIPTION_ID or provider attribute.")
		return
	}

	setAttributeFromTerraformOrEnvironment(data.TenantId, "AZURE_TENANT_ID", []string{"ARM_TENANT_ID", "AZURE_TENANT_ID"}, "")
	setAttributeFromTerraformOrEnvironment(data.ClientId, "AZURE_CLIENT_ID", []string{"ARM_CLIENT_ID"}, "")
	setAttributeFromTerraformOrEnvironment(data.ClientSecret, "AZURE_CLIENT_SECRET", []string{"ARM_CLIENT_SECRET"}, "")
	setAttributeFromTerraformOrEnvironment(data.ClientCertificatePath, "AZURE_CERTIFICATE_PATH", []string{"ARM_CLIENT_CERTIFICATE_PATH"}, "")
	setAttributeFromTerraformOrEnvironment(data.ClientCertificatePassword, "AZURE_CERTIFICATE_PASSWORD", []string{"ARM_CLIENT_CERTIFICATE_PASSWORD"}, "")
	setAttributeFromTerraformOrEnvironment(data.Environment, "AZURE_ENVIRONMENT", []string{"ARM_ENVIRONMENT"}, "public")

	if getBooleanAttributeFromEnvironment(data.UseMsi, []string{"ARM_USE_MSI"}, false) {
		setAttributeFromTerraformOrEnvironment(data.MsiEndpoint, "MSI_ENDPOINT", []string{"ARM_MSI_ENDPOINT"}, "")
	} else if getBooleanAttributeFromEnvironment(data.UseOidc, []string{"ARM_USE_OIDC"}, false) {
		setAttributeFromTerraformOrEnvironment(data.OidcTokenFilePath, "AZURE_FEDERATED_TOKEN_FILE", []string{"ARM_OIDC_TOKEN_FILE_PATH"}, "")

		token := getStringAttributeFromEnvironment(data.OidcToken, []string{"ARM_OIDC_TOKEN"}, "")

		oidcRequestUrl := getStringAttributeFromEnvironment(data.OidcRequestUrl, []string{"ARM_OIDC_REQUEST_URL", "ACTIONS_ID_TOKEN_REQUEST_URL"}, "")
		oidcRequestToken := getStringAttributeFromEnvironment(data.OidcRequestToken, []string{"ARM_OIDC_REQUEST_TOKEN", "ACTIONS_ID_TOKEN_REQUEST_TOKEN"}, "")

		if token != "" && oidcRequestUrl != "" && oidcRequestToken != "" {
			var err error

			token, err = helpers.GetOidcTokenFromGithubActions(data.OidcRequestUrl.ValueString(), data.OidcRequestToken.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error while request token from GH API", err.Error())
				return
			}
		}

		if token != "" {
			f, err := os.CreateTemp("", "token*")
			if err != nil {
				resp.Diagnostics.AddError("Error while setup OIDC token file", err.Error())
				return
			}

			_, err = f.WriteString(token)
			if err != nil {
				resp.Diagnostics.AddError("Error while setup OIDC token file", err.Error())
				return
			}

			_ = os.Setenv("AZURE_FEDERATED_TOKEN_FILE", f.Name())

			defer func(name string) {
				_ = os.Remove(name)
			}(f.Name())
		}
	}

	partnerId := getStringAttributeFromEnvironment(data.PartnerId, []string{"ARM_PARTNER_ID"}, "")
	disableTerraformPartnerId := getBooleanAttributeFromEnvironment(data.UseMsi, []string{"ARM_DISABLE_TERRAFORM_PARTNER_ID"}, false)

	userAgent := buildUserAgent(req.TerraformVersion, p.version, disableTerraformPartnerId, partnerId)

	cred, err := azidentity.NewDefaultAzureCredential(
		&azidentity.DefaultAzureCredentialOptions{
			ClientOptions: azcore.ClientOptions{
				Cloud: p.getCloudConfig(data),
				PerCallPolicies: []policy.Policy{
					clients.WithUserAgent(userAgent),
				},
			},
		},
	)

	if err != nil {
		resp.Diagnostics.AddError("Error while request token for AKS", err.Error())
		return
	}

	client, err := armcontainerservice.NewManagedClustersClient(subscriptionId, cred, &arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: p.getCloudConfig(data),
			PerCallPolicies: []policy.Policy{
				clients.WithUserAgent(userAgent),
			},
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("Error while request token for AKS", err.Error())
		return
	}

	resp.DataSourceData = AzureAksCommandClient{
		cred:   cred,
		client: client,
	}
	resp.ResourceData = AzureAksCommandClient{
		cred:   cred,
		client: client,
	}
}

func (p *AzureAksCommandProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInvokeResource,
	}
}

func (p *AzureAksCommandProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewInvokeDataSource,
	}
}

func (p *AzureAksCommandProvider) getCloudConfig(data AzureAksCommandProviderModel) cloud.Configuration {
	switch data.Environment.ValueString() {
	case "public":
		return cloud.AzurePublic
	case "usgovernment":
		return cloud.AzureGovernment
	case "china":
		return cloud.AzureChina
	default:
		return cloud.AzurePublic
	}
}

func buildUserAgent(terraformVersion string, providerVersion string, disableTerraformPartnerId bool, partnerID string) string {
	if terraformVersion == "" {
		// Terraform 0.12 introduced this field to the protocol
		// We can therefore assume that if it's missing its 0.10 or 0.11
		terraformVersion = "0.11+compatible"
	}

	tfUserAgent := fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io)", terraformVersion)
	providerUserAgent := fmt.Sprintf("terraform-provider-azure-aks-aad-token/%s", providerVersion)
	userAgent := strings.TrimSpace(fmt.Sprintf("%s %s", tfUserAgent, providerUserAgent))

	// append the CloudShell version to the user agent if it exists
	if azureAgent := os.Getenv("AZURE_HTTP_USER_AGENT"); azureAgent != "" {
		userAgent = fmt.Sprintf("%s %s", userAgent, azureAgent)
	}

	if disableTerraformPartnerId && partnerID != "" {
		userAgent = fmt.Sprintf("%s pid-%s", userAgent, partnerID)
	}
	return userAgent
}

func setAttributeFromTerraformOrEnvironment(value types.String, envVar string, envVarNames []string, defaultValue string) {
	configValue := getStringAttributeFromEnvironment(value, envVarNames, defaultValue)

	if configValue != "" {
		_ = os.Setenv(envVar, configValue)
	}
}

func getStringAttributeFromEnvironment(value types.String, envVarNames []string, defaultValue string) string {
	if !value.IsNull() {
		return value.ValueString()
	}

	for _, k := range envVarNames {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}

	return defaultValue
}
func getBooleanAttributeFromEnvironment(value types.Bool, envVarNames []string, defaultValue bool) bool {
	if !value.IsNull() {
		return value.ValueBool()
	}

	for _, k := range envVarNames {
		if v := os.Getenv(k); v != "" {
			if val, err := strconv.ParseBool(v); err == nil {
				return val
			}
		}
	}

	return defaultValue
}
