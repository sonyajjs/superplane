package linear

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superplanehq/superplane/pkg/core"
	contexts "github.com/superplanehq/superplane/test/support/contexts"
)

// ─── Integration Registration ───

func TestLinear_IntegrationRegistration(t *testing.T) {
	integration := &Linear{}

	assert.Equal(t, "linear", integration.Name())
	assert.Equal(t, "Linear", integration.Label())
	assert.Equal(t, "linear", integration.Icon())
	assert.NotEmpty(t, integration.Description())
	assert.NotEmpty(t, integration.Instructions())

	// Verify triggers are registered
	triggers := integration.Triggers()
	assert.Equal(t, 4, len(triggers), "should have 4 triggers")

	// Verify actions are registered
	actions := integration.Actions()
	assert.Equal(t, 3, len(actions), "should have 3 actions")

	// Check trigger names
	triggerNames := make(map[string]bool)
	for _, trigger := range triggers {
		triggerNames[trigger.Name()] = true
	}
	assert.True(t, triggerNames["linear.onIssueStateChange"])
	assert.True(t, triggerNames["linear.onIssueLabelChange"])
	assert.True(t, triggerNames["linear.onCycleChange"])
	assert.True(t, triggerNames["linear.onCommentCreated"])

	// Check action names
	actionNames := make(map[string]bool)
	for _, action := range actions {
		actionNames[action.Name()] = true
	}
	assert.True(t, actionNames["linear.createIssue"])
	assert.True(t, actionNames["linear.updateIssue"])
	assert.True(t, actionNames["linear.addComment"])
}

// ─── CreateIssue ───

func Test__CreateIssue__Setup(t *testing.T) {
	component := CreateIssue{}

	t.Run("valid configuration with required fields", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"teamId": "team-123", "title": "Bug fix"},
		})
		require.NoError(t, err)
	})

	t.Run("missing teamId returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"title": "Bug fix"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "teamId is required")
	})

	t.Run("missing title returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"teamId": "team-123"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("empty teamId returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"teamId": "", "title": "Bug fix"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "teamId is required")
	})

	t.Run("empty title returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"teamId": "team-123", "title": ""},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("invalid configuration format returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: "not a map",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode configuration")
	})
}

func Test__CreateIssue__Execute(t *testing.T) {
	component := CreateIssue{}

	t.Run("fails when configuration decode fails", func(t *testing.T) {
		err := component.Execute(core.ExecutionContext{
			Integration:    &contexts.IntegrationContext{},
			ExecutionState: &contexts.ExecutionStateContext{},
			Configuration:  "not a map",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode configuration")
	})

	t.Run("fails without API key", func(t *testing.T) {
		err := component.Execute(core.ExecutionContext{
			Integration:    &contexts.IntegrationContext{Configuration: map[string]any{}},
			ExecutionState: &contexts.ExecutionStateContext{},
			Configuration: map[string]any{
				"teamId": "team-123",
				"title":  "Bug fix",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API key")
	})
}

func Test__CreateIssue__Metadata(t *testing.T) {
	component := CreateIssue{}

	assert.Equal(t, "linear.createIssue", component.Name())
	assert.Equal(t, "Create Issue", component.Label())
	assert.Equal(t, "linear", component.Icon())
	assert.Equal(t, "blue", component.Color())
	assert.NotEmpty(t, component.Documentation())
	assert.NotNil(t, component.ExampleOutput())

	channels := component.OutputChannels(nil)
	assert.Equal(t, 1, len(channels))
	assert.Equal(t, "default", channels[0].Name)
}

// ─── UpdateIssue ───

func Test__UpdateIssue__Setup(t *testing.T) {
	component := UpdateIssue{}

	t.Run("valid configuration with issueId", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"issueId": "issue-123"},
		})
		require.NoError(t, err)
	})

	t.Run("missing issueId returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"title": "New title"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issueId is required")
	})

	t.Run("empty issueId returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"issueId": ""},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issueId is required")
	})

	t.Run("invalid configuration format returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: "not a map",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode configuration")
	})
}

func Test__UpdateIssue__Metadata(t *testing.T) {
	component := UpdateIssue{}

	assert.Equal(t, "linear.updateIssue", component.Name())
	assert.Equal(t, "Update Issue", component.Label())
	assert.Equal(t, "linear", component.Icon())
	assert.Equal(t, "blue", component.Color())
	assert.NotEmpty(t, component.Documentation())
	assert.NotNil(t, component.ExampleOutput())

	channels := component.OutputChannels(nil)
	assert.Equal(t, 1, len(channels))
	assert.Equal(t, "default", channels[0].Name)
}

// ─── AddComment ───

func Test__AddComment__Setup(t *testing.T) {
	component := AddComment{}

	t.Run("valid configuration", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"issueId": "issue-123", "body": "Great work!"},
		})
		require.NoError(t, err)
	})

	t.Run("missing issueId returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"body": "Great work!"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "issueId is required")
	})

	t.Run("missing body returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"issueId": "issue-123"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "body is required")
	})

	t.Run("empty body returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"issueId": "issue-123", "body": ""},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "body is required")
	})

	t.Run("invalid configuration format returns error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Integration:   &contexts.IntegrationContext{},
			Metadata:      &contexts.MetadataContext{},
			Configuration: "not a map",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode configuration")
	})
}

func Test__AddComment__Execute(t *testing.T) {
	component := AddComment{}

	t.Run("fails when configuration decode fails", func(t *testing.T) {
		err := component.Execute(core.ExecutionContext{
			Integration:    &contexts.IntegrationContext{},
			ExecutionState: &contexts.ExecutionStateContext{},
			Configuration:  "not a map",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode configuration")
	})

	t.Run("fails without API key", func(t *testing.T) {
		err := component.Execute(core.ExecutionContext{
			Integration:    &contexts.IntegrationContext{Configuration: map[string]any{}},
			ExecutionState: &contexts.ExecutionStateContext{},
			Configuration: map[string]any{
				"issueId": "issue-123",
				"body":    "Great work!",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API key")
	})
}

func Test__AddComment__Metadata(t *testing.T) {
	component := AddComment{}

	assert.Equal(t, "linear.addComment", component.Name())
	assert.Equal(t, "Add Comment", component.Label())
	assert.Equal(t, "linear", component.Icon())
	assert.Equal(t, "blue", component.Color())
	assert.NotEmpty(t, component.Documentation())
	assert.NotNil(t, component.ExampleOutput())

	channels := component.OutputChannels(nil)
	assert.Equal(t, 1, len(channels))
	assert.Equal(t, "default", channels[0].Name)
}

// ─── Triggers ───

func Test__OnIssueStateChange__Metadata(t *testing.T) {
	trigger := &OnIssueStateChange{}

	assert.Equal(t, "linear.onIssueStateChange", trigger.Name())
	assert.Equal(t, "On Issue State Change", trigger.Label())
	assert.Equal(t, "linear", trigger.Icon())
	assert.NotEmpty(t, trigger.Documentation())
	assert.NotNil(t, trigger.ExampleData())
}

func Test__OnIssueLabelChange__Metadata(t *testing.T) {
	trigger := &OnIssueLabelChange{}

	assert.Equal(t, "linear.onIssueLabelChange", trigger.Name())
	assert.Equal(t, "On Issue Label Change", trigger.Label())
	assert.Equal(t, "linear", trigger.Icon())
	assert.NotEmpty(t, trigger.Documentation())
	assert.NotNil(t, trigger.ExampleData())
}

func Test__OnCycleChange__Metadata(t *testing.T) {
	trigger := &OnCycleChange{}

	assert.Equal(t, "linear.onCycleChange", trigger.Name())
	assert.Equal(t, "On Cycle Change", trigger.Label())
	assert.Equal(t, "linear", trigger.Icon())
	assert.NotEmpty(t, trigger.Documentation())
	assert.NotNil(t, trigger.ExampleData())
}

func Test__OnCommentCreated__Metadata(t *testing.T) {
	trigger := &OnCommentCreated{}

	assert.Equal(t, "linear.onCommentCreated", trigger.Name())
	assert.Equal(t, "On Comment Created", trigger.Label())
	assert.Equal(t, "linear", trigger.Icon())
	assert.NotEmpty(t, trigger.Documentation())
	assert.NotNil(t, trigger.ExampleData())
}

// ─── Webhook Handler ───

func Test__LinearWebhookHandler__CompareConfig(t *testing.T) {
	handler := &LinearWebhookHandler{}

	t.Run("same event type returns true", func(t *testing.T) {
		configA := map[string]any{"eventType": "issue_state_change"}
		configB := map[string]any{"eventType": "issue_state_change"}
		result, err := handler.CompareConfig(configA, configB)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("different event type returns false", func(t *testing.T) {
		configA := map[string]any{"eventType": "issue_state_change"}
		configB := map[string]any{"eventType": "cycle_change"}
		result, err := handler.CompareConfig(configA, configB)
		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		configA := "not a map"
		configB := map[string]any{"eventType": "issue_state_change"}
		_, err := handler.CompareConfig(configA, configB)
		require.Error(t, err)
	})
}

// ─── Client ───

func Test__Client__NewClient(t *testing.T) {
	t.Run("fails without API key", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{},
		}
		_, err := NewClient(&contexts.HTTPContext{}, integrationCtx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "apiKey")
	})

	t.Run("fails with empty API key", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{"apiKey": ""},
		}
		_, err := NewClient(&contexts.HTTPContext{}, integrationCtx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Linear API key")
	})
}

// ─── ListResources ───

func Test__Linear__ListResources(t *testing.T) {
	integration := &Linear{}

	t.Run("unknown resource type returns error", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Configuration: map[string]any{"apiKey": "test-key"},
		}

		_, err := integration.ListResources("unknown", core.ListResourcesContext{
			Integration: integrationCtx,
			HTTP:        &contexts.HTTPContext{},
		})
		require.Error(t, err)
	})
}