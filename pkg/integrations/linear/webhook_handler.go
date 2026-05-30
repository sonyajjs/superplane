package linear

import (
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/core"
)

type LinearWebhookHandler struct{}

func (h *LinearWebhookHandler) CompareConfig(a, b any) (bool, error) {
	configA := WebhookConfiguration{}
	configB := WebhookConfiguration{}

	if err := mapstructure.Decode(a, &configA); err != nil {
		return false, err
	}

	if err := mapstructure.Decode(b, &configB); err != nil {
		return false, err
	}

	return configA.EventType == configB.EventType, nil
}

func (h *LinearWebhookHandler) Merge(current, requested any) (any, bool, error) {
	return current, false, nil
}

func (h *LinearWebhookHandler) Setup(ctx core.WebhookHandlerContext) (any, error) {
	return nil, nil
}

func (h *LinearWebhookHandler) Cleanup(ctx core.WebhookHandlerContext) error {
	//
	// Linear webhooks are managed manually by the user in the Linear UI.
	// There is no API to programmatically create/delete webhooks at this time.
	//
	return nil
}

// WebhookConfiguration holds the configuration for a Linear webhook subscription.
type WebhookConfiguration struct {
	EventType string `json:"eventType" mapstructure:"eventType"`
}
