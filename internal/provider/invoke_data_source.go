package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &InvokeDataSource{}

func NewInvokeDataSource() datasource.DataSource {
	return &InvokeDataSource{}
}

// InvokeDataSource defines the data source implementation.
type InvokeDataSource struct {
	data AzureAksCommandClient
}

func (d *InvokeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invoke"
}

func (d *InvokeDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	description := "A data source to run a runCommand execution on a AKS. This data-source will execute the command before plan is computed. " +
		"It's recommended to use `azureakscommand_invoke` data source to perform readonly action, " +
		"please use `azureakscommand_invoke` resource, if user wants to perform actions which change a resource's state."

	return getSchema(description), nil
}

func (d *InvokeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(AzureAksCommandClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected AzureAksCommandClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.data = data
}

func (d *InvokeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *InvokeModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Prevent panic if the provider has not been configured.
	if d.data.client == nil || d.data.cred == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}
	runCommand, err := runCommand(ctx, d.data, data.ResourceGroupName.ValueString(), data.Name.ValueString(), data.Command.ValueString(), data.Context.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Error while executing runCommand", err.Error())
	}

	processRunCommand(runCommand, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
