package provider

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// InvokeModel describes the resource data model.
type InvokeModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	ResourceGroupName  types.String `tfsdk:"resource_group_name"`
	Command            types.String `tfsdk:"command"`
	Context            types.String `tfsdk:"context"`
	ExitCode           types.Int64  `tfsdk:"exit_code"`
	Output             types.String `tfsdk:"output"`
	ProvisioningState  types.String `tfsdk:"provisioning_state"`
	ProvisioningReason types.String `tfsdk:"provisioning_reason"`
	StartedAt          types.Int64  `tfsdk:"started_at"`
	FinishedAt         types.Int64  `tfsdk:"finished_at"`
}

func runCommand(ctx context.Context, client AzureAksCommandClient, resourceGroup string, resourceName string, command string, commandContext string) (*armcontainerservice.ManagedClustersClientRunCommandResponse, error) {
	token, err := client.cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{"6dae42f8-4368-4678-94ff-3960e28e3630"}})

	if err != nil {
		return nil, err
	}

	payload := armcontainerservice.RunCommandRequest{
		Command:      &command,
		ClusterToken: &token.Token,
		Context:      &commandContext,
	}

	poller, err := client.client.BeginRunCommand(ctx, resourceGroup, resourceName, payload, nil)
	if err != nil {
		return nil, err
	}

	runCommand, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &runCommand, nil
}

func getSchema(markdownDescription string) tfsdk.Schema {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: markdownDescription,
		Version:             1,
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Required:            true,
				MarkdownDescription: "The name of the Managed Kubernetes Cluster to create. Changing this forces a new resource to be created.",
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
			"resource_group_name": {
				Required:            true,
				MarkdownDescription: "Specifies the Resource Group where the Managed Kubernetes Cluster should exist. Changing this forces a new resource to be created.",
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
			"command": {
				Required:            true,
				MarkdownDescription: "The command to run.",
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
			"context": {
				Optional:            true,
				MarkdownDescription: "A base64 encoded zip file containing the files required by the command.",
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "The command id",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"exit_code": {
				Computed:            true,
				MarkdownDescription: "The exit code of the command",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.Int64Type,
			},
			"output": {
				Computed:            true,
				MarkdownDescription: "The output of the command",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"provisioning_state": {
				Computed:            true,
				MarkdownDescription: "provisioning State",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"provisioning_reason": {
				Computed:            true,
				MarkdownDescription: "An explanation of why provisioningState is set to failed (if so).",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"started_at": {
				Computed:            true,
				MarkdownDescription: "The time when the command started.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.Int64Type,
			},
			"finished_at": {
				Computed:            true,
				MarkdownDescription: "The time when the command finished.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.Int64Type,
			},
		},
	}
}

func processRunCommand(runCommand *armcontainerservice.ManagedClustersClientRunCommandResponse, data *InvokeModel) {
	if runCommand.ID != nil {
		data.Id = types.StringValue(*runCommand.ID)
	} else {
		data.Id = types.StringNull()
	}

	if runCommand.Properties.ExitCode != nil {
		data.ExitCode = types.Int64Value(int64(*runCommand.Properties.ExitCode))
	} else {
		data.ExitCode = types.Int64Null()
	}

	if runCommand.Properties.Logs != nil {
		data.Output = types.StringValue(*runCommand.Properties.Logs)
	} else {
		data.Output = types.StringNull()
	}

	if runCommand.Properties.ProvisioningState != nil {
		data.ProvisioningState = types.StringValue(*runCommand.Properties.ProvisioningState)
	} else {
		data.ProvisioningState = types.StringNull()
	}

	if runCommand.Properties.Reason != nil {
		data.ProvisioningReason = types.StringValue(*runCommand.Properties.Reason)
	} else {
		data.ProvisioningReason = types.StringNull()
	}

	if runCommand.Properties.StartedAt != nil {
		data.StartedAt = types.Int64Value(runCommand.Properties.StartedAt.Unix())
	} else {
		data.StartedAt = types.Int64Null()
	}

	if runCommand.Properties.FinishedAt != nil {
		data.FinishedAt = types.Int64Value(runCommand.Properties.FinishedAt.Unix())
	} else {
		data.FinishedAt = types.Int64Null()
	}
}
