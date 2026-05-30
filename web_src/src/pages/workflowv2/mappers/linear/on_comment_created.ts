import { getBackgroundColorClass } from "@/lib/colors";
import type React from "react";
import type { TriggerEventContext, TriggerRenderer, TriggerRendererContext } from "../types";
import type { TriggerProps } from "@/ui/trigger";
import linearIcon from "@/assets/icons/integrations/linear.svg";
import { renderTimeAgo, renderWithTimeAgo } from "@/components/TimeAgo";
import type { LinearCommentConfiguration, LinearCommentEventData } from "./types";

/**
 * Renderer for the "linear.onCommentCreated" trigger
 */
export const onCommentCreated: TriggerRenderer = {
  getTitleAndSubtitle: (context: TriggerEventContext): { title: string; subtitle: string | React.ReactNode } => {
    const eventData = context.event?.data as LinearCommentEventData | undefined;
    const issue = eventData?.issue;
    const title = issue?.title || "Comment created";

    const userName = eventData?.comment?.user?.name;
    let subtitleText = "";
    if (userName) {
      subtitleText = `by ${userName}`;
    }

    const subtitle = buildCommentSubtitle(subtitleText, context.event?.createdAt);
    return { title, subtitle };
  },

  getRootEventValues: (context: TriggerEventContext): Record<string, string> => {
    const eventData = context.event?.data as LinearCommentEventData | undefined;
    const details: Record<string, string> = {};

    if (eventData?.issue?.identifier) {
      details["Issue"] = eventData.issue.identifier;
    }
    if (eventData?.comment?.user?.name) {
      details["Author"] = eventData.comment.user.name;
    }
    if (eventData?.comment?.body) {
      const body = eventData.comment.body;
      details["Comment"] = body.length > 200 ? body.substring(0, 200) + "..." : body;
    }
    if (eventData?.issue?.url) {
      details["URL"] = eventData.issue.url;
    }

    return details;
  },

  getTriggerProps: (context: TriggerRendererContext) => {
    const { node, definition, lastEvent } = context;
    const configuration = node.configuration as LinearCommentConfiguration | undefined;
    const metadataItems: { icon: string; label: string }[] = [];

    if (configuration?.issueId) {
      metadataItems.push({ icon: "hash", label: configuration.issueId });
    }

    const props: TriggerProps = {
      title: node.name || "On Comment Created",
      iconSrc: linearIcon,
      iconSlug: "message-square",
      iconColor: "text-violet-600",
      collapsedBackground: getBackgroundColorClass("violet"),
      metadata: metadataItems,
    };

    if (lastEvent) {
      const eventData = lastEvent.data as LinearCommentEventData | undefined;
      const userName = eventData?.comment?.user?.name;
      let subtitleText = "";
      if (userName) {
        subtitleText = `by ${userName}`;
      }
      const subtitle = buildCommentSubtitle(subtitleText, lastEvent.createdAt);
      props.lastEventData = {
        title: eventData?.issue?.title || "Comment created",
        subtitle,
        receivedAt: new Date(lastEvent.createdAt),
        state: "triggered",
        eventId: lastEvent.id,
      };
    }

    return props;
  },
};

function buildCommentSubtitle(content: string, createdAt?: string): string | React.ReactNode {
  const trimmed = (content || "").trim();
  if (trimmed && createdAt) {
    return renderWithTimeAgo(trimmed, new Date(createdAt));
  }
  if (createdAt) {
    return renderTimeAgo(new Date(createdAt));
  }
  return trimmed;
}