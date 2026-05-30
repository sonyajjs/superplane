package linear

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/superplanehq/superplane/pkg/core"
)

const linearAPIURL = "https://api.linear.app/graphql"

// Client communicates with the Linear GraphQL API.
type Client struct {
	apiKey string
	http   core.HTTPContext
}

func NewClient(httpCtx core.HTTPContext, ctx core.IntegrationContext) (*Client, error) {
	apiKey, err := ctx.GetConfig("apiKey")
	if err != nil {
		return nil, fmt.Errorf("error reading API key: %v", err)
	}

	if len(apiKey) == 0 {
		return nil, fmt.Errorf("missing Linear API key")
	}

	return &Client{
		apiKey: string(apiKey),
		http:   httpCtx,
	}, nil
}

func (c *Client) doGraphQL(query string, variables map[string]any) ([]byte, error) {
	payload := map[string]any{
		"query":     query,
		"variables": variables,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling GraphQL request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, linearAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.apiKey)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %v", err)
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("Linear API returned %d: %s", res.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

// Viewer represents the authenticated user.
type Viewer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetViewer fetches the currently authenticated user.
func (c *Client) GetViewer() (*Viewer, error) {
	query := `query { viewer { id name } }`
	data, err := c.doGraphQL(query, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Viewer Viewer `json:"viewer"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing viewer response: %v", err)
	}

	return &result.Data.Viewer, nil
}

// Team represents a Linear team.
type Team struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// ListTeams fetches all teams in the workspace.
func (c *Client) ListTeams() ([]Team, error) {
	query := `query { teams { nodes { id key name } } }`
	data, err := c.doGraphQL(query, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Teams struct {
				Nodes []Team `json:"nodes"`
			} `json:"teams"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing teams response: %v", err)
	}

	return result.Data.Teams.Nodes, nil
}

// LinearIssue represents a Linear issue.
type LinearIssue struct {
	ID          string `json:"id"`
	Identifier  string `json:"identifier"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	StateName   string `json:"stateName"`
	TeamKey     string `json:"teamKey"`
	Priority    int    `json:"priority"`
	URL         string `json:"url"`
}

// CreateIssueInput holds the input for creating an issue.
type CreateIssueInput struct {
	TeamID      string   `json:"teamId"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	LabelIDs    []string `json:"labelIds,omitempty"`
}

// CreateIssue creates a new issue in Linear.
func (c *Client) CreateIssue(input CreateIssueInput) (*LinearIssue, error) {
	query := `
mutation CreateIssue($input: IssueCreateInput!) {
	issueCreate(input: $input) {
		success
		issue {
			id
			identifier
			title
			description
			state { name }
			team { key }
			priority
			url
		}
	}
}`

	variables := map[string]any{"input": input}
	data, err := c.doGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			IssueCreate struct {
				Success bool `json:"success"`
				Issue   struct {
					ID          string `json:"id"`
					Identifier  string `json:"identifier"`
					Title       string `json:"title"`
					Description string `json:"description"`
					State       struct {
						Name string `json:"name"`
					} `json:"state"`
					Team struct {
						Key string `json:"key"`
					} `json:"team"`
					Priority int    `json:"priority"`
					URL      string `json:"url"`
				} `json:"issue"`
			} `json:"issueCreate"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing create issue response: %v", err)
	}

	if !result.Data.IssueCreate.Success {
		return nil, fmt.Errorf("Linear issue creation was not successful")
	}

	issue := &LinearIssue{
		ID:          result.Data.IssueCreate.Issue.ID,
		Identifier:  result.Data.IssueCreate.Issue.Identifier,
		Title:       result.Data.IssueCreate.Issue.Title,
		Description: result.Data.IssueCreate.Issue.Description,
		StateName:   result.Data.IssueCreate.Issue.State.Name,
		TeamKey:     result.Data.IssueCreate.Issue.Team.Key,
		Priority:    result.Data.IssueCreate.Issue.Priority,
		URL:         result.Data.IssueCreate.Issue.URL,
	}

	return issue, nil
}

// UpdateIssueInput holds the input for updating an issue.
type UpdateIssueInput struct {
	StateID     string   `json:"stateId,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	LabelIDs    []string `json:"labelIds,omitempty"`
}

// UpdateIssue updates an existing issue in Linear.
func (c *Client) UpdateIssue(issueID string, input UpdateIssueInput) (*LinearIssue, error) {
	query := `
mutation UpdateIssue($id: String!, $input: IssueUpdateInput!) {
	issueUpdate(id: $id, input: $input) {
		success
		issue {
			id
			identifier
			title
			description
			state { name }
			team { key }
			priority
			url
		}
	}
}`

	variables := map[string]any{
		"id":    issueID,
		"input": input,
	}
	data, err := c.doGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			IssueUpdate struct {
				Success bool `json:"success"`
				Issue   struct {
					ID          string `json:"id"`
					Identifier  string `json:"identifier"`
					Title       string `json:"title"`
					Description string `json:"description"`
					State       struct {
						Name string `json:"name"`
					} `json:"state"`
					Team struct {
						Key string `json:"key"`
					} `json:"team"`
					Priority int    `json:"priority"`
					URL      string `json:"url"`
				} `json:"issue"`
			} `json:"issueUpdate"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing update issue response: %v", err)
	}

	if !result.Data.IssueUpdate.Success {
		return nil, fmt.Errorf("Linear issue update was not successful")
	}

	issue := &LinearIssue{
		ID:          result.Data.IssueUpdate.Issue.ID,
		Identifier:  result.Data.IssueUpdate.Issue.Identifier,
		Title:       result.Data.IssueUpdate.Issue.Title,
		Description: result.Data.IssueUpdate.Issue.Description,
		StateName:   result.Data.IssueUpdate.Issue.State.Name,
		TeamKey:     result.Data.IssueUpdate.Issue.Team.Key,
		Priority:    result.Data.IssueUpdate.Issue.Priority,
		URL:         result.Data.IssueUpdate.Issue.URL,
	}

	return issue, nil
}

// CreateComment creates a comment on an issue.
func (c *Client) CreateComment(issueID, body string) (string, error) {
	query := `
mutation CreateComment($input: CommentCreateInput!) {
	commentCreate(input: $input) {
		success
		comment {
			id
		}
	}
}`

	variables := map[string]any{
		"input": map[string]any{
			"issueId": issueID,
			"body":    body,
		},
	}
	data, err := c.doGraphQL(query, variables)
	if err != nil {
		return "", err
	}

	var result struct {
		Data struct {
			CommentCreate struct {
				Success bool `json:"success"`
				Comment struct {
					ID string `json:"id"`
				} `json:"comment"`
			} `json:"commentCreate"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("error parsing create comment response: %v", err)
	}

	if !result.Data.CommentCreate.Success {
		return "", fmt.Errorf("Linear comment creation was not successful")
	}

	return result.Data.CommentCreate.Comment.ID, nil
}

// ListWorkflowStates lists all workflow states for a team.
func (c *Client) ListWorkflowStates(teamID string) ([]WorkflowState, error) {
	query := `
query ListStates($teamId: String!) {
	team(id: $teamId) {
		states {
			nodes {
				id
				name
				type
			}
		}
	}
}`

	variables := map[string]any{"teamId": teamID}
	data, err := c.doGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Team struct {
				States struct {
					Nodes []WorkflowState `json:"nodes"`
				} `json:"states"`
			} `json:"team"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing workflow states response: %v", err)
	}

	return result.Data.Team.States.Nodes, nil
}

// WorkflowState represents a Linear workflow state.
type WorkflowState struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ListLabels lists all labels for a team.
func (c *Client) ListLabels(teamID string) ([]Label, error) {
	query := `
query ListLabels($teamId: String!) {
	team(id: $teamId) {
		labels {
			nodes {
				id
				name
				color
			}
		}
	}
}`

	variables := map[string]any{"teamId": teamID}
	data, err := c.doGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Team struct {
				Labels struct {
					Nodes []Label `json:"nodes"`
				} `json:"labels"`
			} `json:"team"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing labels response: %v", err)
	}

	return result.Data.Team.Labels.Nodes, nil
}

// Label represents a Linear label.
type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}
