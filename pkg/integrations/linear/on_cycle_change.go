package linear

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

type OnCycleChange struct{}

type OnCycleChangeConfiguration struct {
	TeamID string `json:"teamId" mapstructure:"teamId"`
}

func (t *OnCycleChange) Name() string {
	return "linear.onCycleChange"
}

func (t *OnCycleChange) Label() string {
	return "On Cycle Change"
}

func (t *OnCycleChange) Description() string {
	return "Listen to cycle changes in Linear"
}

func (t *OnCycleChange) Documentation() string {
	return `The On Cycle Change trigger starts a workflow execution when issues are added to or removed from a Linear cycle.

## Use Cases

- **Sprint tracking**: Monitor when issues enter or leave a sprint cycle
- **Planning automation**: Automatically update related issues when cycle changes occur
- **Notification workflows**: Notify teams when cycle assignments change
- **Reporting**: Track velocity and cycle progress

## Configuration

- **Team**: The Linear team to monitor for cycle changes (leave empty to listen to all teams)

## Event Data

Each cycle change event includes the issue data with updated cycle information.`
}

func (t *OnCycleChange) Icon() string {
	return "linear"
}

func (t *OnCycleChange) Color() string {
	return "blue"
}



func (t *OnCycleChange) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "teamId",
			Label:       "Team",
			Type:        configuration.FieldTypeIntegrationResource,
			Required:    false,
			Description: "The Linear team to monitor for cycle changes (leave empty to listen to all teams)",
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type: "team",
				},
			},
		},
	}
}

func (t *OnCycleChange) Setup(ctx core.TriggerContext) error {
	return nil
}

func (t *OnCycleChange) Hooks() []core.Hook {
	return []core.Hook{}
}

func (t *OnCycleChange) HandleHook(ctx core.TriggerHookContext) (map[string]any, error) {
	return nil, nil
}

func (t *OnCycleChange) HandleWebhook(ctx core.WebhookRequestContext) (int, *core.WebhookResponseBody, error) {
	var config OnCycleChangeConfiguration
	if err := mapstructure.Decode(ctx.Configuration, &config); err != nil {
		ctx.Logger.Errorf("Failed to decode configuration: %v", err)
		return http.StatusInternalServerError, nil, fmt.Errorf("failed to decode configuration: %w", err)
	}

	var payload LinearWebhookPayload
	if err := json.Unmarshal(ctx.Body, &payload); err != nil {
		ctx.Logger.Errorf("Failed to parse request body: %v", err)
		return http.StatusBadRequest, nil, fmt.Errorf("error parsing request body: %v", err)
	}

	//
	// Cycle changes come as update events on Issue or Cycle entities
	//
	if payload.Type != "Issue" && payload.Type != "Cycle" {
		ctx.Logger.Infof("Ignoring event - type %q is not a cycle change", payload.Type)
		return http.StatusOK, nil, nil
	}

	if payload.Type == "Issue" && payload.Action != "update" {
		ctx.Logger.Infof("Ignoring event - action %q is not an update", payload.Action)
		return http.StatusOK, nil, nil
	}

	//
	// If team filter is set, only process events for the specified team
	//
	if config.TeamID != "" {
		teamData, ok := payload.Data["team"]
		if !ok {
			ctx.Logger.Infof("Ignoring event - no team information in payload")
			return http.StatusOK, nil, nil
		}
		teamMap, ok := teamData.(map[string]any)
		if !ok {
			ctx.Logger.Infof("Ignoring event - invalid team data in payload")
			return http.StatusOK, nil, nil
		}
		teamID, ok := teamMap["id"].(string)
		if !ok || teamID != config.TeamID {
			ctx.Logger.Infof("Ignoring event - team does not match configured team")
			return http.StatusOK, nil, nil
		}
	}

	if err := ctx.Events.Emit("linear.issue.cycleChange", payload.Data); err != nil {
		ctx.Logger.Errorf("Failed to emit event: %v", err)
		return http.StatusInternalServerError, nil, fmt.Errorf("error emitting event: %v", err)
	}

	return http.StatusOK, nil, nil
}

func (t *OnCycleChange) Cleanup(ctx core.TriggerContext) error {
	return nil
}
