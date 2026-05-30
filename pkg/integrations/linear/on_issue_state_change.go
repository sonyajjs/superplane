package linear

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

type OnIssueStateChange struct{}

type OnIssueStateChangeConfiguration struct {
	TeamID string `json:"teamId" mapstructure:"teamId"`
}

func (t *OnIssueStateChange) Name() string {
	return "linear.onIssueStateChange"
}

func (t *OnIssueStateChange) Label() string {
	return "On Issue State Change"
}

func (t *OnIssueStateChange) Description() string {
	return "Listen to issue state changes in Linear"
}

func (t *OnIssueStateChange) Documentation() string {
	return `The On Issue State Change trigger starts a workflow execution when an issue's state changes in Linear.

## Use Cases

- **Status tracking**: Automate workflows when issues move between states (e.g., Todo → In Progress → Done)
- **Deployment triggers**: Start deployment workflows when issues are moved to "In Review"
- **Notification workflows**: Send notifications when issues are resolved
- **Cross-system sync**: Mirror Linear issue status changes to other project management tools

## Configuration

- **Team**: The Linear team to monitor for state changes

## Event Data

Each state change event includes:
- **action**: The action that occurred (update)
- **type**: The type of entity (Issue)
- **data**: The issue data including the new state

## Webhook Setup

This trigger requires a webhook configured in Linear. When setting up the webhook in Linear, select the "Issue" resource type and the "Update" action.`
}

func (t *OnIssueStateChange) Icon() string {
	return "linear"
}

func (t *OnIssueStateChange) Color() string {
	return "blue"
}



func (t *OnIssueStateChange) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "teamId",
			Label:       "Team",
			Type:        configuration.FieldTypeIntegrationResource,
			Required:    false,
			Description: "The Linear team to monitor for state changes (leave empty to listen to all teams)",
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type: "team",
				},
			},
		},
	}
}

func (t *OnIssueStateChange) Setup(ctx core.TriggerContext) error {
	return nil
}

func (t *OnIssueStateChange) Hooks() []core.Hook {
	return []core.Hook{}
}

func (t *OnIssueStateChange) HandleHook(ctx core.TriggerHookContext) (map[string]any, error) {
	return nil, nil
}

func (t *OnIssueStateChange) HandleWebhook(ctx core.WebhookRequestContext) (int, *core.WebhookResponseBody, error) {
	var config OnIssueStateChangeConfiguration
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
	// Only handle issue update events (state changes come as updates)
	//
	if payload.Type != "Issue" || payload.Action != "update" {
		ctx.Logger.Infof("Ignoring event - type %q action %q is not an issue state change", payload.Type, payload.Action)
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
			ctx.Logger.Infof("Ignoring event - team %q does not match configured team %q", teamID, config.TeamID)
			return http.StatusOK, nil, nil
		}
	}

	if err := ctx.Events.Emit("linear.issue.stateChange", payload.Data); err != nil {
		ctx.Logger.Errorf("Failed to emit event: %v", err)
		return http.StatusInternalServerError, nil, fmt.Errorf("error emitting event: %v", err)
	}

	return http.StatusOK, nil, nil
}

func (t *OnIssueStateChange) Cleanup(ctx core.TriggerContext) error {
	return nil
}
