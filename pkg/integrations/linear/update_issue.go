package linear

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

const UpdateIssuePayloadType = "linear.issue.updated"

type UpdateIssue struct{}

type UpdateIssueSpec struct {
	IssueID     string   `json:"issueId" mapstructure:"issueId"`
	StateID     string   `json:"stateId,omitempty" mapstructure:"stateId"`
	Priority    *int     `json:"priority,omitempty" mapstructure:"priority"`
	Title       string   `json:"title,omitempty" mapstructure:"title"`
	Description string   `json:"description,omitempty" mapstructure:"description"`
	LabelIDs    []string `json:"labelIds,omitempty" mapstructure:"labelIds"`
}

func (u *UpdateIssue) Name() string {
	return "linear.updateIssue"
}

func (u *UpdateIssue) Label() string {
	return "Update Issue"
}

func (u *UpdateIssue) Description() string {
	return "Update an existing issue in Linear"
}

func (u *UpdateIssue) Documentation() string {
	return `The Update Issue component updates an existing issue in Linear.

## Use Cases

- **Status updates**: Move issues between workflow states
- **Priority changes**: Adjust issue priority based on external events
- **Label management**: Add or remove labels on issues
- **Description updates**: Update issue descriptions with new information

## Configuration

- **Issue ID**: The ID of the issue to update (required)
- **State**: The new workflow state for the issue
- **Priority**: The new priority level
- **Title**: Update the issue title
- **Description**: Update the issue description

## Output

Returns the updated issue.`
}

func (u *UpdateIssue) Icon() string {
	return "linear"
}

func (u *UpdateIssue) Color() string {
	return "blue"
}



func (u *UpdateIssue) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (u *UpdateIssue) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "issueId",
			Label:       "Issue ID",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Description: "The ID of the issue to update",
		},
		{
			Name:        "stateId",
			Label:       "State",
			Type:        configuration.FieldTypeIntegrationResource,
			Required:    false,
			Description: "The new workflow state for the issue",
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type: "workflowState",
					Parameters: []configuration.ParameterRef{
						{
							Name:      "teamId",
							ValueFrom: &configuration.ParameterValueFrom{Field: "teamId"},
						},
					},
				},
			},
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
		{
			Name:        "title",
			Label:       "Title",
			Type:        configuration.FieldTypeString,
			Required:    false,
			Description: "Update the issue title",
		},
		{
			Name:        "description",
			Label:       "Description",
			Type:        configuration.FieldTypeString,
			Required:    false,
			Description: "Update the issue description",
		},
	}
}

func (u *UpdateIssue) Setup(ctx core.SetupContext) error {
	spec := UpdateIssueSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("failed to decode configuration: %v", err)
	}

	if spec.IssueID == "" {
		return fmt.Errorf("issueId is required")
	}

	return nil
}

func (u *UpdateIssue) Execute(ctx core.ExecutionContext) error {
	spec := UpdateIssueSpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("failed to decode configuration: %v", err)
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return fmt.Errorf("failed to create Linear client: %v", err)
	}

	input := UpdateIssueInput{
		StateID:     spec.StateID,
		Priority:    spec.Priority,
		Title:       spec.Title,
		Description: spec.Description,
		LabelIDs:    spec.LabelIDs,
	}

	issue, err := client.UpdateIssue(spec.IssueID, input)
	if err != nil {
		return fmt.Errorf("failed to update Linear issue: %v", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		UpdateIssuePayloadType,
		[]any{issue},
	)
}

func (u *UpdateIssue) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (u *UpdateIssue) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (u *UpdateIssue) HandleWebhook(ctx core.WebhookRequestContext) (int, *core.WebhookResponseBody, error) {
	return http.StatusOK, nil, nil
}

func (u *UpdateIssue) Cleanup(ctx core.SetupContext) error {
	return nil
}

func (u *UpdateIssue) Hooks() []core.Hook {
	return []core.Hook{}
}

func (u *UpdateIssue) HandleHook(ctx core.ActionHookContext) error {
	return nil
}
