package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
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

func (d *InvokeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A data source to run a runCommand execution on a AKS. This data-source will execute the command before plan is computed. " +
			"It's recommended to use `azureakscommand_invoke` data source to perform readonly action, " +
			"please use `azureakscommand_invoke` resource, if user wants to perform actions which change a resource's state.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the Managed Kubernetes Cluster to create. Changing this forces a new resource to be created.",
			},
			"resource_group_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Specifies the Resource Group where the Managed Kubernetes Cluster should exist. Changing this forces a new resource to be created.",
			},
			"command": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The command to run.",
			},
			"context": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A base64 encoded zip file containing the files required by the command.",
			},
			"triggers": schema.MapAttribute{
				Optional:            true,
				MarkdownDescription: "A map of arbitrary strings that, when changed, will force the null resource to be replaced, re-running any associated provisioners.",
				ElementType:         types.StringType,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The runCommand id",
			},
			"exit_code": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The exit code of the command",
			},
			"output": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The output of the command",
			},
			"provisioning_state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "provisioning state",
			},
			"provisioning_reason": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "An explanation of why provisioning_state is set to failed (if so).",
			},
			"started_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The time as unix timestamp when the command started.",
			},
			"finished_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The time as unix timestamp when the command finished.",
			},
		},
	}
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
	if d.data.managedClustersClient == nil || d.data.tokenCredential == nil {
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
