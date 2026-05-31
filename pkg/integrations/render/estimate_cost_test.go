package render

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/test/support/contexts"
)

func Test__Render_EstimateCost__Setup(t *testing.T) {
	component := &EstimateCost{}

	t.Run("empty configuration -> success (valid, all services)", func(t *testing.T) {
		err := component.Setup(core.SetupContext{Configuration: map[string]any{}})
		require.NoError(t, err)
	})

	t.Run("with service IDs -> success", func(t *testing.T) {
		err := component.Setup(core.SetupContext{Configuration: map[string]any{"serviceIds": "srv-1,srv-2"}})
		require.NoError(t, err)
	})
}

func Test__Render_EstimateCost__Execute(t *testing.T) {
	component := &EstimateCost{}

	t.Run("list services and estimate cost -> emits breakdown", func(t *testing.T) {
		httpCtx := &contexts.HTTPContext{
			Responses: []*http.Response{
				// First response: ListWorkspaces (for workspaceIDForIntegration)
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(
						`[{"owner":{"id":"tea-123","name":"my-team"},"cursor":""}]`,
					)),
				},
				// Second response: ListServices
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(
						`[{"service":{"id":"srv-1","name":"api","type":"web","plan":"standard","suspended":"not_suspended"},"cursor":""},{"service":{"id":"srv-2","name":"worker","type":"worker","plan":"starter","suspended":"not_suspended"},"cursor":""},{"service":{"id":"srv-3","name":"db","type":"pserv","plan":"starter","suspended":"not_suspended"},"cursor":""}]`,
					)),
				},
			},
		}

		executionState := &contexts.ExecutionStateContext{KVs: map[string]string{}}

		err := component.Execute(core.ExecutionContext{
			HTTP:           httpCtx,
			Integration:    &contexts.IntegrationContext{Configuration: map[string]any{"apiKey": "rnd_test"}},
			ExecutionState: executionState,
			Configuration:  map[string]any{},
		})

		require.NoError(t, err)

		assert.Equal(t, core.DefaultOutputChannel.Name, executionState.Channel)
		assert.Equal(t, EstimateCostPayloadType, executionState.Type)
		require.Len(t, executionState.Payloads, 1)

		emittedPayload := readMap(executionState.Payloads[0])
		result := readMap(emittedPayload["data"])

		assert.Equal(t, float64(3), result["serviceCount"])
		// standard web = 19, starter worker = 0, starter pserv = 0 => total = 19
		assert.Equal(t, float64(19), result["totalMonthlyUSD"])

		breakdown, ok := result["breakdown"].([]any)
		require.True(t, ok)
		require.Len(t, breakdown, 3)

		svc1 := readMap(breakdown[0])
		assert.Equal(t, "srv-1", svc1["serviceId"])
		assert.Equal(t, "api", svc1["serviceName"])
		assert.Equal(t, float64(19), svc1["monthlyUSD"])

		svc2 := readMap(breakdown[1])
		assert.Equal(t, "srv-2", svc2["serviceId"])
		assert.Equal(t, float64(0), svc2["monthlyUSD"])

		svc3 := readMap(breakdown[2])
		assert.Equal(t, "srv-3", svc3["serviceId"])
		assert.Equal(t, float64(0), svc3["monthlyUSD"])
	})

	t.Run("empty service list -> emits zero cost", func(t *testing.T) {
		httpCtx := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(
						`[{"owner":{"id":"tea-123","name":"my-team"},"cursor":""}]`,
					)),
				},
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`[]`)),
				},
			},
		}

		executionState := &contexts.ExecutionStateContext{KVs: map[string]string{}}

		err := component.Execute(core.ExecutionContext{
			HTTP:           httpCtx,
			Integration:    &contexts.IntegrationContext{Configuration: map[string]any{"apiKey": "rnd_test"}},
			ExecutionState: executionState,
			Configuration:  map[string]any{},
		})

		require.NoError(t, err)

		assert.Equal(t, EstimateCostPayloadType, executionState.Type)
		emittedPayload := readMap(executionState.Payloads[0])
		result := readMap(emittedPayload["data"])
		assert.Equal(t, float64(0), result["totalMonthlyUSD"])
		assert.Equal(t, float64(0), result["serviceCount"])
	})
}

func Test__Render_EstimateCost__LookupCost(t *testing.T) {
	tests := []struct {
		name     string
		plan     string
		svcType  string
		expected float64
	}{
		{"standard web", "standard", "web", 19},
		{"starter web", "starter", "web", 0},
		{"pro web", "pro", "web", 49},
		{"standard worker", "standard", "worker", 49},
		{"starter worker", "starter", "worker", 0},
		{"standard pserv", "standard", "pserv", 25},
		{"starter pserv", "starter", "pserv", 0},
		{"standard postgres", "standard", "postgres", 45},
		{"starter postgres", "starter", "postgres", 7},
		{"standard redis", "standard", "redis", 35},
		{"starter redis", "starter", "redis", 0},
		{"unknown plan", "enterprise", "web", 0},
		{"unknown type", "standard", "custom", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lookupCost(tt.plan, tt.svcType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
