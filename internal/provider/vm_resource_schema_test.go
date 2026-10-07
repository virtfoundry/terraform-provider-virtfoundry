package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// The API defaults display_name to the VM name, so a configuration that omits it still gets a
// value back after apply. A non-computed attribute turns that into "inconsistent result after apply".
func TestVMDisplayNameIsOptionalAndComputed(t *testing.T) {
	var resp resource.SchemaResponse
	(&vmResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)

	attr, ok := resp.Schema.Attributes["display_name"].(schema.StringAttribute)
	if !ok {
		t.Fatal("display_name is not a string attribute")
	}
	if !attr.Optional || !attr.Computed {
		t.Fatalf("display_name must be Optional and Computed, got optional=%v computed=%v", attr.Optional, attr.Computed)
	}
}
