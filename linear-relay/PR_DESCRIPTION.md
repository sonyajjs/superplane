# feat: Add Linear project management integration with 4 triggers, 3 actions, and webhook relay

## Summary

This PR adds a full-featured **Linear** project management integration to SuperPlane, enabling teams to automate workflows that connect Linear issues, cycles, labels, and comments with any other integration in the SuperPlane ecosystem.

The integration consists of two complementary parts:

1. **In-platform integration** (`pkg/integrations/linear/`) — 4 triggers, 3 actions, a GraphQL API client, a webhook handler, and resource listing — all wired into the SuperPlane component registry and the visual canvas.
2. **Standalone webhook relay** (`linear-relay/`) — a lightweight Go microservice deployed on Render that receives Linear webhooks, validates HMAC-SHA256 signatures, deduplicates events via PostgreSQL, transforms payloads, and forwards them to SuperPlane.

Together, these enable real-time, bidirectional automation between Linear and the rest of your stack.

---

## Triggers (4)

| # | Component ID | Label | Description |
|---|---|---|---|
| 1 | `linear.onIssueStateChange` | On Issue State Change | Fires when a Linear issue moves between states (e.g. Backlog → In Progress → Done). Ideal for triggering deployments, notifications, or status updates. |
| 2 | `linear.onIssueLabelChange` | On Issue Label Change | Fires when labels are added or removed from a Linear issue. Use this to route issues by category, escalate high-priority bugs, or filter infra-change issues. |
| 3 | `linear.onCycleChange` | On Cycle Change | Fires when a Linear cycle is created, updated, or completed. Use this to gate releases, sync sprint boards, or send cycle summaries. |
| 4 | `linear.onCommentCreated` | On Comment Created | Fires when a comment is added to a Linear issue. Use this to trigger review reminders, sync discussion threads, or surface blockers. |

All triggers are registered with the SuperPlane webhook handler and route incoming Linear webhook events to the appropriate subscribed canvas nodes.

---

## Actions (3)

| # | Component ID | Label | Description |
|---|---|---|---|
| 1 | `linear.createIssue` | Create Issue | Creates a new Linear issue in a specified team. Supports title, description, priority, and label assignment. Useful for incident escalation, syncing bugs from external tools, or auto-creating sprint tasks. |
| 2 | `linear.updateIssue` | Update Issue | Updates an existing Linear issue's state, labels, assignee, priority, or other fields. Ideal for automating state transitions when PRs are opened, CI passes, or deployments succeed. |
| 3 | `linear.addComment` | Add Comment | Adds a comment to an existing Linear issue. Useful for posting deployment links, CI results, or cross-referencing external events directly in the issue thread. |

Each action validates its required configuration fields during setup and executes against the Linear GraphQL API using the stored API key.

---

## Canvas Demo Flows (3)

The `templates/canvases/linear-pr-sync.yaml` file ships three ready-to-use canvas flows:

### Flow 1: PR Opened → Move Issue to In Review + Deploy Preview

```
GitHub: PR Opened
  ├─→ Linear: Move Issue to In Review → Slack: Notify Team Review Ready
  └─→ Render: Set PREVIEW_BRANCH on Web Service → Render: Deploy Preview
```

When a developer opens a pull request, SuperPlane automatically:
- Extracts the Linear issue key (e.g. `TEAM-123`) from the PR body
- Moves the corresponding Linear issue to **In Review**
- Notifies the team in Slack
- Sets the preview branch on the Render Web Service and triggers a preview deploy

### Flow 2: Issue Done (infra-change) → Deploy to Production

```
Linear: Issue Moved to Done → Filter: Has infra-change Label
  → Render: Update RELEASE_VERSION on Web Service → Render: Deploy to Production → Slack: Notify Deployed
```

When an issue labeled `infra-change` reaches **Done**, SuperPlane:
- Filters on the `infra-change` label
- Updates the release version environment variable on the Render Web Service
- Deploys to production
- Posts a deployment notification to Slack

### Flow 3: PagerDuty Incident → Create Linear Issue

```
PagerDuty: Incident Triggered → Linear: Create Incident Issue → Slack: Notify Incident Issue Created
```

When PagerDuty fires a high-urgency incident, SuperPlane:
- Creates a high-priority Linear issue in the appropriate team
- Includes the incident title, severity, and PagerDuty URL
- Posts a Slack notification with the new issue link

---

## Webhook Relay Architecture

The standalone `linear-relay` microservice provides a secure, durable bridge between Linear's outbound webhooks and SuperPlane's inbound webhook processor.

```
┌────────────┐         ┌──────────────────────────┐         ┌──────────────┐
│   Linear   │  HTTPS  │   superplane-linear-relay │  HTTPS  │  SuperPlane  │
│  Workspace │ ──────► │   (Render Web Service)    │ ──────► │  Webhook     │
│            │  POST   │                            │  POST   │  Processor   │
└────────────┘         │  ┌──────────────────────┐  │         └──────────────┘
                       │  │ /webhook              │  │
                       │  │  ┌─ HMAC validation   │  │
                       │  │  ├─ Event ID dedup   │  │
                       │  │  ├─ Payload transform│  │
                       │  │  └─ Forward to SP    │  │
                       │  └──────────────────────┘  │
                       │  ┌──────────────────────┐  │
                       │  │ /health               │  │
                       │  │  └─ DB connectivity   │  │
                       │  └──────────────────────┘  │
                       │                            │
                       │  ┌──────────────────────┐  │
                       │  │ PostgreSQL (Render)   │  │
                       │  │  webhook_events table │  │
                       │  └──────────────────────┘  │
                       └──────────────────────────┘
```

**Key features:**
- **HMAC-SHA256 signature verification** — rejects requests without a valid `X-Linear-Signature` header
- **Idempotent deduplication** — stores event IDs in PostgreSQL, returns `duplicate` for replays
- **Payload transformation** — normalizes Linear's webhook format into SuperPlane's canonical schema: `{source, type, action, data, timestamp, event_id}`
- **Resilient forwarding** — always returns `200 OK` to Linear to prevent retry storms; forwarding failures are logged but non-blocking
- **Health checks** — `/health` endpoint verifies database connectivity for Render's monitoring

---

## Render Services (2)

| Service | Type | Plan | Purpose |
|---|---|---|---|
| `superplane-linear-relay` | Web Service (Docker) | Starter | Runs the Go webhook relay binary. Receives Linear webhooks, validates, deduplicates, transforms, and forwards. |
| `superplane-linear-relay-db` | PostgreSQL | Starter | Managed PostgreSQL for webhook event deduplication. Stores `event_id` + `created_at`; auto-migrated on startup. |

Both services are defined in `linear-relay/render.yaml` and can be deployed to Render with a single `render blueprint apply` command.

---

## Testing

**41 tests passing** across the Linear integration:

| Category | Count | Description |
|---|---|---|
| Integration registration | 1 | Validates 4 triggers and 3 actions are properly registered |
| CreateIssue setup + execute + metadata | 7 | Required field validation, configuration decoding, output channels |
| UpdateIssue setup + metadata | 5 | Required field validation, configuration decoding |
| AddComment setup + execute + metadata | 7 | Required field validation, API key checks, output channels |
| OnIssueStateChange metadata | 1 | Name, label, icon, documentation, example data |
| OnIssueLabelChange metadata | 1 | Name, label, icon, documentation, example data |
| OnCycleChange metadata | 1 | Name, label, icon, documentation, example data |
| OnCommentCreated metadata | 1 | Name, label, icon, documentation, example data |
| WebhookHandler CompareConfig | 3 | Same/different event type matching, invalid config handling |
| Client NewClient | 2 | Missing and empty API key validation |
| ListResources | 1 | Unknown resource type error handling |

All tests run via `make test PKG_TEST_PACKAGES=./pkg/integrations/linear`.

---

## Frontend Mappers (9 files)

The `web_src/src/pages/workflowv2/mappers/linear/` directory contains **9 files** providing canvas UI components for all triggers and actions:

| File | Component |
|---|---|
| `index.ts` | Barrel export for all Linear mappers |
| `types.ts` | Shared TypeScript types for Linear configuration |
| `on_issue_state_change.ts` | On Issue State Change trigger mapper |
| `on_issue_label_change.ts` | On Issue Label Change trigger mapper |
| `on_cycle_change.ts` | On Cycle Change trigger mapper |
| `on_comment_created.ts` | On Comment Created trigger mapper |
| `create_issue.ts` | Create Issue action mapper |
| `update_issue.ts` | Update Issue action mapper |
| `add_comment.ts` | Add Comment action mapper |

These mappers wire each Linear component into the visual canvas editor, providing proper form fields, validation, and example data for the workflow builder UI.

---

## Screenshots / Video

> *Placeholder — add screenshots of the canvas flows in the SuperPlane UI and a video walkthrough before final submission.*

---

## Checklist

- [x] DCO sign-off on all commits
- [x] All 41 integration tests passing (`make test PKG_TEST_PACKAGES=./pkg/integrations/linear`)
- [x] Frontend mappers implemented for all 7 components (4 triggers + 3 actions)
- [x] Canvas template YAML provided (`templates/canvases/linear-pr-sync.yaml`)
- [x] Webhook relay service with HMAC validation and deduplication (`linear-relay/`)
- [x] Render blueprint for one-click deployment (`linear-relay/render.yaml`)
- [x] README documentation for the webhook relay (`linear-relay/README.md`)
- [x] Health check endpoint for Render monitoring (`/health`)
- [x] Conventional Commits PR title format