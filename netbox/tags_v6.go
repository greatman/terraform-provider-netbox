package netbox

import (
	"fmt"
	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/extras"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var tagsV6Schema = schema.SetAttribute{
	Optional:    true,
	ElementType: types.StringType,
}

func getNestedTagListFromResourceDataSetV6(client *client.NetBoxAPI, d types.Set, diag diag.Diagnostics) []*models.NestedTag {
	elements := make([]types.String, 0, len(d.Elements()))
	var tags = []*models.NestedTag{}
	for _, tag := range elements {
		params := extras.NewExtrasTagsListParams()
		params.Name = tag.ValueStringPointer()
		limit := int64(2)
		params.Limit = &limit
		res, err := client.Extras.ExtrasTagsList(params, nil)
		if err != nil {
			diag.AddError(
				fmt.Sprintf("Error retrieving tag %s from netbox", tag.String()),
				fmt.Sprintf("API Error trying to retrieve tag %s from netbox", tag.String()),
			)
			return tags
		}
		payload := res.GetPayload()
		switch *payload.Count {
		case int64(0):
			diag.AddError(
				fmt.Sprintf("Error retrieving tag %s from netbox", tag.String()),
				fmt.Sprintf("Could not locate referenced tag %s in netbox", tag.String()),
			)
			return tags
		case int64(1):
			tags = append(tags, &models.NestedTag{
				Name: payload.Results[0].Name,
				Slug: payload.Results[0].Slug,
			})
		default:
			diag.AddError(
				fmt.Sprintf("Error retrieving tag %s from netbox", tag.String()),
				fmt.Sprintf("Could not map tag %s to unique tag in netbox", tag.String()),
			)
			return tags
		}
	}
	return tags
}
