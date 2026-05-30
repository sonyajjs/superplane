import { getBackgroundColorClass } from "@/lib/colors";
import type React from "react";
import type { TriggerEventContext, TriggerRenderer, TriggerRendererContext } from "../types";
import type { TriggerProps } from "@/ui/trigger";
import linearIcon from "@/assets/icons/integrations/linear.svg";
import { renderTimeAgo, renderWithTimeAgo } from "@/components/TimeAgo";
import type { LinearCycleConfiguration, LinearCycleChangeEventData } from "./types";

/**
 * Renderer for the "linear.onCycleChange" trigger
 */
export const onCycleChange: TriggerRenderer = {
  getTitleAndSubtitle: (context: TriggerEventContext): { title: string; subtitle: string | React.ReactNode } => {
    const eventData = context.event?.data as LinearCycleChangeEventData | undefined;
    const cycleName = eventData?.cycle?.name || "Cycle change";
    const title = cycleName;

    const issueCount = eventData?.issues?.length;
    let subtitleText = "";
    if (issueCount !== undefined && issueCount > 0) {
      subtitleText = `${issueCount} issue${issueCount !== 1 ? "s" : ""} affected`;
    }

    const subtitle = buildCycleSubtitle(subtitleText, context.event?.createdAt);
    return { title, subtitle };
  },

  getRootEventValues: (context: TriggerEventContext): Record<string, string> => {
    const eventData = context.event?.data as LinearCycleChangeEventData | undefined;
    const details: Record<string, string> = {};

    if (eventData?.cycle?.name) {
      details["Cycle"] = eventData.cycle.name;
    }
    if (eventData?.team?.name) {
      details["Team"] = eventData.team.name;
    }
    if (eventData?.issues?.length) {
      details["Issues"] = eventData.issues.map((i) => i.identifier).join(", ");
    }

    return details;
  },

  getTriggerProps: (context: TriggerRendererContext) => {
    const { node, definition, lastEvent } = context;
    const configuration = node.configuration as LinearCycleConfiguration | undefined;
    const metadataItems: { icon: string; label: string }[] = [];

    if (configuration?.teamId) {
      metadataItems.push({ icon: "users", label: configuration.teamId });
    }

    const props: TriggerProps = {
      title: node.name || "On Cycle Change",
      iconSrc: linearIcon,
      iconSlug: "calendar",
      iconColor: "text-violet-600",
      collapsedBackground: getBackgroundColorClass("violet"),
      metadata: metadataItems,
    };

    if (lastEvent) {
      const eventData = lastEvent.data as LinearCycleChangeEventData | undefined;
      const cycleName = eventData?.cycle?.name || "Cycle change";
      const issueCount = eventData?.issues?.length;
      let subtitleText = "";
      if (issueCount !== undefined && issueCount > 0) {
        subtitleText = `${issueCount} issue${issueCount !== 1 ? "s" : ""} affected`;
      }
      const subtitle = buildCycleSubtitle(subtitleText, lastEvent.createdAt);
      props.lastEventData = {
        title: cycleName,
        subtitle,
        receivedAt: new Date(lastEvent.createdAt),
        state: "triggered",
        eventId: lastEvent.id,
      };
    }

    return props;
  },
};

function buildCycleSubtitle(content: string, createdAt?: string): string | React.ReactNode {
  const trimmed = (content || "").trim();
  if (trimmed && createdAt) {
    return renderWithTimeAgo(trimmed, new Date(createdAt));
  }
  if (createdAt) {
    return renderTimeAgo(new Date(createdAt));
  }
  return trimmed;
}