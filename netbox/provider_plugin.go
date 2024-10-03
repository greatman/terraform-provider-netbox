package netbox

import (
	"context"
	"fmt"
	netboxclient "github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/status"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
	"os"
	"strconv"
	"strings"
)

func New() provider.Provider {
	return &netboxProvider{}
}

type netboxProvider struct {
	client *netboxclient.NetBoxAPI
}

type netboxProviderModel struct {
	ApiToken                    types.String `tfsdk:"api_token"`
	ServerUrl                   types.String `tfsdk:"server_url"`
	SkipVersionCheck            types.Bool   `tfsdk:"skip_version_check"`
	AllowInsecureHttps          types.Bool   `tfsdk:"allow_insecure_https"`
	Headers                     types.Map    `tfsdk:"headers"`
	StripTrailingSlashesFromUrl types.Bool   `tfsdk:"strip_trailing_slashes_from_url"`
	RequestTimeout              types.Int32  `tfsdk:"request_timeout"`
}

func (p *netboxProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var data netboxProviderModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	apiToken := os.Getenv("NETBOX_API_TOKEN")
	serverUrl := os.Getenv("NETBOX_SERVER_URL")
	requestTimeoutString := os.Getenv("NETBOX_REQUEST_TIMEOUT")
	var requestTimeout int32
	if requestTimeoutString == "" {
		requestTimeout = 10
	} else {
		requestTimeoutConversion, err := strconv.ParseInt(requestTimeoutString, 10, 32)

		if err != nil {
			response.Diagnostics.AddError("Invalid Request Timeout environment variable.", "TODO")
		}
		requestTimeout = int32(requestTimeoutConversion)
	}

	if !data.ApiToken.IsNull() {
		apiToken = data.ApiToken.ValueString()
	}

	if !data.ServerUrl.IsNull() {
		serverUrl = data.ServerUrl.ValueString()
	}

	if !data.RequestTimeout.IsNull() {
		requestTimeout = data.RequestTimeout.ValueInt32()
	}

	if apiToken == "" {
		response.Diagnostics.AddError(
			"Missing API Token Configuration",
			"TODO DETAIL")
	}

	if serverUrl == "" {
		response.Diagnostics.AddError(
			"Missing server URL configuration.",
			"TODO details")
	}

	config := Config{
		APIToken:       apiToken,
		ServerURL:      serverUrl,
		RequestTimeout: int(requestTimeout),
	}
	netboxClient, clientError := config.Client()
	if clientError != nil {
		response.Diagnostics.AddError("Error creating netbox client.", clientError.Error())
		return
	}

	if !data.SkipVersionCheck.ValueBool() {
		req := status.NewStatusListParams()
		res, err := netboxClient.Status.StatusList(req, nil)
		if err != nil {
			response.Diagnostics.AddError("Error getting netbox status.", err.Error())
			return
		}
		netboxVersion := res.GetPayload().(map[string]interface{})["netbox-version"].(string)

		supportedVersions := []string{"4.0.0", "4.0.1", "4.0.2", "4.0.3", "4.0.5", "4.0.6", "4.0.7", "4.0.8", "4.0.9", "4.0.10"}

		if !slices.Contains(supportedVersions, netboxVersion) {
			response.Diagnostics.AddWarning("Possibly unsupported Netbox version", fmt.Sprintf("Your Netbox version is v%v. The provider was successfully tested against the following versions:\n\n  %v\n\nUnexpected errors may occur.", netboxVersion, strings.Join(supportedVersions, ", ")))
		}
	}
	p.client = netboxClient
	response.ResourceData = p
	response.DataSourceData = netboxClient
}

func (p *netboxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *netboxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{

		func() resource.Resource {
			return &resourceNetboxSitev6{}
		},
	}
}

func (p *netboxProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "netbox"
	resp.Version = "DEV"
}
func (p *netboxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				MarkdownDescription: "Netbox API authentication token. Can be set via the `NETBOX_API_TOKEN` environment variable.",
				Optional:            true,
			},
			"server_url": schema.StringAttribute{
				MarkdownDescription: "Location of Netbox server including scheme (http or https) and optional port. Can be set via the `NETBOX_SERVER_URL` environment variable.",
				Optional:            true,
			},
			"skip_version_check": schema.BoolAttribute{
				MarkdownDescription: "If true, do not try to determine the running Netbox version at provider startup. Disables warnings about possibly unsupported Netbox version. Also useful for local testing on terraform plans. Can be set via the `NETBOX_SKIP_VERSION_CHECK` environment variable. Defaults to `false`.",
				Optional:            true,
			},
			"allow_insecure_https": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Flag to set whether to allow https with invalid certificates. Can be set via the `NETBOX_ALLOW_INSECURE_HTTPS` environment variable. Defaults to `false`.",
			},
			"headers": schema.MapAttribute{
				Optional:            true,
				MarkdownDescription: "Set these header on all requests to Netbox. Can be set via the `NETBOX_HEADERS` environment variable.",
				ElementType:         types.StringType,
			},
			"strip_trailing_slashes_from_url": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "If true, strip trailing slashes from the `server_url` parameter and print a warning when doing so. Note that using trailing slashes in the `server_url` parameter will usually lead to errors. Can be set via the `NETBOX_STRIP_TRAILING_SLASHES_FROM_URL` environment variable. Defaults to `true`.",
			},
			"request_timeout": schema.Int32Attribute{
				Optional:            true,
				MarkdownDescription: "Netbox API HTTP request timeout in seconds. Can be set via the `NETBOX_REQUEST_TIMEOUT` environment variable.",
			},
		},
	}
}
