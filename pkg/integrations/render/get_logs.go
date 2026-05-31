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

const GetLogsPayloadType = "render.service.logs"

type GetLogs struct{}

type GetLogsConfiguration struct {
	Service string `json:"service" mapstructure:"service"`
	Lines   int    `json:"lines" mapstructure:"lines"`
}

func (c *GetLogs) Name() string {
	return "render.getLogs"
}

func (c *GetLogs) Label() string {
	return "Get Logs"
}

func (c *GetLogs) Description() string {
	return "Fetch recent log lines for a Render service"
}

func (c *GetLogs) Documentation() string {
	return `The Get Logs component fetches recent log lines for a Render service.

## Use Cases

- **Debugging**: Fetch logs after a failed deploy to diagnose issues
- **Monitoring**: Periodically check service health via log analysis
- **Incident response**: Pull recent logs when alerts fire

## Configuration

- **Service**: Render service to fetch logs from
- **Lines**: Number of recent log lines to retrieve (default: 50)

## Output

Emits a ` + "`render.service.logs`" + ` payload containing the log entries.`
}

func (c *GetLogs) Icon() string {
	return "file-text"
}

func (c *GetLogs) Color() string {
	return "gray"
}

func (c *GetLogs) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *GetLogs) Configuration() []configuration.Field {
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
			Description: "Render service to fetch logs from",
		},
		{
			Name:        "lines",
			Label:       "Log Lines",
			Type:        configuration.FieldTypeInt,
			Required:    false,
			Default:     50,
			Description: "Number of recent log lines to retrieve",
		},
	}
}

func decodeGetLogsConfiguration(configuration any) (GetLogsConfiguration, error) {
	spec := GetLogsConfiguration{}
	if err := mapstructure.Decode(configuration, &spec); err != nil {
		return GetLogsConfiguration{}, fmt.Errorf("failed to decode configuration: %w", err)
	}

	spec.Service = strings.TrimSpace(spec.Service)
	if spec.Service == "" {
		return GetLogsConfiguration{}, fmt.Errorf("service is required")
	}

	if spec.Lines <= 0 {
		spec.Lines = 50
	}

	return spec, nil
}

func (c *GetLogs) Setup(ctx core.SetupContext) error {
	_, err := decodeGetLogsConfiguration(ctx.Configuration)
	return err
}

func (c *GetLogs) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *GetLogs) Execute(ctx core.ExecutionContext) error {
	spec, err := decodeGetLogsConfiguration(ctx.Configuration)
	if err != nil {
		return err
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return err
	}

	logs, err := client.GetLogs(spec.Service, spec.Lines)
	if err != nil {
		return err
	}

	logEntries := make([]any, 0, len(logs))
	for _, entry := range logs {
		logEntries = append(logEntries, map[string]any{
			"timestamp": entry.Timestamp,
			"message":   entry.Message,
		})
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		GetLogsPayloadType,
		[]any{map[string]any{
			"serviceId": spec.Service,
			"lines":     logEntries,
			"count":     len(logEntries),
		}},
	)
}

func (c *GetLogs) HandleWebhook(ctx core.WebhookRequestContext) (int, *core.WebhookResponseBody, error) {
	return http.StatusOK, nil, nil
}

func (c *GetLogs) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *GetLogs) Cleanup(ctx core.SetupContext) error {
	return nil
}

func (c *GetLogs) Hooks() []core.Hook {
	return []core.Hook{}
}

func (c *GetLogs) HandleHook(ctx core.ActionHookContext) error {
	return nil
}
