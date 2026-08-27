package modelutil

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

func TestNormalizedJSON(t *testing.T) {
	got := NormalizedJSON(json.RawMessage(`{ "b": 2, "a": 1 }`), "null")
	want := jsontypes.NewNormalizedValue(`{"a":1,"b":2}`)
	if !got.Equal(want) {
		t.Fatalf("got %q, want equivalent to %q", got.ValueString(), want.ValueString())
	}
	if string(RawJSON(got)) != got.ValueString() {
		t.Fatalf("RawJSON did not preserve normalized JSON")
	}
}

func TestNormalizedJSONEmpty(t *testing.T) {
	if got := NormalizedJSON(nil, "[]").ValueString(); got != "[]" {
		t.Fatalf("got %q, want []", got)
	}
}
