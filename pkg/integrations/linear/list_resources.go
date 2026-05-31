package linear

import (
	"fmt"

	"github.com/superplanehq/superplane/pkg/core"
)

func (l *Linear) ListResources(resourceType string, ctx core.ListResourcesContext) ([]core.IntegrationResource, error) {
	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return nil, fmt.Errorf("error creating Linear client: %v", err)
	}

	switch resourceType {
	case "team":
		teams, err := client.ListTeams()
		if err != nil {
			return nil, fmt.Errorf("error listing teams: %v", err)
		}

		resources := make([]core.IntegrationResource, len(teams))
		for i, team := range teams {
			resources[i] = core.IntegrationResource{
				ID:   team.ID,
				Name: team.Name,
			}
		}
		return resources, nil

	case "workflowState":
		teamID := ctx.Parameters["teamId"]
		if teamID == "" {
			return nil, fmt.Errorf("teamId parameter is required for workflowState resource type")
		}

		states, err := client.ListWorkflowStates(teamID)
		if err != nil {
			return nil, fmt.Errorf("error listing workflow states: %v", err)
		}

		resources := make([]core.IntegrationResource, len(states))
		for i, state := range states {
			resources[i] = core.IntegrationResource{
				ID:   state.ID,
				Name: fmt.Sprintf("%s (%s)", state.Name, state.Type),
			}
		}
		return resources, nil

	case "label":
		teamID := ctx.Parameters["teamId"]
		if teamID == "" {
			return nil, fmt.Errorf("teamId parameter is required for label resource type")
		}

		labels, err := client.ListLabels(teamID)
		if err != nil {
			return nil, fmt.Errorf("error listing labels: %v", err)
		}

		resources := make([]core.IntegrationResource, len(labels))
		for i, label := range labels {
			resources[i] = core.IntegrationResource{
				ID:   label.ID,
				Name: label.Name,
			}
		}
		return resources, nil

	default:
		return nil, fmt.Errorf("unknown resource type: %s", resourceType)
	}
}
