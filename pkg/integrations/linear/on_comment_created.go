package linear

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

type OnCommentCreated struct{}

type OnCommentCreatedConfiguration struct {
	TeamID string `json:"teamId" mapstructure:"teamId"`
}

func (t *OnCommentCreated) Name() string {
	return "linear.onCommentCreated"
}

func (t *OnCommentCreated) Label() string {
	return "On Comment Created"
}

func (t *OnCommentCreated) Description() string {
	return "Listen to new comment events in Linear"
}

func (t *OnCommentCreated) Documentation() string {
	return `The On Comment Created trigger starts a workflow execution when a new comment is created in Linear.

## Use Cases

- **Auto-respond**: Automatically respond to comments with specific keywords
- **Notification workflows**: Notify team members about new comments on specific issues
- **SLA tracking**: Track response times for customer-facing comments
- **Cross-system sync**: Mirror comments to other project management tools

## Configuration

- **Team**: The Linear team to monitor for new comments (leave empty to listen to all teams)

## Event Data

Each comment event includes the comment data with the parent issue reference.`
}

func (t *OnCommentCreated) Icon() string {
	return "linear"
}

func (t *OnCommentCreated) Color() string {
	return "blue"
}



func (t *OnCommentCreated) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "teamId",
			Label:       "Team",
			Type:        configuration.FieldTypeIntegrationResource,
			Required:    false,
			Description: "The Linear team to monitor for new comments (leave empty to listen to all teams)",
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type: "team",
				},
			},
		},
	}
}

func (t *OnCommentCreated) Setup(ctx core.TriggerContext) error {
	return nil
}

func (t *OnCommentCreated) Hooks() []core.Hook {
	return []core.Hook{}
}

func (t *OnCommentCreated) HandleHook(ctx core.TriggerHookContext) (map[string]any, error) {
	return nil, nil
}

func (t *OnCommentCreated) HandleWebhook(ctx core.WebhookRequestContext) (int, *core.WebhookResponseBody, error) {
	var config OnCommentCreatedConfiguration
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
	// Only handle comment create events
	//
	if payload.Type != "Comment" || payload.Action != "create" {
		ctx.Logger.Infof("Ignoring event - type %q action %q is not a comment creation", payload.Type, payload.Action)
		return http.StatusOK, nil, nil
	}

	//
	// If team filter is set, only process events for the specified team
	//
	if config.TeamID != "" {
		issueData, ok := payload.Data["issue"]
		if !ok {
			ctx.Logger.Infof("Ignoring event - no issue information in payload")
			return http.StatusOK, nil, nil
		}
		issueMap, ok := issueData.(map[string]any)
		if !ok {
			ctx.Logger.Infof("Ignoring event - invalid issue data in payload")
			return http.StatusOK, nil, nil
		}
		teamData, ok := issueMap["team"]
		if !ok {
			ctx.Logger.Infof("Ignoring event - no team information in issue")
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

	if err := ctx.Events.Emit("linear.comment.created", payload.Data); err != nil {
		ctx.Logger.Errorf("Failed to emit event: %v", err)
		return http.StatusInternalServerError, nil, fmt.Errorf("error emitting event: %v", err)
	}

	return http.StatusOK, nil, nil
}

func (t *OnCommentCreated) Cleanup(ctx core.TriggerContext) error {
	return nil
}
