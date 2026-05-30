package linear

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

type OnIssueLabelChange struct{}

type OnIssueLabelChangeConfiguration struct {
	TeamID string `json:"teamId" mapstructure:"teamId"`
}

func (t *OnIssueLabelChange) Name() string {
	return "linear.onIssueLabelChange"
}

func (t *OnIssueLabelChange) Label() string {
	return "On Issue Label Change"
}

func (t *OnIssueLabelChange) Description() string {
	return "Listen to issue label changes in Linear"
}

func (t *OnIssueLabelChange) Documentation() string {
	return `The On Issue Label Change trigger starts a workflow execution when labels are added or removed from a Linear issue.

## Use Cases

- **Auto-triage**: Automatically assign issues based on labels
- **Workflow routing**: Route issues to different teams based on labels
- **Notification workflows**: Notify relevant stakeholders when specific labels are applied
- **Cross-system sync**: Mirror label changes to other project management tools

## Configuration

- **Team**: The Linear team to monitor for label changes (leave empty to listen to all teams)

## Event Data

Each label change event includes the full issue data with updated labels.`
}

func (t *OnIssueLabelChange) Icon() string {
	return "linear"
}

func (t *OnIssueLabelChange) Color() string {
	return "blue"
}



func (t *OnIssueLabelChange) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "teamId",
			Label:       "Team",
			Type:        configuration.FieldTypeIntegrationResource,
			Required:    false,
			Description: "The Linear team to monitor for label changes (leave empty to listen to all teams)",
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type: "team",
				},
			},
		},
	}
}

func (t *OnIssueLabelChange) Setup(ctx core.TriggerContext) error {
	return nil
}

func (t *OnIssueLabelChange) Hooks() []core.Hook {
	return []core.Hook{}
}

func (t *OnIssueLabelChange) HandleHook(ctx core.TriggerHookContext) (map[string]any, error) {
	return nil, nil
}

func (t *OnIssueLabelChange) HandleWebhook(ctx core.WebhookRequestContext) (int, *core.WebhookResponseBody, error) {
	var config OnIssueLabelChangeConfiguration
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
	// Label changes come as update events on Issue or IssueLabel entities
	//
	if payload.Type != "Issue" && payload.Type != "IssueLabel" {
		ctx.Logger.Infof("Ignoring event - type %q is not an issue label change", payload.Type)
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

	if err := ctx.Events.Emit("linear.issue.labelChange", payload.Data); err != nil {
		ctx.Logger.Errorf("Failed to emit event: %v", err)
		return http.StatusInternalServerError, nil, fmt.Errorf("error emitting event: %v", err)
	}

	return http.StatusOK, nil, nil
}

func (t *OnIssueLabelChange) Cleanup(ctx core.TriggerContext) error {
	return nil
}
