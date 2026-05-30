import type { ComponentBaseMapper, EventStateRegistry, TriggerRenderer } from "../types";
import { buildActionStateRegistry } from "../utils";
import { createIssueMapper } from "./create_issue";
import { updateIssueMapper } from "./update_issue";
import { addCommentMapper } from "./add_comment";
import { onIssueStateChange } from "./on_issue_state_change";
import { onIssueLabelChange } from "./on_issue_label_change";
import { onCycleChange } from "./on_cycle_change";
import { onCommentCreated } from "./on_comment_created";

export const componentMappers: Record<string, ComponentBaseMapper> = {
  createIssue: createIssueMapper,
  updateIssue: updateIssueMapper,
  addComment: addCommentMapper,
};

export const triggerRenderers: Record<string, TriggerRenderer> = {
  onIssueStateChange: onIssueStateChange,
  onIssueLabelChange: onIssueLabelChange,
  onCycleChange: onCycleChange,
  onCommentCreated: onCommentCreated,
};

export const eventStateRegistry: Record<string, EventStateRegistry> = {
  createIssue: buildActionStateRegistry("created"),
  updateIssue: buildActionStateRegistry("updated"),
  addComment: buildActionStateRegistry("added"),
};

export { createIssueMapper, updateIssueMapper, addCommentMapper, onIssueStateChange, onIssueLabelChange, onCycleChange, onCommentCreated };