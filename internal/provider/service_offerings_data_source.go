package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/virtfoundry/terraform-provider-virtfoundry/internal/virtfoundry"
)

var _ datasource.DataSource = &serviceOfferingsDataSource{}

type serviceOfferingsDataSource struct {
	client *virtfoundry.Client
}

type serviceOfferingsModel struct {
	Offerings types.List `tfsdk:"offerings"`
}

type serviceOfferingModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	CPU         types.Int64  `tfsdk:"cpu"`
	MemoryMi    types.Int64  `tfsdk:"memory_mi"`
	State       types.String `tfsdk:"state"`
}

var offeringAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"name":         types.StringType,
	"display_name": types.StringType,
	"cpu":          types.Int64Type,
	"memory_mi":    types.Int64Type,
	"state":        types.StringType,
}

func NewServiceOfferingsDataSource() datasource.DataSource { return &serviceOfferingsDataSource{} }

func (d *serviceOfferingsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "virtfoundry_service_offerings"
}

func (d *serviceOfferingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists VirtFoundry service offerings (CPU/memory catalog).",
		Attributes: map[string]schema.Attribute{
			"offerings": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true},
						"name":         schema.StringAttribute{Computed: true},
						"display_name": schema.StringAttribute{Computed: true},
						"cpu":          schema.Int64Attribute{Computed: true},
						"memory_mi":    schema.Int64Attribute{Computed: true},
						"state":        schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *serviceOfferingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*virtfoundry.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *virtfoundry.Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *serviceOfferingsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "VirtFoundry client is nil")
		return
	}
	items, err := d.client.ListServiceOfferings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("List service offerings failed", err.Error())
		return
	}
	elems := make([]attr.Value, len(items))
	for i, o := range items {
		obj, diags := types.ObjectValue(offeringAttrTypes, map[string]attr.Value{
			"id":           types.StringValue(o.ID),
			"name":         types.StringValue(o.Name),
			"display_name": types.StringValue(o.DisplayName),
			"cpu":          types.Int64Value(int64(o.CPU)),
			"memory_mi":    types.Int64Value(o.MemoryMi),
			"state":        types.StringValue(o.State),
		})
		resp.Diagnostics.Append(diags...)
		elems[i] = obj
	}
	list, diags := types.ListValue(types.ObjectType{AttrTypes: offeringAttrTypes}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, serviceOfferingsModel{Offerings: list})...)
}
