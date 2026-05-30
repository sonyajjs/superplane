import { getBackgroundColorClass } from "@/lib/colors";
import type React from "react";
import type { TriggerEventContext, TriggerRenderer, TriggerRendererContext } from "../types";
import type { TriggerProps } from "@/ui/trigger";
import linearIcon from "@/assets/icons/integrations/linear.svg";
import { renderTimeAgo, renderWithTimeAgo } from "@/components/TimeAgo";
import type { LinearLabelChangeConfiguration, LinearLabelChangeEventData } from "./types";

/**
 * Renderer for the "linear.onIssueLabelChange" trigger
 */
export const onIssueLabelChange: TriggerRenderer = {
  getTitleAndSubtitle: (context: TriggerEventContext): { title: string; subtitle: string | React.ReactNode } => {
    const eventData = context.event?.data as LinearLabelChangeEventData | undefined;
    const issue = eventData?.issue;
    const title = issue?.title || "Issue label changed";

    const addedNames = eventData?.added_labels?.map((l) => l.name).filter(Boolean) || [];
    const removedNames = eventData?.removed_labels?.map((l) => l.name).filter(Boolean) || [];
    const parts: string[] = [];
    if (addedNames.length > 0) parts.push(`+${addedNames.join(", ")}`);
    if (removedNames.length > 0) parts.push(`-${removedNames.join(", ")}`);

    const subtitle = buildLabelChangeSubtitle(parts.join(" "), context.event?.createdAt);
    return { title, subtitle };
  },

  getRootEventValues: (context: TriggerEventContext): Record<string, string> => {
    const eventData = context.event?.data as LinearLabelChangeEventData | undefined;
    const details: Record<string, string> = {};

    if (eventData?.issue?.identifier) {
      details["Issue"] = eventData.issue.identifier;
    }
    if (eventData?.added_labels?.length) {
      details["Added Labels"] = eventData.added_labels.map((l) => l.name).join(", ");
    }
    if (eventData?.removed_labels?.length) {
      details["Removed Labels"] = eventData.removed_labels.map((l) => l.name).join(", ");
    }
    if (eventData?.issue?.url) {
      details["URL"] = eventData.issue.url;
    }

    return details;
  },

  getTriggerProps: (context: TriggerRendererContext) => {
    const { node, definition, lastEvent } = context;
    const configuration = node.configuration as LinearLabelChangeConfiguration | undefined;
    const metadataItems: { icon: string; label: string }[] = [];

    if (configuration?.teamId) {
      metadataItems.push({ icon: "users", label: configuration.teamId });
    }
    if (configuration?.labelNames?.length) {
      metadataItems.push({
        icon: "funnel",
        label: "Labels: " + configuration.labelNames.join(", "),
      });
    }

    const props: TriggerProps = {
      title: node.name || "On Issue Label Change",
      iconSrc: linearIcon,
      iconSlug: "tag",
      iconColor: "text-violet-600",
      collapsedBackground: getBackgroundColorClass("violet"),
      metadata: metadataItems,
    };

    if (lastEvent) {
      const eventData = lastEvent.data as LinearLabelChangeEventData | undefined;
      const addedNames = eventData?.added_labels?.map((l) => l.name).filter(Boolean) || [];
      const removedNames = eventData?.removed_labels?.map((l) => l.name).filter(Boolean) || [];
      const parts: string[] = [];
      if (addedNames.length > 0) parts.push(`+${addedNames.join(", ")}`);
      if (removedNames.length > 0) parts.push(`-${removedNames.join(", ")}`);

      const subtitle = buildLabelChangeSubtitle(parts.join(" "), lastEvent.createdAt);
      props.lastEventData = {
        title: eventData?.issue?.title || "Issue label changed",
        subtitle,
        receivedAt: new Date(lastEvent.createdAt),
        state: "triggered",
        eventId: lastEvent.id,
      };
    }

    return props;
  },
};

function buildLabelChangeSubtitle(content: string, createdAt?: string): string | React.ReactNode {
  const trimmed = (content || "").trim();
  if (trimmed && createdAt) {
    return renderWithTimeAgo(trimmed, new Date(createdAt));
  }
  if (createdAt) {
    return renderTimeAgo(new Date(createdAt));
  }
  return trimmed;
}