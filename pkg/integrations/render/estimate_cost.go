package render

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

const EstimateCostPayloadType = "render.cost.estimate"

type EstimateCost struct{}

type EstimateCostConfiguration struct {
	ServiceIDs []string `json:"serviceIds" mapstructure:"serviceIds"`
}

type PlanPricing struct {
	Plan       string  `json:"plan"`
	Type       string  `json:"type"`
	MonthlyUSD float64 `json:"monthlyUSD"`
}

// defaultPricing maps Render plan+type to estimated monthly USD cost.
// Prices are approximate and based on Render's public pricing.
var defaultPricing = []PlanPricing{
	{Plan: "starter", Type: "web", MonthlyUSD: 0},
	{Plan: "standard", Type: "web", MonthlyUSD: 19},
	{Plan: "pro", Type: "web", MonthlyUSD: 49},
	{Plan: "starter", Type: "pserv", MonthlyUSD: 0},
	{Plan: "standard", Type: "pserv", MonthlyUSD: 25},
	{Plan: "pro", Type: "pserv", MonthlyUSD: 75},
	{Plan: "starter", Type: "worker", MonthlyUSD: 0},
	{Plan: "standard", Type: "worker", MonthlyUSD: 49},
	{Plan: "pro", Type: "worker", MonthlyUSD: 99},
	{Plan: "starter", Type: "cron", MonthlyUSD: 0},
	{Plan: "standard", Type: "cron", MonthlyUSD: 19},
	{Plan: "starter", Type: "postgres", MonthlyUSD: 7},
	{Plan: "standard", Type: "postgres", MonthlyUSD: 45},
	{Plan: "pro", Type: "postgres", MonthlyUSD: 99},
	{Plan: "starter", Type: "redis", MonthlyUSD: 0},
	{Plan: "standard", Type: "redis", MonthlyUSD: 35},
}

type serviceCostBreakdown struct {
	ServiceID      string  `json:"serviceId"`
	ServiceName    string  `json:"serviceName"`
	Type           string  `json:"type"`
	Plan           string  `json:"plan"`
	MonthlyUSD     float64 `json:"monthlyUSD"`
	Suspended      bool    `json:"suspended"`
}

type costEstimateResult struct {
	Breakdown     []serviceCostBreakdown `json:"breakdown"`
	TotalMonthly  float64                `json:"totalMonthlyUSD"`
	ServiceCount  int                    `json:"serviceCount"`
}

func (c *EstimateCost) Name() string {
	return "render.estimateCost"
}

func (c *EstimateCost) Label() string {
	return "Estimate Cost"
}

func (c *EstimateCost) Description() string {
	return "Estimate monthly cost for Render services based on their plans"
}

func (c *EstimateCost) Documentation() string {
	return `The Estimate Cost component calculates the projected monthly cost for Render services.

## Use Cases

- **Budget monitoring**: Check projected spend against a budget threshold
- **Cost optimization**: Identify expensive services that could be downgraded
- **FinOps reporting**: Generate cost reports for stakeholders

## Configuration

- **Service IDs**: List of service IDs to include. Empty = all services.

## Output

Emits a ` + "`render.cost.estimate`" + ` payload containing:
- Per-service breakdown (name, type, plan, monthly cost)
- Total projected monthly cost
- Total service count

## Notes

Prices are approximate based on Render's public pricing. Suspended services
still incur costs if their plan is not free.`
}

func (c *EstimateCost) Icon() string {
	return "dollar-sign"
}

func (c *EstimateCost) Color() string {
	return "green"
}

func (c *EstimateCost) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *EstimateCost) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "serviceIds",
			Label:       "Service IDs",
			Type:        configuration.FieldTypeText,
			Required:    false,
			Description: "Comma-separated list of service IDs to estimate. Leave empty for all services.",
		},
	}
}

func decodeEstimateCostConfiguration(configuration any) (EstimateCostConfiguration, error) {
	spec := EstimateCostConfiguration{}
	if err := mapstructure.Decode(configuration, &spec); err != nil {
		return EstimateCostConfiguration{}, fmt.Errorf("failed to decode configuration: %w", err)
	}

	return spec, nil
}

func (c *EstimateCost) Setup(ctx core.SetupContext) error {
	_, err := decodeEstimateCostConfiguration(ctx.Configuration)
	return err
}

func (c *EstimateCost) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *EstimateCost) Execute(ctx core.ExecutionContext) error {
	spec, err := decodeEstimateCostConfiguration(ctx.Configuration)
	if err != nil {
		return err
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return err
	}

	workspaceID, err := workspaceIDForIntegration(client, ctx.Integration)
	if err != nil {
		return err
	}

	services, err := client.ListServices(workspaceID)
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	filterIDs := make(map[string]bool)
	for _, id := range spec.ServiceIds {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			filterIDs[trimmed] = true
		}
	}

	breakdown := make([]serviceCostBreakdown, 0, len(services))
	var totalMonthly float64

	for _, svc := range services {
		if len(filterIDs) > 0 && !filterIDs[svc.ID] {
			continue
		}

		plan := planFromService(svc)
		svcType := serviceTypeFromService(svc)
		monthlyCost := lookupCost(plan, svcType)

		suspended := strings.EqualFold(svc.Suspended, "suspended")

		breakdown = append(breakdown, serviceCostBreakdown{
			ServiceID:   svc.ID,
			ServiceName: svc.Name,
			Type:        svcType,
			Plan:        plan,
			MonthlyUSD:  monthlyCost,
			Suspended:   suspended,
		})
		totalMonthly += monthlyCost
	}

	result := costEstimateResult{
		Breakdown:    breakdown,
		TotalMonthly: totalMonthly,
		ServiceCount: len(breakdown),
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		EstimateCostPayloadType,
		[]any{result},
	)
}

func planFromService(svc Service) string {
	// Extract plan from serviceDetails or use type as heuristic
	if svc.ServiceDetails != nil {
		if plan, ok := svc.ServiceDetails["plan"].(string); ok && plan != "" {
			return strings.ToLower(strings.TrimSpace(plan))
		}
	}
	// Default: assume starter if no plan info
	return "starter"
}

func serviceTypeFromService(svc Service) string {
	if svc.Type != "" {
		return strings.ToLower(strings.TrimSpace(svc.Type))
	}
	return "web"
}

func lookupCost(plan, svcType string) float64 {
	for _, pricing := range defaultPricing {
		if strings.EqualFold(pricing.Plan, plan) && strings.EqualFold(pricing.Type, svcType) {
			return pricing.MonthlyUSD
		}
	}
	// Unknown plan/type: assume free (conservative)
	return 0
}

func (c *EstimateCost) HandleWebhook(ctx core.WebhookRequestContext) (int, *core.WebhookResponseBody, error) {
	return http.StatusOK, nil, nil
}

func (c *EstimateCost) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *EstimateCost) Cleanup(ctx core.SetupContext) error {
	return nil
}

func (c *EstimateCost) Hooks() []core.Hook {
	return []core.Hook{}
}

func (c *EstimateCost) HandleHook(ctx core.ActionHookContext) error {
	return nil
}
