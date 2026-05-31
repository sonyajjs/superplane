package linear

import (
	_ "embed"
	"sync"

	"github.com/superplanehq/superplane/pkg/utils"
)

//go:embed payloads/example_data_on_issue_state_change.json
var exampleDataOnIssueStateChangeBytes []byte

//go:embed payloads/example_data_on_issue_label_change.json
var exampleDataOnIssueLabelChangeBytes []byte

//go:embed payloads/example_data_on_cycle_change.json
var exampleDataOnCycleChangeBytes []byte

//go:embed payloads/example_data_on_comment_created.json
var exampleDataOnCommentCreatedBytes []byte

//go:embed payloads/example_output_create_issue.json
var exampleOutputCreateIssueBytes []byte

//go:embed payloads/example_output_update_issue.json
var exampleOutputUpdateIssueBytes []byte

//go:embed payloads/example_output_add_comment.json
var exampleOutputAddCommentBytes []byte

var exampleDataOnIssueStateChangeOnce sync.Once
var exampleDataOnIssueStateChange map[string]any

var exampleDataOnIssueLabelChangeOnce sync.Once
var exampleDataOnIssueLabelChange map[string]any

var exampleDataOnCycleChangeOnce sync.Once
var exampleDataOnCycleChange map[string]any

var exampleDataOnCommentCreatedOnce sync.Once
var exampleDataOnCommentCreated map[string]any

var exampleOutputCreateIssueOnce sync.Once
var exampleOutputCreateIssue map[string]any

var exampleOutputUpdateIssueOnce sync.Once
var exampleOutputUpdateIssue map[string]any

var exampleOutputAddCommentOnce sync.Once
var exampleOutputAddComment map[string]any

func (t *OnIssueStateChange) ExampleData() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleDataOnIssueStateChangeOnce, exampleDataOnIssueStateChangeBytes, &exampleDataOnIssueStateChange)
}

func (t *OnIssueLabelChange) ExampleData() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleDataOnIssueLabelChangeOnce, exampleDataOnIssueLabelChangeBytes, &exampleDataOnIssueLabelChange)
}

func (t *OnCycleChange) ExampleData() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleDataOnCycleChangeOnce, exampleDataOnCycleChangeBytes, &exampleDataOnCycleChange)
}

func (t *OnCommentCreated) ExampleData() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleDataOnCommentCreatedOnce, exampleDataOnCommentCreatedBytes, &exampleDataOnCommentCreated)
}

func (c *CreateIssue) ExampleOutput() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleOutputCreateIssueOnce, exampleOutputCreateIssueBytes, &exampleOutputCreateIssue)
}

func (u *UpdateIssue) ExampleOutput() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleOutputUpdateIssueOnce, exampleOutputUpdateIssueBytes, &exampleOutputUpdateIssue)
}

func (a *AddComment) ExampleOutput() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleOutputAddCommentOnce, exampleOutputAddCommentBytes, &exampleOutputAddComment)
}