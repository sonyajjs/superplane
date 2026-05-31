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

func Test__Render_GetLogs__Setup(t *testing.T) {
	component := &GetLogs{}

	t.Run("missing service -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{Configuration: map[string]any{}})
		require.ErrorContains(t, err, "service is required")
	})

	t.Run("valid configuration -> success", func(t *testing.T) {
		err := component.Setup(core.SetupContext{Configuration: map[string]any{"service": "srv-123"}})
		require.NoError(t, err)
	})

	t.Run("default lines -> success", func(t *testing.T) {
		err := component.Setup(core.SetupContext{Configuration: map[string]any{"service": "srv-123", "lines": 0}})
		require.NoError(t, err)
	})
}

func Test__Render_GetLogs__Execute(t *testing.T) {
	component := &GetLogs{}

	t.Run("successful log fetch -> emits payload", func(t *testing.T) {
		httpCtx := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(
						`[{"timestamp":"2025-01-01T00:00:00Z","message":"server started"},{"timestamp":"2025-01-01T00:01:00Z","message":"request received"}]`,
					)),
				},
			},
		}

		executionState := &contexts.ExecutionStateContext{KVs: map[string]string{}}

		err := component.Execute(core.ExecutionContext{
			HTTP:           httpCtx,
			Integration:    &contexts.IntegrationContext{Configuration: map[string]any{"apiKey": "rnd_test"}},
			ExecutionState: executionState,
			Configuration:  map[string]any{"service": "srv-123", "lines": 10},
		})

		require.NoError(t, err)

		assert.Equal(t, core.DefaultOutputChannel.Name, executionState.Channel)
		assert.Equal(t, GetLogsPayloadType, executionState.Type)
		require.Len(t, executionState.Payloads, 1)

		emittedPayload := readMap(executionState.Payloads[0])
		data := readMap(emittedPayload["data"])
		assert.Equal(t, "srv-123", data["serviceId"])
		assert.Equal(t, float64(2), data["count"]) // JSON numbers are float64

		lines, ok := data["lines"].([]any)
		require.True(t, ok)
		require.Len(t, lines, 2)

		line1 := readMap(lines[0])
		assert.Equal(t, "2025-01-01T00:00:00Z", line1["timestamp"])
		assert.Equal(t, "server started", line1["message"])

		require.Len(t, httpCtx.Requests, 1)
		request := httpCtx.Requests[0]
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Contains(t, request.URL.Path, "/v1/services/srv-123/logs")
		assert.Equal(t, "10", request.URL.Query().Get("limit"))
	})

	t.Run("API error -> returns error", func(t *testing.T) {
		httpCtx := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader(`{"error": "internal error"}`)),
				},
			},
		}

		executionState := &contexts.ExecutionStateContext{KVs: map[string]string{}}

		err := component.Execute(core.ExecutionContext{
			HTTP:           httpCtx,
			Integration:    &contexts.IntegrationContext{Configuration: map[string]any{"apiKey": "rnd_test"}},
			ExecutionState: executionState,
			Configuration:  map[string]any{"service": "srv-123"},
		})

		require.Error(t, err)
	})
}
