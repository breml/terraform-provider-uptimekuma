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
	_ resource.Resource                = &NotificationNostrResource{}
	_ resource.ResourceWithImportState = &NotificationNostrResource{}
)

// NewNotificationNostrResource returns a new instance of the Nostr notification resource.
func NewNotificationNostrResource() resource.Resource {
	return &NotificationNostrResource{}
}

// NotificationNostrResource defines the resource implementation.
type NotificationNostrResource struct {
	client *kuma.Client
}

// NotificationNostrResourceModel describes the resource data model.
type NotificationNostrResourceModel struct {
	NotificationBaseModel

	Sender     types.String `tfsdk:"sender"`
	Recipients types.String `tfsdk:"recipients"`
	Relays     types.String `tfsdk:"relays"`
}

// Metadata returns the metadata for the resource.
func (*NotificationNostrResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_nostr"
}

// Schema returns the schema for the resource.
func (*NotificationNostrResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Nostr notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"sender": schema.StringAttribute{
				MarkdownDescription: "Sender private key in Nostr format (nsec encoded)",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"recipients": schema.StringAttribute{
				MarkdownDescription: "Newline-delimited list of recipient public keys (npub encoded)",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"relays": schema.StringAttribute{
				MarkdownDescription: "Newline-delimited list of Nostr relay URLs",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Nostr notification resource with the API client.
func (r *NotificationNostrResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *kuma.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	r.client = client
}

// Create creates a new Nostr notification resource.
func (r *NotificationNostrResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationNostrResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nostr := notification.Nostr{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		NostrDetails: notification.NostrDetails{
			Sender:     data.Sender.ValueString(),
			Recipients: data.Recipients.ValueString(),
			Relays:     data.Relays.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, nostr)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Nostr notification resource.
func (r *NotificationNostrResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationNostrResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, err := r.client.GetNotification(ctx, id)
	// Handle error.
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	nostr := notification.Nostr{}
	err = base.As(&nostr)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "nostr"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(nostr.Name)
	data.IsActive = types.BoolValue(nostr.IsActive)
	data.IsDefault = types.BoolValue(nostr.IsDefault)
	data.ApplyExisting = types.BoolValue(nostr.ApplyExisting)

	data.Sender = types.StringValue(nostr.Sender)
	data.Recipients = types.StringValue(nostr.Recipients)
	data.Relays = types.StringValue(nostr.Relays)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Nostr notification resource.
func (r *NotificationNostrResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationNostrResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nostr := notification.Nostr{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		NostrDetails: notification.NostrDetails{
			Sender:     data.Sender.ValueString(),
			Recipients: data.Recipients.ValueString(),
			Relays:     data.Relays.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, nostr)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Nostr notification resource.
func (r *NotificationNostrResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationNostrResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*NotificationNostrResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
