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

interface EstimateCostConfiguration {
  serviceIds?: string;
}

interface ServiceCostBreakdown {
  serviceId?: string;
  serviceName?: string;
  type?: string;
  plan?: string;
  monthlyUSD?: number;
  suspended?: boolean;
}

interface EstimateCostResult {
  breakdown?: ServiceCostBreakdown[];
  totalMonthlyUSD?: number;
  serviceCount?: number;
}

function metadataList(node: NodeInfo): MetadataItem[] {
  const metadata: MetadataItem[] = [];
  const configuration = node.configuration as EstimateCostConfiguration | undefined;

  if (configuration?.serviceIds) {
    metadata.push({ icon: "server", label: `Services: ${configuration.serviceIds}` });
  } else {
    metadata.push({ icon: "server", label: "All services" });
  }

  return metadata;
}

export const estimateCostMapper: ComponentBaseMapper = {
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
    const result = outputs?.default?.[0]?.data as EstimateCostResult | undefined;

    return {
      "Services": result?.serviceCount != null ? String(result.serviceCount) : "-",
      "Monthly Cost": result?.totalMonthlyUSD != null ? `$${result.totalMonthlyUSD.toFixed(2)}` : "-",
    };
  },
};
