package linear

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

const AddCommentPayloadType = "linear.comment"

type AddComment struct{}

type AddCommentSpec struct {
	IssueID string `json:"issueId" mapstructure:"issueId"`
	Body    string `json:"body" mapstructure:"body"`
}

func (a *AddComment) Name() string {
	return "linear.addComment"
}

func (a *AddComment) Label() string {
	return "Add Comment"
}

func (a *AddComment) Description() string {
	return "Add a comment to an issue in Linear"
}

func (a *AddComment) Documentation() string {
	return `The Add Comment component adds a comment to an existing issue in Linear.

## Use Cases

- **Auto-responses**: Automatically comment on issues based on workflow events
- **Status updates**: Post status updates as comments on issues
- **Cross-system sync**: Mirror comments from other systems to Linear

## Configuration

- **Issue ID**: The ID of the issue to comment on (required)
- **Body**: The comment text (required, supports expressions)

## Output

Returns the comment data including:
- **id**: The comment ID
- **body**: The comment text
- **issueId**: The issue ID`
}

func (a *AddComment) Icon() string {
	return "linear"
}

func (a *AddComment) Color() string {
	return "blue"
}



func (a *AddComment) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (a *AddComment) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "issueId",
			Label:       "Issue ID",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The ID of the issue to comment on",
		},
		{
			Name:        "body",
			Label:       "Comment",
			Type:        configuration.FieldTypeText,
			Required:    true,
			Description: "The comment text to add",
		},
	}
}

func (a *AddComment) Setup(ctx core.SetupContext) error {
	spec := AddCommentSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("failed to decode configuration: %v", err)
	}

	if spec.IssueID == "" {
		return fmt.Errorf("issueId is required")
	}

	if spec.Body == "" {
		return fmt.Errorf("body is required")
	}

	return nil
}

func (a *AddComment) Execute(ctx core.ExecutionContext) error {
	spec := AddCommentSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("failed to decode configuration: %v", err)
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return fmt.Errorf("failed to create Linear client: %v", err)
	}

	commentID, err := client.CreateComment(spec.IssueID, spec.Body)
	if err != nil {
		return fmt.Errorf("failed to add comment: %v", err)
	}

	result := map[string]any{
		"id":      commentID,
		"body":    spec.Body,
		"issueId": spec.IssueID,
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		AddCommentPayloadType,
		[]any{result},
	)
}

func (a *AddComment) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (a *AddComment) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (a *AddComment) HandleWebhook(ctx core.WebhookRequestContext) (int, *core.WebhookResponseBody, error) {
	return http.StatusOK, nil, nil
}

func (a *AddComment) Cleanup(ctx core.SetupContext) error {
	return nil
}

func (a *AddComment) Hooks() []core.Hook {
	return []core.Hook{}
}

func (a *AddComment) HandleHook(ctx core.ActionHookContext) error {
	return nil
}
