package modelutil

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringPointer converts a types.String to a *string. If the value is null or
// unknown, it returns nil.
func StringPointer(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

// BoolPointer converts a types.Bool to a *bool. If the value is null or
// unknown, it returns nil.
func BoolPointer(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

// Float64Pointer converts a types.Float64 to a *float64. If the value is null
// or unknown, it returns nil.
func Float64Pointer(v types.Float64) *float64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	f := v.ValueFloat64()
	return &f
}

// Int64Pointer converts a types.Int64 to a *int64. If the value is null or
// unknown, it returns nil.
func Int64Pointer(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}

// StringFromPtr returns a types.String from a *string.
func StringFromPtr(p *string) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(*p)
}

// BoolFromPtr returns a types.Bool from a *bool.
func BoolFromPtr(p *bool) types.Bool {
	if p == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*p)
}

// Float64FromPtr returns a types.Float64 from a *float64.
func Float64FromPtr(p *float64) types.Float64 {
	if p == nil {
		return types.Float64Null()
	}
	return types.Float64Value(*p)
}

// Int64FromPtr returns a types.Int64 from a *int64.
func Int64FromPtr(p *int64) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*p)
}

// StringPtr returns a pointer to s, or nil if s is empty. Useful when
// building API request structs from Terraform config values that may be
// the empty string sentinel for "unset".
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StringValue dereferences p, returning "" if p is nil.
func StringValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// BoolValue dereferences p, returning def if p is nil.
func BoolValue(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

func StringSlice(ctx context.Context, value types.List) ([]string, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}
	var result []string
	diags := value.ElementsAs(ctx, &result, false)
	return result, diags
}

func StringList(ctx context.Context, value []string) (types.List, diag.Diagnostics) {
	if value == nil {
		value = []string{}
	}
	return types.ListValueFrom(ctx, types.StringType, value)
}

func RawJSON(value jsontypes.Normalized) json.RawMessage {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return json.RawMessage(value.ValueString())
}

func NormalizedJSON(value json.RawMessage, empty string) jsontypes.Normalized {
	if len(value) == 0 || string(value) == "null" {
		return jsontypes.NewNormalizedValue(empty)
	}
	var decoded any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return jsontypes.NewNormalizedValue(string(value))
	}
	normalized, err := json.Marshal(decoded)
	if err != nil {
		return jsontypes.NewNormalizedValue(string(value))
	}
	return jsontypes.NewNormalizedValue(string(normalized))
}
