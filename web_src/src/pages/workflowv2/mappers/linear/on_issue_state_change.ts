import { getBackgroundColorClass } from "@/lib/colors";
import type React from "react";
import type { TriggerEventContext, TriggerRenderer, TriggerRendererContext } from "../types";
import type { TriggerProps } from "@/ui/trigger";
import linearIcon from "@/assets/icons/integrations/linear.svg";
import { renderTimeAgo, renderWithTimeAgo } from "@/components/TimeAgo";
import type { LinearStateChangeConfiguration, LinearStateChangeEventData } from "./types";

/**
 * Renderer for the "linear.onIssueStateChange" trigger
 */
export const onIssueStateChange: TriggerRenderer = {
  getTitleAndSubtitle: (context: TriggerEventContext): { title: string; subtitle: string | React.ReactNode } => {
    const eventData = context.event?.data as LinearStateChangeEventData | undefined;
    const issue = eventData?.issue;
    const fromState = eventData?.from_state?.name;
    const toState = eventData?.to_state?.name;

    const title = issue?.title || "Issue state changed";

    let subtitleText = "";
    if (fromState && toState) {
      subtitleText = `${fromState} → ${toState}`;
    } else if (toState) {
      subtitleText = `State: ${toState}`;
    }

    const subtitle = buildLinearSubtitle(subtitleText, context.event?.createdAt);

    return { title, subtitle };
  },

  getRootEventValues: (context: TriggerEventContext): Record<string, string> => {
    const eventData = context.event?.data as LinearStateChangeEventData | undefined;
    const details: Record<string, string> = {};

    if (eventData?.issue?.identifier) {
      details["Issue"] = eventData.issue.identifier;
    }
    if (eventData?.from_state?.name) {
      details["From State"] = eventData.from_state.name;
    }
    if (eventData?.to_state?.name) {
      details["To State"] = eventData.to_state.name;
    }
    if (eventData?.issue?.url) {
      details["URL"] = eventData.issue.url;
    }

    return details;
  },

  getTriggerProps: (context: TriggerRendererContext) => {
    const { node, definition, lastEvent } = context;
    const configuration = node.configuration as LinearStateChangeConfiguration | undefined;
    const metadataItems: { icon: string; label: string }[] = [];

    if (configuration?.teamId) {
      metadataItems.push({ icon: "users", label: configuration.teamId });
    }
    if (configuration?.stateTypes?.length) {
      metadataItems.push({
        icon: "funnel",
        label: "States: " + configuration.stateTypes.join(", "),
      });
    }

    const props: TriggerProps = {
      title: node.name || "On Issue State Change",
      iconSrc: linearIcon,
      iconSlug: "arrow-right-circle",
      iconColor: "text-violet-600",
      collapsedBackground: getBackgroundColorClass("violet"),
      metadata: metadataItems,
    };

    if (lastEvent) {
      const eventData = lastEvent.data as LinearStateChangeEventData | undefined;
      const fromState = eventData?.from_state?.name;
      const toState = eventData?.to_state?.name;
      let subtitleText = "";
      if (fromState && toState) {
        subtitleText = `${fromState} → ${toState}`;
      }
      const subtitle = buildLinearSubtitle(subtitleText, lastEvent.createdAt);
      props.lastEventData = {
        title: eventData?.issue?.title || "Issue state changed",
        subtitle,
        receivedAt: new Date(lastEvent.createdAt),
        state: "triggered",
        eventId: lastEvent.id,
      };
    }

    return props;
  },
};

function buildLinearSubtitle(content: string, createdAt?: string): string | React.ReactNode {
  const trimmed = (content || "").trim();
  if (trimmed && createdAt) {
    return renderWithTimeAgo(trimmed, new Date(createdAt));
  }
  if (createdAt) {
    return renderTimeAgo(new Date(createdAt));
  }
  return trimmed;
}