package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationDingDingResource{}
	_ resource.ResourceWithImportState = &NotificationDingDingResource{}
)

func NewNotificationDingDingResource() resource.Resource {
	return &NotificationDingDingResource{}
}

type NotificationDingDingResource struct {
	client *kuma.Client
}

type NotificationDingDingResourceModel struct {
	NotificationBaseModel

	WebHookURL types.String `tfsdk:"webhook_url"`
	SecretKey  types.String `tfsdk:"secret_key"`
	Mentioning types.String `tfsdk:"mentioning"`
}

func (r *NotificationDingDingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_dingding"
}

func (r *NotificationDingDingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "DingDing notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "DingDing webhook URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "DingDing secret key for webhook signature",
				Optional:            true,
				Sensitive:           true,
			},
			"mentioning": schema.StringAttribute{
				MarkdownDescription: "DingDing mentioning configuration",
				Optional:            true,
			},
		}),
	}
}

func (r *NotificationDingDingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *kuma.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *NotificationDingDingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NotificationDingDingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dingding := notification.DingDing{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		DingDingDetails: notification.DingDingDetails{
			WebHookURL: data.WebHookURL.ValueString(),
			SecretKey:  data.SecretKey.ValueString(),
			Mentioning: data.Mentioning.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, dingding)
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.Id = types.Int64Value(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationDingDingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationDingDingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.Id.ValueInt64()

	base, err := r.client.GetNotification(ctx, id)
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	dingding := notification.DingDing{}
	err = base.As(&dingding)
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "DingDing"`, err.Error())
		return
	}

	data.Id = types.Int64Value(id)
	data.Name = types.StringValue(dingding.Name)
	data.IsActive = types.BoolValue(dingding.IsActive)
	data.IsDefault = types.BoolValue(dingding.IsDefault)
	data.ApplyExisting = types.BoolValue(dingding.ApplyExisting)

	data.WebHookURL = types.StringValue(dingding.WebHookURL)
	if dingding.SecretKey != "" {
		data.SecretKey = types.StringValue(dingding.SecretKey)
	}
	if dingding.Mentioning != "" {
		data.Mentioning = types.StringValue(dingding.Mentioning)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationDingDingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NotificationDingDingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dingding := notification.DingDing{
		Base: notification.Base{
			ID:            data.Id.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		DingDingDetails: notification.DingDingDetails{
			WebHookURL: data.WebHookURL.ValueString(),
			SecretKey:  data.SecretKey.ValueString(),
			Mentioning: data.Mentioning.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, dingding)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationDingDingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NotificationDingDingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}
}

func (r *NotificationDingDingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
