package datasource

import (
	"context"
	"fmt"
	"terraform-provider-sneller/sneller/api"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &usersDataSource{}
	_ datasource.DataSourceWithConfigure = &usersDataSource{}
)

type usersDataSource struct {
	client *api.Client
}

type usersDataSourceModel struct {
	ID    types.String               `tfsdk:"id"`
	Users []usersUserDataSourceModel `tfsdk:"users"`
}

type usersUserDataSourceModel struct {
	UserID    types.String `tfsdk:"user_id"`
	Email     types.String `tfsdk:"email"`
	IsEnabled types.Bool   `tfsdk:"is_enabled"`
}

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides all users.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Terraform identifier.",
				Computed:    true,
			},
			"users": schema.SetNestedAttribute{
				Description: "Set of users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_id": schema.StringAttribute{
							Description: "User identifier.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "Email address",
							Computed:    true,
						},
						"is_enabled": schema.BoolAttribute{
							Description: "User enabled.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*api.Client)
}

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data usersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantInfo, err := d.client.Tenant(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info: %v", err.Error()),
		)
		return
	}

	users, err := d.client.Users(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get users",
			fmt.Sprintf("Unable to get user list: %v", err.Error()),
		)
		return
	}

	data.ID = types.StringValue(tenantInfo.TenantID)
	data.Users = make([]usersUserDataSourceModel, 0, len(users))
	for _, user := range users {
		data.Users = append(data.Users, usersUserDataSourceModel{
			UserID:    types.StringValue(user.UserID),
			Email:     types.StringValue(user.Email),
			IsEnabled: types.BoolValue(user.IsEnabled),
		})
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
