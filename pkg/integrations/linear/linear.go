package linear

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/pkg/registry"
)

func init() {
	registry.RegisterIntegrationWithWebhookHandler("linear", &Linear{}, &LinearWebhookHandler{})
}

type Linear struct{}

type Metadata struct {
	TeamID   string `json:"teamId,omitempty" mapstructure:"teamId,omitempty"`
	TeamName string `json:"teamName,omitempty" mapstructure:"teamName,omitempty"`
	TeamKey  string `json:"teamKey,omitempty" mapstructure:"teamKey,omitempty"`
	UserID   string `json:"userId,omitempty" mapstructure:"userId,omitempty"`
	UserName string `json:"userName,omitempty" mapstructure:"userName,omitempty"`
}

const installationInstructions = `
To connect Linear to SuperPlane:

1. Go to your Linear workspace settings and navigate to **API**.
2. Click **Create API Key** and copy the generated key.
3. Paste the **API Key** into SuperPlane.
4. Configure a webhook URL in Linear settings pointing to the SuperPlane webhook endpoint.
`

func (l *Linear) Name() string {
	return "linear"
}

func (l *Linear) Label() string {
	return "Linear"
}

func (l *Linear) Icon() string {
	return "linear"
}

func (l *Linear) Description() string {
	return "Manage and react to changes in your Linear workspace"
}

func (l *Linear) Instructions() string {
	return installationInstructions
}

func (l *Linear) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "apiKey",
			Label:       "API Key",
			Type:        configuration.FieldTypeString,
			Sensitive:   true,
			Required:    true,
			Description: "Personal API key from Linear workspace settings",
		},
	}
}

func (l *Linear) Actions() []core.Action {
	return []core.Action{
		&CreateIssue{},
		&UpdateIssue{},
		&AddComment{},
	}
}

func (l *Linear) Triggers() []core.Trigger {
	return []core.Trigger{
		&OnIssueStateChange{},
		&OnIssueLabelChange{},
		&OnCycleChange{},
		&OnCommentCreated{},
	}
}

func (l *Linear) Cleanup(ctx core.IntegrationCleanupContext) error {
	return nil
}

func (l *Linear) Hooks() []core.Hook {
	return []core.Hook{}
}

func (l *Linear) HandleHook(ctx core.IntegrationHookContext) error {
	return nil
}

func (l *Linear) Sync(ctx core.SyncContext) error {
	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return fmt.Errorf("error creating Linear client: %v", err)
	}

	viewer, err := client.GetViewer()
	if err != nil {
		return fmt.Errorf("error verifying Linear credentials: %v", err)
	}

	teams, err := client.ListTeams()
	if err != nil {
		return fmt.Errorf("error listing Linear teams: %v", err)
	}

	metadata := Metadata{
		UserID:   viewer.ID,
		UserName: viewer.Name,
	}

	if len(teams) > 0 {
		metadata.TeamID = teams[0].ID
		metadata.TeamName = teams[0].Name
		metadata.TeamKey = teams[0].Key
	}

	ctx.Integration.SetMetadata(metadata)
	ctx.Integration.Ready()
	return nil
}

func (l *Linear) HandleRequest(ctx core.HTTPRequestContext) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Logger.Errorf("error reading request body: %v", err)
		ctx.Response.WriteHeader(http.StatusBadRequest)
		return
	}

	//
	// Validate the webhook signature
	//
	signature := ctx.Request.Header.Get("X-Linear-Signature")
	if signature == "" {
		ctx.Logger.Errorf("missing X-Linear-Signature header")
		ctx.Response.WriteHeader(http.StatusForbidden)
		return
	}

	secret, err := ctx.Integration.GetConfig("apiKey")
	if err != nil || len(secret) == 0 {
		ctx.Logger.Errorf("error getting API key for signature verification")
		ctx.Response.WriteHeader(http.StatusInternalServerError)
		return
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		ctx.Logger.Errorf("invalid Linear webhook signature")
		ctx.Response.WriteHeader(http.StatusForbidden)
		return
	}

	var payload LinearWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		ctx.Logger.Errorf("error unmarshaling Linear webhook payload: %v", err)
		ctx.Response.WriteHeader(http.StatusBadRequest)
		return
	}

	subscriptions, err := ctx.Integration.ListSubscriptions()
	if err != nil {
		ctx.Logger.Errorf("error listing subscriptions: %v", err)
		ctx.Response.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, subscription := range subscriptions {
		err = subscription.SendMessage(payload.Data)
		if err != nil {
			ctx.Logger.Errorf("error sending message from Linear: %v", err)
		}
	}

	ctx.Response.WriteHeader(http.StatusOK)
}

// LinearWebhookPayload represents the payload sent by Linear webhooks.
type LinearWebhookPayload struct {
	Action    string         `json:"action"`
	Type      string         `json:"type"`
	CreatedAt string         `json:"createdAt"`
	UpdatedBy string         `json:"updatedBy"`
	Data      map[string]any `json:"data"`
}
