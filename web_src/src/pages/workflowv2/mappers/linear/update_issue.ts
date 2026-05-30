import type {
  ComponentBaseContext,
  ComponentBaseMapper,
  ExecutionDetailsContext,
  NodeInfo,
  SubtitleContext,
} from "../types";
import type { ComponentBaseProps, EventSection } from "@/ui/componentBase";
import type React from "react";
import { getBackgroundColorClass } from "@/lib/colors";
import { getState, getStateMap, getTriggerRenderer } from "..";
import type { MetadataItem } from "@/ui/metadataList";
import linearIcon from "@/assets/icons/integrations/linear.svg";
import { renderTimeAgo } from "@/components/TimeAgo";
import type { LinearIssueConfiguration } from "./types";

export const updateIssueMapper: ComponentBaseMapper = {
  props(context: ComponentBaseContext): ComponentBaseProps {
    const lastExecution = context.lastExecutions.length > 0 ? context.lastExecutions[0] : null;
    const componentName = context.componentDefinition.name || "linear.updateIssue";

    return {
      iconSrc: linearIcon,
      iconSlug: "pencil",
      iconColor: "text-violet-600",
      collapsedBackground: getBackgroundColorClass("violet"),
      collapsed: context.node.isCollapsed,
      title:
        context.node.name ||
        context.componentDefinition.label ||
        context.componentDefinition.name ||
        "Update Issue",
      eventSections: lastExecution ? updateIssueEventSections(context.nodes, lastExecution, componentName) : undefined,
      metadata: updateIssueMetadataList(context.node),
      includeEmptyState: !lastExecution,
      eventStateMap: getStateMap(componentName),
    };
  },

  getExecutionDetails(context: ExecutionDetailsContext): Record<string, string> {
    const details: Record<string, string> = {};

    if (context.execution.createdAt) {
      details["Started at"] = new Date(context.execution.createdAt).toLocaleString();
    }

    if (context.execution.state === "STATE_FINISHED" && context.execution.updatedAt) {
      details["Finished at"] = new Date(context.execution.updatedAt).toLocaleString();
    }

    return details;
  },

  subtitle(context: SubtitleContext): string | React.ReactNode {
    const configuration = context.node.configuration as LinearIssueConfiguration | undefined;
    if (configuration?.issueId) {
      return `Issue ${configuration.issueId}`;
    }
    if (!context.execution.createdAt) return "";
    return renderTimeAgo(new Date(context.execution.createdAt));
  },
};

function updateIssueMetadataList(node: NodeInfo): MetadataItem[] {
  const metadata: MetadataItem[] = [];
  const configuration = node.configuration as LinearIssueConfiguration | undefined;

  if (configuration?.issueId) {
    metadata.push({ icon: "hash", label: configuration.issueId });
  }

  if (configuration?.teamId) {
    metadata.push({ icon: "users", label: `Team: ${configuration.teamId}` });
  }

  return metadata;
}

function updateIssueEventSections(nodes: NodeInfo[], execution: any, componentName: string): EventSection[] {
  const rootTriggerNode = nodes.find((n: NodeInfo) => n.id === execution.rootEvent?.nodeId);
  if (!rootTriggerNode) {
    return [];
  }

  const rootTriggerRenderer = getTriggerRenderer(rootTriggerNode.componentName || "");
  const { title } = rootTriggerRenderer.getTitleAndSubtitle({ event: execution.rootEvent });

  return [
    {
      receivedAt: new Date(execution.createdAt),
      eventTitle: title,
      eventSubtitle: renderTimeAgo(new Date(execution.createdAt)),
      eventState: getState(componentName)(execution),
      eventId: execution.rootEvent?.id || "",
    },
  ];
}