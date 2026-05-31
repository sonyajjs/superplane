package linear

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

const CreateIssuePayloadType = "linear.issue"

type CreateIssue struct{}

type CreateIssueSpec struct {
	TeamID      string   `json:"teamId" mapstructure:"teamId"`
	Title       string   `json:"title" mapstructure:"title"`
	Description string   `json:"description" mapstructure:"description"`
	Priority    *int     `json:"priority,omitempty" mapstructure:"priority"`
	LabelIDs    []string `json:"labelIds,omitempty" mapstructure:"labelIds"`
}

func (c *CreateIssue) Name() string {
	return "linear.createIssue"
}

func (c *CreateIssue) Label() string {
	return "Create Issue"
}

func (c *CreateIssue) Description() string {
	return "Create a new issue in Linear"
}

func (c *CreateIssue) Documentation() string {
	return `The Create Issue component creates a new issue in Linear.

## Use Cases

- **Task creation**: Automatically create tasks from workflow events
- **Bug tracking**: Create bugs from error detection systems
- **Feature requests**: Generate feature request issues from external inputs

## Configuration

- **Team**: The Linear team to create the issue in
- **Title**: The issue title (required, supports expressions)
- **Description**: Optional description text
- **Priority**: Optional priority level (0 = No priority, 1 = Urgent, 2 = High, 3 = Medium, 4 = Low)

## Output

Returns the created issue including:
- **id**: The issue ID
- **identifier**: The issue identifier (e.g., TEAM-42)
- **title**: The issue title
- **stateName**: The current state name
- **teamKey**: The team key
- **url**: The issue URL in Linear`
}

func (c *CreateIssue) Icon() string {
	return "linear"
}

func (c *CreateIssue) Color() string {
	return "blue"
}



func (c *CreateIssue) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *CreateIssue) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "teamId",
			Label:       "Team",
			Type:        configuration.FieldTypeIntegrationResource,
			Required:    true,
			Description: "The Linear team to create the issue in",
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type: "team",
				},
			},
		},
		{
			Name:        "title",
			Label:       "Title",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The issue title",
		},
		{
			Name:        "description",
			Label:       "Description",
			Type:        configuration.FieldTypeString,
			Required:    false,
			Description: "Optional description text",
		},
		{
			Name:        "priority",
			Label:       "Priority",
			Type:        configuration.FieldTypeSelect,
			Required:    false,
			Description: "Priority level (0=No priority, 1=Urgent, 2=High, 3=Medium, 4=Low)",
			TypeOptions: &configuration.TypeOptions{
				Select: &configuration.SelectTypeOptions{
					Options: []configuration.FieldOption{
						{Label: "No priority", Value: "0"},
						{Label: "Urgent", Value: "1"},
						{Label: "High", Value: "2"},
						{Label: "Medium", Value: "3"},
						{Label: "Low", Value: "4"},
					},
				},
			},
		},
	}
}

func (c *CreateIssue) Setup(ctx core.SetupContext) error {
	spec := CreateIssueSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("failed to decode configuration: %v", err)
	}

	if spec.TeamID == "" {
		return fmt.Errorf("teamId is required")
	}

	if spec.Title == "" {
		return fmt.Errorf("title is required")
	}

	return nil
}

func (c *CreateIssue) Execute(ctx core.ExecutionContext) error {
	spec := CreateIssueSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("failed to decode configuration: %v", err)
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return fmt.Errorf("failed to create Linear client: %v", err)
	}

	input := CreateIssueInput{
		TeamID:      spec.TeamID,
		Title:       spec.Title,
		Description: spec.Description,
		Priority:    spec.Priority,
		LabelIDs:    spec.LabelIDs,
	}

	issue, err := client.CreateIssue(input)
	if err != nil {
		return fmt.Errorf("failed to create Linear issue: %v", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		CreateIssuePayloadType,
		[]any{issue},
	)
}

func (c *CreateIssue) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *CreateIssue) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *CreateIssue) HandleWebhook(ctx core.WebhookRequestContext) (int, *core.WebhookResponseBody, error) {
	return http.StatusOK, nil, nil
}

func (c *CreateIssue) Cleanup(ctx core.SetupContext) error {
	return nil
}

func (c *CreateIssue) Hooks() []core.Hook {
	return []core.Hook{}
}

func (c *CreateIssue) HandleHook(ctx core.ActionHookContext) error {
	return nil
}
