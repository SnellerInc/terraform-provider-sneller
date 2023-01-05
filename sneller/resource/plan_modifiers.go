package resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func BoolDefaultValue(v bool) planmodifier.Bool {
	return &boolDefaultValuePlanModifier{types.BoolValue(v)}
}

type boolDefaultValuePlanModifier struct {
	DefaultValue types.Bool
}

var _ planmodifier.Bool = (*boolDefaultValuePlanModifier)(nil)

func (apm *boolDefaultValuePlanModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("defaults to '%v'", apm.DefaultValue.ValueBool())
}

func (apm *boolDefaultValuePlanModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("defaults to `%v`", apm.DefaultValue.ValueBool())
}

func (apm *boolDefaultValuePlanModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, res *planmodifier.BoolResponse) {
	// If the attribute configuration is not null, we are done here
	if !req.ConfigValue.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan modifier in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}

	res.PlanValue = apm.DefaultValue
}

func StringDefaultValue(v string) planmodifier.String {
	return &stringDefaultValuePlanModifier{types.StringValue(v)}
}

type stringDefaultValuePlanModifier struct {
	DefaultValue types.String
}

var _ planmodifier.String = (*stringDefaultValuePlanModifier)(nil)

func (apm *stringDefaultValuePlanModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("defaults to '%v'", apm.DefaultValue.ValueString())
}

func (apm *stringDefaultValuePlanModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("defaults to `%v`", apm.DefaultValue.ValueString())
}

func (apm *stringDefaultValuePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, res *planmodifier.StringResponse) {
	// If the attribute configuration is not null, we are done here
	if !req.ConfigValue.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan modifier in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}

	res.PlanValue = apm.DefaultValue
}

func Int64DefaultValue(v int64) planmodifier.Int64 {
	return &int64DefaultValuePlanModifier{types.Int64Value(v)}
}

type int64DefaultValuePlanModifier struct {
	DefaultValue types.Int64
}

var _ planmodifier.Int64 = (*int64DefaultValuePlanModifier)(nil)

func (apm *int64DefaultValuePlanModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("defaults to '%v'", apm.DefaultValue.ValueInt64())
}

func (apm *int64DefaultValuePlanModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("defaults to `%v`", apm.DefaultValue.ValueInt64())
}

func (apm *int64DefaultValuePlanModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, res *planmodifier.Int64Response) {
	// If the attribute configuration is not null, we are done here
	if !req.ConfigValue.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan modifier in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}

	res.PlanValue = apm.DefaultValue
}
