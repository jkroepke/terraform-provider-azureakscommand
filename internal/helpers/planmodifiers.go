package helpers

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"strconv"
)

type EnvVarModifier struct {
	EnvVarNames  []string
	DefaultValue any
}

func (m EnvVarModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if !req.AttributePlan.IsNull() {
		return
	}

	var str types.String
	diags := tfsdk.ValueAs(ctx, req.AttributePlan, &str)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	input := multiEnvDefaultFunc(m)

	switch v := m.DefaultValue.(type) {
	case bool:
		if val, err := strconv.ParseBool(input.(string)); err == nil {
			resp.AttributePlan = types.BoolValue(val)
		}
	case string:
		resp.AttributePlan = types.StringValue(v)
	}
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m EnvVarModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %s", m.EnvVarNames)
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m EnvVarModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%s`", m.EnvVarNames)
}

func multiEnvDefaultFunc(envVarModifier EnvVarModifier) any {
	for _, k := range envVarModifier.EnvVarNames {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}

	return envVarModifier.DefaultValue
}
