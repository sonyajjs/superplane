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

export const createIssueMapper: ComponentBaseMapper = {
  props(context: ComponentBaseContext): ComponentBaseProps {
    const lastExecution = context.lastExecutions.length > 0 ? context.lastExecutions[0] : null;
    const componentName = context.componentDefinition.name || "linear.createIssue";

    return {
      iconSrc: linearIcon,
      iconSlug: "plus-circle",
      iconColor: "text-violet-600",
      collapsedBackground: getBackgroundColorClass("violet"),
      collapsed: context.node.isCollapsed,
      title:
        context.node.name ||
        context.componentDefinition.label ||
        context.componentDefinition.name ||
        "Create Issue",
      eventSections: lastExecution ? createIssueEventSections(context.nodes, lastExecution, componentName) : undefined,
      metadata: createIssueMetadataList(context.node),
      specs: createIssueSpecs(context.node),
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
    if (configuration?.title) {
      return configuration.title;
    }
    if (!context.execution.createdAt) return "";
    return renderTimeAgo(new Date(context.execution.createdAt));
  },
};

function createIssueMetadataList(node: NodeInfo): MetadataItem[] {
  const metadata: MetadataItem[] = [];
  const configuration = node.configuration as LinearIssueConfiguration | undefined;

  if (configuration?.teamId) {
    metadata.push({ icon: "users", label: `Team: ${configuration.teamId}` });
  }

  return metadata;
}

function createIssueSpecs(node: NodeInfo): ComponentBaseProps["specs"] {
  const specs: ComponentBaseProps["specs"] = [];
  const configuration = node.configuration as LinearIssueConfiguration | undefined;

  if (configuration?.title) {
    specs.push({
      title: "title",
      tooltipTitle: "title",
      iconSlug: "file-text",
      value: configuration.title,
      contentType: "text",
    });
  }

  if (configuration?.description) {
    specs.push({
      title: "description",
      tooltipTitle: "description",
      iconSlug: "align-left",
      value: configuration.description,
      contentType: "text",
    });
  }

  return specs;
}

function createIssueEventSections(nodes: NodeInfo[], execution: any, componentName: string): EventSection[] {
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