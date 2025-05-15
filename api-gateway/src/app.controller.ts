import { Controller, Get, Post, Body, Query, HttpCode } from '@nestjs/common';
import { AppService } from './app.service';

@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}

  @Get()
  getHello(): string {
    return this.appService.getHello();
  }

  // Endpoint to receive metrics from agents
  @Post('v1/metrics')
  @HttpCode(200) // Respond with 200 OK for successful receipt, though 201/202 might be more appropriate later
  receiveMetrics(@Body() metricsData: any) { // Define a type/interface for metricsData later
    console.log(`Received metrics from agent: ${JSON.stringify(metricsData)}`);
    // Here, you would typically save the metrics to a database or process them.
    return { message: 'Metrics received successfully' };
  }

  // Endpoint for agents to fetch tasks
  @Get('v1/tasks')
  getTasks(@Query('node') nodeId: string) {
    console.log(`Agent node [${nodeId}] requested tasks.`);
    // Here, you would fetch tasks for the specific node from a database or queue.
    // For now, returning an empty array.
    return []; 
  }
} 