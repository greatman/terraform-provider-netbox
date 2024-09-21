package netbox

import (
	"context"
	"fmt"
	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/dcim"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type netboxSiteModel struct {
	ID              types.Int64   `tfsdk:"id"`
	Name            types.String  `tfsdk:"name"`
	Slug            types.String  `tfsdk:"slug"`
	Status          types.String  `tfsdk:"status"`
	Description     types.String  `tfsdk:"description"`
	Facility        types.String  `tfsdk:"facility"`
	Longitude       types.Float64 `tfsdk:"longitude"`
	Latitude        types.Float64 `tfsdk:"latitude"`
	PhysicalAddress types.String  `tfsdk:"physical_address"`
	ShippingAddress types.String  `tfsdk:"shipping_address"`
	RegionID        types.Int64   `tfsdk:"region_id"`
	GroupID         types.Int64   `tfsdk:"group_id"`
	TenantID        types.Int64   `tfsdk:"tenant_id"`
	Timezone        types.String  `tfsdk:"timezone"`
	ASNIDs          types.Set     `tfsdk:"asn_ids"`
	Tags            types.Set     `tfsdk:"tags"`
}

func generateWritableSite(data *netboxSiteModel, api *client.NetBoxAPI, diag diag.Diagnostics) models.WritableSite {
	site := models.WritableSite{}
	site.Name = data.Name.ValueStringPointer()

	site.Slug = data.Slug.ValueStringPointer()

	site.Status = data.Status.ValueString()

	if !data.Description.IsNull() {
		site.Description = data.Description.ValueString()
	}

	if !data.Facility.IsNull() {
		site.Facility = data.Facility.ValueString()
	}

	if !data.Longitude.IsNull() {
		site.Longitude = data.Longitude.ValueFloat64Pointer()
	}

	if !data.Latitude.IsNull() {
		site.Latitude = data.Latitude.ValueFloat64Pointer()
	}

	if !data.PhysicalAddress.IsNull() {
		site.PhysicalAddress = data.PhysicalAddress.ValueString()
	}

	if !data.ShippingAddress.IsNull() {
		site.ShippingAddress = data.ShippingAddress.ValueString()
	}

	if !data.RegionID.IsNull() {
		site.Region = data.RegionID.ValueInt64Pointer()
	}

	if !data.GroupID.IsNull() {
		site.Group = data.GroupID.ValueInt64Pointer()
	}

	if !data.TenantID.IsNull() {
		site.Tenant = data.TenantID.ValueInt64Pointer()
	}

	if !data.Tags.IsNull() {
		site.Tags = getNestedTagListFromResourceDataSetV6(api, data.Tags, diag)
	} else {
		site.Tags = []*models.NestedTag{}
	}

	if !data.Timezone.IsNull() {
		site.TimeZone = data.Timezone.ValueStringPointer()
	}

	if !data.ASNIDs.IsNull() {
		site.Asns = toInt64ListV6(data.ASNIDs)
	} else {
		site.Asns = []int64{}
	}
	//TODO Support custom fields
	return site
}

type resourceNetboxSitev6 struct {
	ApiClient *client.NetBoxAPI
}

func (r *resourceNetboxSitev6) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	api_client, ok := request.ProviderData.(*client.NetBoxAPI)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.NetBoxAPI, got: %T, Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	r.ApiClient = api_client
}

func (r *resourceNetboxSitev6) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *resourceNetboxSitev6) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_site"
}

func (r *resourceNetboxSitev6) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Site ID",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "",
			},
			"slug": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				}, //TODO : Slug generator, didn't find how to do it yet.
			},
			"status": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(resourceNetboxSiteStatusOptions...),
				},
				MarkdownDescription: buildValidValueDescription(resourceNetboxSiteStatusOptions),
				Default:             stringdefault.StaticString("active"),
			},
			"description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			"facility": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
				},
			},
			"longitude": schema.Float64Attribute{
				Optional:            true,
				MarkdownDescription: "The longitude of the site.",
			},
			"latitude": schema.Float64Attribute{
				Optional:            true,
				MarkdownDescription: "The latitude of the site.",
			},
			"physical_address": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The physical address of the site.",
			},
			"shipping_address": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The shipping address of the site.",
			},
			"region_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "", //TODO Description
			},
			"group_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "", //TODO Description
			},
			"tenant_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "", //TODO Description
			},
			"timezone": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "", //TODO Description
			},
			"asn_ids": schema.SetAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
			},
			tagsKey: tagsV6Schema,
		},
	}
}

func (r *resourceNetboxSitev6) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data netboxSiteModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	//Validate that the APIClient exist.
	if r.ApiClient == nil {
		response.Diagnostics.AddError(
			"Create: Unconfigured API Client",
			"Expected configured API Client. Please report this issue to the provider developers.",
		)
		return
	}

	site := generateWritableSite(&data, r.ApiClient, response.Diagnostics)

	params := dcim.NewDcimSitesCreateParams().WithData(&site)

	res, err := r.ApiClient.Dcim.DcimSitesCreate(params, nil)
	if err != nil {
		response.Diagnostics.AddError(
			"Error while creating the Site",
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(res.GetPayload().ID)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceNetboxSitev6) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	if r.ApiClient == nil {
		response.Diagnostics.AddError(
			"Read: Unconfigured API Client",
			"Expected configured API Client. Please report this issue to the provider developers.",
		)
		return
	}

	if response.Diagnostics.HasError() {
		return
	}

	var data netboxSiteModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	params := dcim.NewDcimSitesReadParams().WithID(data.ID.ValueInt64())
	res, err := r.ApiClient.Dcim.DcimSitesRead(params, nil)
	if err != nil {
		if errresp, ok := err.(*dcim.DcimSitesReadDefault); ok {
			errorcode := errresp.Code()
			if errorcode == 404 {
				response.State.RemoveResource(ctx)
				return
			}
		}
	}

	site := res.GetPayload()
	data.Name = types.StringPointerValue(site.Name)
	data.Slug = types.StringPointerValue(site.Slug)
	data.Status = types.StringPointerValue(site.Status.Value)
	if site.Description != "" {
		data.Description = types.StringValue(site.Description)
	}

	if site.Facility != "" {
		data.Facility = types.StringValue(site.Facility)
	}
	data.Longitude = types.Float64PointerValue(site.Longitude)
	data.Latitude = types.Float64PointerValue(site.Latitude)
	data.PhysicalAddress = types.StringValue(site.PhysicalAddress)
	data.ShippingAddress = types.StringValue(site.ShippingAddress)
	if site.Region != nil {
		data.RegionID = types.Int64Value(site.Region.ID)
	}

	if site.Group != nil {
		data.GroupID = types.Int64Value(site.Group.ID)
	}

	if site.Tenant != nil {
		data.TenantID = types.Int64Value(site.Tenant.ID)
	}

	data.Timezone = types.StringPointerValue(site.TimeZone)
	asns, d := types.SetValueFrom(ctx, types.Int64Type, getIDsFromNestedASNList(site.Asns))
	response.Diagnostics.Append(d...)
	data.ASNIDs = asns

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)

}

func (r *resourceNetboxSitev6) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	if r.ApiClient == nil {
		response.Diagnostics.AddError(
			"Read: Unconfigured API Client",
			"Expected configured API Client. Please report this issue to the provider developers.",
		)
		return
	}

	if response.Diagnostics.HasError() {
		return
	}

	var data netboxSiteModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	site := generateWritableSite(&data, r.ApiClient, response.Diagnostics)

	params := dcim.NewDcimSitesPartialUpdateParams().WithID(data.ID.ValueInt64()).WithData(&site)

	_, err := r.ApiClient.Dcim.DcimSitesPartialUpdate(params, nil)
	if err != nil {
		response.Diagnostics.AddError(
			"Error while updating the Site",
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceNetboxSitev6) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	if r.ApiClient == nil {
		response.Diagnostics.AddError(
			"Read: Unconfigured API Client",
			"Expected configured API Client. Please report this issue to the provider developers.",
		)
		return
	}

	if response.Diagnostics.HasError() {
		return
	}

	var data netboxSiteModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	params := dcim.NewDcimSitesDeleteParams().WithID(data.ID.ValueInt64())
	_, err := r.ApiClient.Dcim.DcimSitesDelete(params, nil)
	if err != nil {
		if errresp, ok := err.(*dcim.DcimSitesDeleteDefault); ok {
			if errresp.Code() == 404 {
				response.State.RemoveResource(ctx)
			} else {
				response.Diagnostics.AddError("Unable to delete the site.", errresp.Error())
				return
			}
		}
	}
}
