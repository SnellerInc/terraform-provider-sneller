package datasource

import (
	"context"
	"fmt"
	"terraform-provider-sneller/sneller/api"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

type userDataSource struct {
	client *api.Client
}

type userDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	UserID      types.String `tfsdk:"user_id"`
	Email       types.String `tfsdk:"email"`
	IsEnabled   types.Bool   `tfsdk:"is_enabled"`
	IsAdmin     types.Bool   `tfsdk:"is_admin"`
	IsFederated types.Bool   `tfsdk:"is_federated"`
	Locale      types.String `tfsdk:"locale"`
	GivenName   types.String `tfsdk:"given_name"`
	FamilyName  types.String `tfsdk:"family_name"`
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Obtains a specific Sneller user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Terraform identifier.",
				Computed:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "User identifier.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address.",
				Computed:    true,
			},
			"is_enabled": schema.BoolAttribute{
				Description: "User enabled.",
				Computed:    true,
			},
			"is_admin": schema.BoolAttribute{
				Description: "Administrator.",
				Computed:    true,
			},
			"is_federated": schema.BoolAttribute{
				Description: "User is using a federated identity provider.",
				Computed:    true,
			},
			"locale": schema.StringAttribute{
				Description: "User's locale.",
				Computed:    true,
			},
			"given_name": schema.StringAttribute{
				Description: "User's given name.",
				Computed:    true,
			},
			"family_name": schema.StringAttribute{
				Description: "User's family name.",
				Computed:    true,
			},
		},
	}
}

func (d *userDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*api.Client)
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data userDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.UserID.ValueString()

	tenantInfo, err := d.client.Tenant(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info: %v", err.Error()),
		)
		return
	}

	user, err := d.client.User(ctx, userID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get user",
			fmt.Sprintf("Unable to get user %q: %v", userID, err.Error()),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", tenantInfo.TenantID, user.UserID))
	data.UserID = types.StringValue(user.UserID)
	data.Email = types.StringValue(user.Email)
	data.IsEnabled = types.BoolValue(user.IsEnabled)
	data.IsAdmin = types.BoolValue(slices.Contains(user.Groups, api.AdminGroup))
	data.IsFederated = types.BoolValue(user.IsFederated)
	data.Locale = types.StringValue(user.Locale)
	data.GivenName = types.StringValue(user.GivenName)
	data.FamilyName = types.StringValue(user.FamilyName)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
