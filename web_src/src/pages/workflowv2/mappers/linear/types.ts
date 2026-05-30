export interface LinearIssue {
  id: string;
  identifier: string;
  title: string;
  description: string;
  state: { id: string; name: string; type: string };
  priority: number;
  labels: { nodes: Array<{ id: string; name: string; color: string }> };
  assignee: { id: string; name: string };
  team: { id: string; name: string; key: string };
  url: string;
}

export interface LinearIssueConfiguration {
  title?: string;
  description?: string;
  teamId?: string;
  issueId?: string;
  body?: string;
}

export interface LinearCycleConfiguration {
  cycleId?: string;
  teamId?: string;
}

export interface LinearStateChangeConfiguration {
  teamId?: string;
  stateTypes?: string[];
}

export interface LinearLabelChangeConfiguration {
  teamId?: string;
  labelNames?: string[];
}

export interface LinearCommentConfiguration {
  issueId?: string;
  body?: string;
}

export interface LinearStateChangeEventData {
  issue?: LinearIssue;
  from_state?: { id: string; name: string; type: string };
  to_state?: { id: string; name: string; type: string };
}

export interface LinearLabelChangeEventData {
  issue?: LinearIssue;
  added_labels?: Array<{ id: string; name: string; color: string }>;
  removed_labels?: Array<{ id: string; name: string; color: string }>;
}

export interface LinearCycleChangeEventData {
  cycle?: { id: string; name: string; number: number };
  team?: { id: string; name: string; key: string };
  issues?: Array<{ id: string; identifier: string; title: string }>;
}

export interface LinearCommentEventData {
  comment?: { id: string; body: string; user?: { id: string; name: string } };
  issue?: LinearIssue;
}