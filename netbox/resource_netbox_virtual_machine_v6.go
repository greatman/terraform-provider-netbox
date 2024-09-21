package netbox

import (
	"context"
	"fmt"
	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/virtualization"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceNetboxVirtualMachineV6 struct {
	ApiClient *client.NetBoxAPI
}

type netboxVirtualMachineModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ClusterID types.Int64  `tfsdk:"cluster_id"`
	Status    types.String `tfsdk:"status"`
}

func (r *resourceNetboxVirtualMachineV6) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_virtual_machine"
}

func (r *resourceNetboxVirtualMachineV6) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Virtual Machine ID",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Virtual Machine",
			},
			"cluster_id": schema.Int64Attribute{
				Description: "Cluster ID",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the machine",
				Required:    true,
			},
		},
	}
}

func (r *resourceNetboxVirtualMachineV6) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//Retrieve data from the Plan
	var data netboxVirtualMachineModel
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

	vm := models.WritableVirtualMachineWithConfigContext{
		Name:   data.Name.ValueStringPointer(),
		Status: data.Status.String(),
	}
	if !data.ClusterID.IsNull() {
		vm.Cluster = data.ClusterID.ValueInt64Pointer()
	}

	params := virtualization.NewVirtualizationVirtualMachinesCreateParams().WithData(&vm)
	res, err := r.ApiClient.Virtualization.VirtualizationVirtualMachinesCreate(params, nil)
	if err != nil {
		response.Diagnostics.AddError("An error occured while creating the Virtual Machine", err.Error())
		return
	}
	data.ID = types.Int64Value(res.GetPayload().ID)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceNetboxVirtualMachineV6) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	if r.ApiClient == nil {
		response.Diagnostics.AddError(
			"Read: Unconfigured API Client",
			"Expected configured API Client. Please report this issue to the provider developers.",
		)
		return
	}

	var data netboxVirtualMachineModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	params := virtualization.NewVirtualizationVirtualMachinesReadParams().WithID(data.ID.ValueInt64())
	res, err := r.ApiClient.Virtualization.VirtualizationVirtualMachinesRead(params, nil)
	if err != nil {
		if errresp, ok := err.(*virtualization.VirtualizationVirtualMachinesReadDefault); ok {
			errorcode := errresp.Code()
			if errorcode == 404 {
				response.State.RemoveResource(ctx)
				return
			}
		}
	}
	vm := res.GetPayload()
	data.Name = types.StringPointerValue(vm.Name)

	if vm.Cluster != nil {
		data.ClusterID = types.Int64Value(vm.Cluster.ID)
	} else {
		data.ClusterID = types.Int64Null()
	}

	data.Status = types.StringPointerValue(vm.Status.Value)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceNetboxVirtualMachineV6) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *resourceNetboxVirtualMachineV6) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *resourceNetboxVirtualMachineV6) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *resourceNetboxVirtualMachineV6) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
