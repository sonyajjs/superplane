import type { ComponentBaseProps } from "@/ui/componentBase";
import type React from "react";
import type {
  ComponentBaseContext,
  ComponentBaseMapper,
  ExecutionDetailsContext,
  NodeInfo,
  OutputPayload,
  SubtitleContext,
} from "../types";
import type { MetadataItem } from "@/ui/metadataList";
import { renderTimeAgo } from "@/components/TimeAgo";
import { stringOrDash } from "./common";
import { baseProps } from "./base";

interface GetLogsConfiguration {
  service?: string;
  lines?: number;
}

interface GetLogsOutput {
  serviceId?: string;
  lines?: Array<{
    timestamp?: string;
    message?: string;
  }>;
  count?: number;
}

function metadataList(node: NodeInfo): MetadataItem[] {
  const metadata: MetadataItem[] = [];
  const configuration = node.configuration as GetLogsConfiguration | undefined;

  if (configuration?.service) {
    metadata.push({ icon: "server", label: `Service: ${configuration.service}` });
  }
  if (configuration?.lines) {
    metadata.push({ icon: "file-text", label: `Lines: ${configuration.lines}` });
  }

  return metadata;
}

export const getLogsMapper: ComponentBaseMapper = {
  props(context: ComponentBaseContext): ComponentBaseProps {
    const base = baseProps(context.nodes, context.node, context.componentDefinition, context.lastExecutions);
    return { ...base, metadata: metadataList(context.node) };
  },

  subtitle(context: SubtitleContext): string | React.ReactNode {
    const timestamp = context.execution.updatedAt || context.execution.createdAt;
    return timestamp ? renderTimeAgo(new Date(timestamp)) : "";
  },

  getExecutionDetails(context: ExecutionDetailsContext): Record<string, string> {
    const outputs = context.execution.outputs as { default?: OutputPayload[] } | undefined;
    const result = outputs?.default?.[0]?.data as GetLogsOutput | undefined;

    return {
      "Service ID": stringOrDash(result?.serviceId),
      "Log Lines": result?.count != null ? String(result.count) : "-",
    };
  },
};
