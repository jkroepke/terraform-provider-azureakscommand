package provider

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// InvokeModel describes the resource data model.
type InvokeModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	ResourceGroupName  types.String `tfsdk:"resource_group_name"`
	Command            types.String `tfsdk:"command"`
	Context            types.String `tfsdk:"context"`
	Triggers           types.Map    `tfsdk:"triggers"`
	ExitCode           types.Int64  `tfsdk:"exit_code"`
	Output             types.String `tfsdk:"output"`
	ProvisioningState  types.String `tfsdk:"provisioning_state"`
	ProvisioningReason types.String `tfsdk:"provisioning_reason"`
	StartedAt          types.Int64  `tfsdk:"started_at"`
	FinishedAt         types.Int64  `tfsdk:"finished_at"`
}

func runCommand(ctx context.Context, client AzureAksCommandClient, resourceGroup string, resourceName string, command string, commandContext string) (*armcontainerservice.ManagedClustersClientRunCommandResponse, error) {
	payload := armcontainerservice.RunCommandRequest{
		Command: &command,
		Context: &commandContext,
	}

	res, err := client.client.Get(ctx, resourceGroup, resourceName, nil)
	if *res.Properties.AADProfile.Managed {
		token, err := client.cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{"6dae42f8-4368-4678-94ff-3960e28e3630"}})

		if err != nil {
			if !strings.Contains(err.Error(), "AADSTS1002012") {
				return nil, err
			}

			token, err = client.cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{"6dae42f8-4368-4678-94ff-3960e28e3630/.default"}})

			if err != nil {
				return nil, err
			}
		}

		payload.ClusterToken = &token.Token
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
