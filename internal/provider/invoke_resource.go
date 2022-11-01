package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &InvokeResource{}
var _ resource.ResourceWithImportState = &InvokeResource{}

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

func (r *InvokeResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	description := "A resource to managed a runCommand execution on a AKS"

	return getSchema(description), nil
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
	if r.data.client == nil || r.data.cred == nil {
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
	if r.data.client == nil || r.data.cred == nil {
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

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	resp.Diagnostics.AddError("Unsupported operation", "azureakscommand_invoke does not support update")

	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InvokeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *InvokeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *InvokeResource) ImportState(_ context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Unsupported operation", "azureakscommand_invoke does not support import")
}
