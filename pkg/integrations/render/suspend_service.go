package render

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

const SuspendServicePayloadType = "render.service.suspended"

type SuspendService struct{}

type SuspendServiceConfiguration struct {
	Service string `json:"service" mapstructure:"service"`
}

func (c *SuspendService) Name() string {
	return "render.suspendService"
}

func (c *SuspendService) Label() string {
	return "Suspend Service"
}

func (c *SuspendService) Description() string {
	return "Suspend a Render service to stop it and save costs"
}

func (c *SuspendService) Documentation() string {
	return `The Suspend Service component suspends a Render service, stopping it and reducing costs.

## Use Cases

- **Cost optimization**: Suspend non-production services during off-hours
- **Incident response**: Quickly stop a misbehaving service
- **Environment management**: Tear down temporary environments

## Configuration

- **Service**: Render service to suspend

## Output

Emits a ` + "`render.service.suspended`" + ` payload confirming the service was suspended.`
}

func (c *SuspendService) Icon() string {
	return "pause"
}

func (c *SuspendService) Color() string {
	return "gray"
}

func (c *SuspendService) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *SuspendService) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:     "service",
			Label:    "Service",
			Type:     configuration.FieldTypeIntegrationResource,
			Required: true,
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type: "service",
				},
			},
			Description: "Render service to suspend",
		},
	}
}

func decodeSuspendServiceConfiguration(configuration any) (SuspendServiceConfiguration, error) {
	spec := SuspendServiceConfiguration{}
	if err := mapstructure.Decode(configuration, &spec); err != nil {
		return SuspendServiceConfiguration{}, fmt.Errorf("failed to decode configuration: %w", err)
	}

	spec.Service = strings.TrimSpace(spec.Service)
	if spec.Service == "" {
		return SuspendServiceConfiguration{}, fmt.Errorf("service is required")
	}

	return spec, nil
}

func (c *SuspendService) Setup(ctx core.SetupContext) error {
	_, err := decodeSuspendServiceConfiguration(ctx.Configuration)
	return err
}

func (c *SuspendService) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *SuspendService) Execute(ctx core.ExecutionContext) error {
	spec, err := decodeSuspendServiceConfiguration(ctx.Configuration)
	if err != nil {
		return err
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return err
	}

	if err := client.SuspendService(spec.Service); err != nil {
		return err
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		SuspendServicePayloadType,
		[]any{map[string]any{
			"serviceId": spec.Service,
			"suspended": true,
		}},
	)
}

func (c *SuspendService) HandleWebhook(ctx core.WebhookRequestContext) (int, *core.WebhookResponseBody, error) {
	return http.StatusOK, nil, nil
}

func (c *SuspendService) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *SuspendService) Cleanup(ctx core.SetupContext) error {
	return nil
}

func (c *SuspendService) Hooks() []core.Hook {
	return []core.Hook{}
}

func (c *SuspendService) HandleHook(ctx core.ActionHookContext) error {
	return nil
}
