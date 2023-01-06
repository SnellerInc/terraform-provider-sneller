package resource

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-sneller/sneller/api"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

type userResource struct {
	client *api.Client
}

type userResourceModel struct {
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

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configure a Sneller user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Terraform identifier.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"user_id": schema.StringAttribute{
				Description: "User identifier.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address of the user.",
				Required:    true,
			},
			"is_enabled": schema.BoolAttribute{
				Description:   "Flag indicating whether the user is enabled.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{BoolDefaultValue(true)},
			},
			"is_admin": schema.BoolAttribute{
				Description:   "Flag indicating whether the user is an administrator.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{BoolDefaultValue(false)},
			},
			"is_federated": schema.BoolAttribute{
				Description: "Flag indicating whether the user is using an federated identity provider.",
				Computed:    true,
			},
			"locale": schema.StringAttribute{
				Description:   "User's locale (i.e. `en-US`).",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{StringDefaultValue("")},
			},
			"given_name": schema.StringAttribute{
				Description:   "User's given name.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{StringDefaultValue("")},
			},
			"family_name": schema.StringAttribute{
				Description:   "User's family name.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{StringDefaultValue("")},
			},
		},
	}
}

func (r *userResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*api.Client)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var data userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(data.ID.ValueString(), "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Cannot parse ID",
			fmt.Sprintf("Invalid ID %q", data.ID.ValueString()),
		)
		return
	}
	tenantID := parts[0]
	userID := parts[1]

	tenantInfo, err := r.client.Tenant(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info: %v", err.Error()),
		)
		return
	}

	user, err := r.client.User(ctx, userID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get user",
			fmt.Sprintf("Unable to get user %q: %v", userID, err.Error()),
		)
		return
	}
	if tenantInfo.TenantID != tenantID {
		resp.Diagnostics.AddError(
			"Invalid tenant",
			fmt.Sprintf("Expected tenant %s, but got %s", tenantID, tenantInfo.TenantID),
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

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	email := data.Email.ValueString()

	isAdmin := false
	if !data.IsAdmin.IsNull() {
		isAdmin = data.IsAdmin.ValueBool()
	}
	var locale, givenName, familyName *string
	if !data.Locale.IsNull() {
		locale = ptr(data.Locale.ValueString())
	}
	if !data.GivenName.IsNull() {
		givenName = ptr(data.GivenName.ValueString())
	}
	if !data.FamilyName.IsNull() {
		familyName = ptr(data.FamilyName.ValueString())
	}

	tenantInfo, err := r.client.Tenant(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info: %v", err.Error()),
		)
		return
	}

	userID, err := r.client.CreateUser(ctx, email, isAdmin, locale, givenName, familyName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot create user",
			fmt.Sprintf("Unable to create user %q: %v", email, err.Error()),
		)
		return
	}

	if locale == nil {
		*locale = ""
	}
	if givenName == nil {
		*givenName = ""
	}
	if familyName == nil {
		*familyName = ""
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", tenantInfo.TenantID, userID))
	data.UserID = types.StringValue(userID)
	data.Email = types.StringValue(email)
	data.IsEnabled = types.BoolValue(true)
	data.IsAdmin = types.BoolValue(isAdmin)
	data.IsFederated = types.BoolValue(false)
	data.Locale = types.StringValue(*locale)
	data.GivenName = types.StringValue(*givenName)
	data.FamilyName = types.StringValue(*familyName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, oldData userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &oldData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(data.ID.ValueString(), "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Cannot parse ID",
			fmt.Sprintf("Invalid ID %q", data.ID.ValueString()),
		)
		return
	}
	tenantID := parts[0]
	userID := parts[1]

	tenantInfo, err := r.client.Tenant(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info: %v", err.Error()),
		)
		return
	}
	if tenantInfo.TenantID != tenantID {
		resp.Diagnostics.AddError(
			"Invalid tenant",
			fmt.Sprintf("Expected tenant %s, but got %s", tenantID, tenantInfo.TenantID),
		)
		return
	}

	var isEnabled, isAdmin *bool
	if data.IsEnabled.ValueBool() != oldData.IsEnabled.ValueBool() {
		isEnabled = ptr(data.IsEnabled.ValueBool())
	}
	if data.IsAdmin.ValueBool() != oldData.IsAdmin.ValueBool() {
		isAdmin = ptr(data.IsAdmin.ValueBool())
	}

	var email, locale, givenName, familyName *string
	if data.Email.ValueString() != oldData.Email.ValueString() {
		email = ptr(data.Email.ValueString())
	}
	if data.Locale.ValueString() != oldData.Locale.ValueString() {
		locale = ptr(data.Locale.ValueString())
	}
	if data.GivenName.ValueString() != oldData.GivenName.ValueString() {
		givenName = ptr(data.GivenName.ValueString())
	}
	if data.FamilyName.ValueString() != oldData.FamilyName.ValueString() {
		familyName = ptr(data.FamilyName.ValueString())
	}

	err = r.client.UpdateUser(ctx, userID, email, isEnabled, isAdmin, locale, givenName, familyName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot update user",
			fmt.Sprintf("Unable to update user %s (%s): %v", userID, oldData.Email.ValueString(), err.Error()),
		)
		return
	}

	data.UserID = oldData.UserID
	data.IsFederated = oldData.IsFederated

	if isEnabled != nil {
		data.IsEnabled = types.BoolValue(*isEnabled)
	}
	if isAdmin != nil {
		data.IsAdmin = types.BoolValue(*isAdmin)
	}
	if email != nil {
		data.Email = types.StringValue(*email)
	}
	if locale != nil {
		data.Locale = types.StringValue(*locale)
	}
	if givenName != nil {
		data.GivenName = types.StringValue(*givenName)
	}
	if familyName != nil {
		data.FamilyName = types.StringValue(*familyName)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(data.ID.ValueString(), "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Cannot parse ID",
			fmt.Sprintf("Invalid ID %q", data.ID.ValueString()),
		)
		return
	}
	tenantID := parts[0]
	userID := parts[1]

	tenantInfo, err := r.client.Tenant(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info: %v", err.Error()),
		)
		return
	}
	if tenantInfo.TenantID != tenantID {
		resp.Diagnostics.AddError(
			"Invalid tenant",
			fmt.Sprintf("Expected tenant %s, but got %s", tenantID, tenantInfo.TenantID),
		)
		return
	}

	err = r.client.DeleteUser(ctx, userID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot delete user",
			fmt.Sprintf("Unable to delete user %s (%s): %v", userID, data.Email, err.Error()),
		)
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
