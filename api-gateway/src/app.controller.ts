import { Controller, Get, Post, Body, Query, HttpCode, Logger } from '@nestjs/common';
import { AppService } from './app.service';

// Define a simple interface for the structure of metrics we expect from agent
// This should ideally be in a shared types file if agent and gateway are in a monorepo
// or published as a package if they are separate.
interface Metric {
  cpu_percent: number;
  ram_bytes?: number;
  ram_percent?: number;
  disk_percent?: number;
}

interface AgentMetricsPayload {
  node_hostname: string;
  metrics_data: {
    _system?: Metric;
    [projectName: string]: Metric | undefined;
  };
  timestamp: number;
}

@Controller()
export class AppController {
  private readonly logger = new Logger(AppController.name);

  constructor(private readonly appService: AppService) {}

  @Get()
  getHello(): string {
    return this.appService.getHello();
  }

  // Endpoint to receive metrics from agents
  @Post('v1/metrics')
  @HttpCode(200) // Respond with 200 OK for successful receipt, though 201/202 might be more appropriate later
  receiveMetrics(@Body() metricsData: AgentMetricsPayload) { // MODIFIED: Use specific type
    this.logger.log(`Received metrics from agent ${metricsData.node_hostname}`); // MODIFIED: More specific log
    this.appService.storeMetrics(metricsData); // MODIFIED: Call service to store metrics
    return { message: 'Metrics received and stored successfully' }; // MODIFIED: More specific message
  }

  // Endpoint for agents to fetch tasks
  @Get('v1/tasks')
  getTasks(@Query('node') nodeId: string) {
    this.logger.log(`Agent node [${nodeId}] requested tasks.`);
    // Here, you would fetch tasks for the specific node from a database or queue.
    // For now, returning an empty array.
    return []; 
  }

  // New endpoint for the dashboard to fetch all node statuses
  @Get('v1/status')
  getNodeStatuses() {
    this.logger.log('Dashboard requested node statuses.');
    return this.appService.getNodeStatuses();
  }
} 