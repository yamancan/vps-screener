import { Injectable, Logger } from '@nestjs/common';

// Define a simple interface for the structure of metrics we expect
export interface Metric {
  cpu_percent: number;
  ram_bytes?: number;
  ram_percent?: number;
  disk_percent?: number;
  // Add other specific metrics if they are sent by the agent
}

export interface NodeData {
  node_hostname: string;
  metrics_data: {
    _system?: Metric;
    [projectName: string]: Metric | undefined; // For project-specific metrics
  };
  timestamp: number; // Unix timestamp from agent
}

export interface StoredNodeInfo {
  nodeId: string;
  metrics: {
    _system?: Metric;
    [projectName: string]: Metric | undefined;
  };
  last_updated_gateway: Date; // Timestamp when gateway last received data for this node
  last_agent_timestamp: Date; // Timestamp from the agent's data
}

@Injectable()
export class AppService {
  private readonly logger = new Logger(AppService.name);
  private latestMetrics: Map<string, StoredNodeInfo> = new Map();

  getHello(): string {
    return 'Hello from API Gateway!';
  }

  // Method to store/update metrics from an agent
  storeMetrics(data: NodeData): void {
    if (!data || !data.node_hostname || !data.metrics_data) {
      this.logger.warn('Received invalid metrics data structure', JSON.stringify(data));
      return;
    }

    const { node_hostname, metrics_data, timestamp } = data;

    this.latestMetrics.set(node_hostname, {
      nodeId: node_hostname,
      metrics: metrics_data,
      last_updated_gateway: new Date(),
      last_agent_timestamp: new Date(timestamp * 1000), // Convert Unix to JS Date
    });

    this.logger.log(`Metrics updated for node: ${node_hostname}`);
  }

  // Method to get statuses for all nodes, suitable for the dashboard
  getNodeStatuses(): StoredNodeInfo[] {
    return Array.from(this.latestMetrics.values());
  }
} 