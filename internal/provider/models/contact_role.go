package models

import (
	"context"
	"github.com/e-breuninger/terraform-provider-netbox/internal/provider/helpers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netbox-community/go-netbox/v4"
)

type ContactRoleTerraformModel struct {
	CustomFields types.Map    `tfsdk:"custom_fields"`
	Description  types.String `tfsdk:"description"`
	Id           types.Int32  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Slug         types.String `tfsdk:"slug"`
	Tags         types.List   `tfsdk:"tags"`
}

func (data *ContactRoleTerraformModel) ReadAPI(ctx context.Context, contactRole *netbox.ContactRole) diag.Diagnostics {
	var diags = diag.Diagnostics{}
	data.Id = types.Int32Value(contactRole.Id)
	data.Name = types.StringValue(contactRole.Name)
	data.Slug = types.StringValue(contactRole.Slug)
	data.Description = types.StringPointerValue(contactRole.Description)
	tags := helpers.ReadTagsFromAPI(contactRole.Tags)
	tagsdata, diagdata := types.ListValueFrom(ctx, types.Int32Type, tags)
	if diagdata.HasError() {
		diags.AddError(
			"Error while reading Webhook",
			"") //TODO Better handling
		return diags
	}
	data.Tags = tagsdata

	customFields, diagData := types.MapValueFrom(ctx, types.StringType, helpers.ReadCustomFieldsFromAPI(contactRole.CustomFields))

	//Let's only add custom fields that we know
	if data.CustomFields.IsUnknown() || data.CustomFields.IsNull() {
		data.CustomFields = customFields
	} else {
		for k, _ := range data.CustomFields.Elements() {
			if val, ok := customFields.Elements()[k]; ok {
				data.CustomFields.Elements()[k] = val
			}
		}
	}
	if diagData.HasError() {
		diags.Append()
	}
	return diags
}
