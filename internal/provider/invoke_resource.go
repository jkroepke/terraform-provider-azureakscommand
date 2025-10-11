package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &InvokeResource{}

func NewInvokeResource() resource.Resource {
	return &InvokeResource{}
}

// InvokeResource defines the resource implementation.
type InvokeResource struct {
	data AzureAksCommandClient
}

func (r *InvokeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invoke"
}

func (r *InvokeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	description := "A resource to managed a runCommand execution on a AKS" +
		"\n\n" +
		"The `triggers` argument allows specifying an arbitrary set of values that, when changed, will cause the resource to be replaced."

	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: description,
		Version:             1,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the Managed Kubernetes Cluster to create. Changing this forces a new resource to be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_group_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Specifies the Resource Group where the Managed Kubernetes Cluster should exist. Changing this forces a new resource to be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"command": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The command to run.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"context": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A base64 encoded zip file containing the files required by the command.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"triggers": schema.MapAttribute{
				Optional:            true,
				MarkdownDescription: "A map of arbitrary strings that, when changed, will force the null resource to be replaced, re-running any associated provisioners.",
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The runCommand id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"exit_code": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The exit code of the command",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"output": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The output of the command",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"provisioning_state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "provisioning state",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"provisioning_reason": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "An explanation of why provisioning_state is set to failed (if so).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"started_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The time as unix timestamp when the command started.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"finished_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The time as unix timestamp when the command finished.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *InvokeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.data = data
}

func (r *InvokeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *InvokeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Prevent panic if the provider has not been configured.
	if r.data.managedClustersClient == nil || r.data.tokenCredential == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	runCommand, err := runCommand(ctx, r.data, data.ResourceGroupName.ValueString(), data.Name.ValueString(), data.Command.ValueString(), data.Context.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Error while executing runCommand", err.Error())
	}

	if resp.Diagnostics.HasError() {
		return
	}

	processRunCommand(runCommand, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InvokeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *InvokeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Prevent panic if the provider has not been configured.
	if r.data.managedClustersClient == nil || r.data.tokenCredential == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InvokeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *InvokeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddError("Unsupported operation", "azureakscommand_invoke does not support update")
}

func (r *InvokeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *InvokeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}
