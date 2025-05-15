# Installation Guide

This guide will help you set up the VPS Screener monitoring solution.

## Prerequisites

- Docker and Docker Compose
- Git
- Node.js 18+ (for development)
- Go 1.21+ (for development)

## Quick Installation

1. Clone the repository:
```bash
git clone https://github.com/yamancan/vps-screener.git
cd vps-screener
```

2. Configure the environment:
```bash
# Copy example configs
cp agent/config.example.yaml agent/config.yaml
cp api-gateway/.env.example api-gateway/.env
cp dashboard/.env.example dashboard/.env
```

3. Start the services:
```bash
docker-compose up -d
```

4. Verify the installation:
- Dashboard: http://localhost:3000
- API Gateway: http://localhost:3001
- Agent metrics endpoint: http://localhost:8080/metrics

## Manual Installation

### API Gateway

1. Navigate to the API Gateway directory:
```bash
cd api-gateway
```

2. Install dependencies:
```bash
npm install
```

3. Build the application:
```bash
npm run build
```

4. Start the service:
```bash
npm run start:prod
```

### Dashboard

1. Navigate to the Dashboard directory:
```bash
cd dashboard
```

2. Install dependencies:
```bash
npm install
```

3. Start the development server:
```bash
npm run dev
```

### Agent

1. Navigate to the Agent directory:
```bash
cd agent
```

2. Install Go dependencies:
```bash
go mod download
```

3. Build the agent:
```bash
go build -o vps-agent
```

4. Configure the agent:
```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your settings
```

5. Run the agent:
```bash
./vps-agent
```

## Configuration

### Agent Configuration

The agent's configuration is managed through `config.yaml`. Key settings include:

```yaml
api_gateway:
  url: "http://localhost:3001"
  token: "your-jwt-token"

interval: 30  # seconds

projects:
  my_project:
    match:
      systemd_service: my-app.service
    plugin: plugins/my_metrics.py
```

### API Gateway Configuration

The API Gateway uses environment variables for configuration. Create a `.env` file with:

```env
PORT=3001
JWT_SECRET=your-secret-key
DATABASE_URL=postgresql://user:password@localhost:5432/vps_screener
```

### Dashboard Configuration

The dashboard also uses environment variables. Create a `.env` file with:

```env
VITE_API_URL=http://localhost:3001
```

## Troubleshooting

### Common Issues

1. **Agent can't connect to API Gateway**
   - Check the API Gateway URL in agent's config.yaml
   - Verify network connectivity
   - Check JWT token validity

2. **Dashboard shows no data**
   - Verify API Gateway is running
   - Check browser console for errors
   - Verify environment variables

3. **Plugins not working**
   - Check plugin file permissions
   - Verify plugin output format
   - Check agent logs for errors

### Logs

- Agent logs: `journalctl -u vps-agent`
- API Gateway logs: `docker-compose logs api-gateway`
- Dashboard logs: `docker-compose logs dashboard`

## Security Considerations

1. **JWT Tokens**
   - Use strong, unique tokens
   - Rotate tokens regularly
   - Store tokens securely

2. **Network Security**
   - Use HTTPS in production
   - Restrict API Gateway access
   - Use firewall rules

3. **Plugin Security**
   - Review plugin code
   - Run plugins with minimal privileges
   - Validate plugin output

## Next Steps

- [Plugin Development Guide](plugin-development.md)
- [API Documentation](api.md)
- [Contributing Guide](../CONTRIBUTING.md) 