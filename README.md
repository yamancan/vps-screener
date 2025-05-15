# VPS Screener

[![CI](https://github.com/yamancan/vps-screener/actions/workflows/ci.yml/badge.svg)](https://github.com/yamancan/vps-screener/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight, open-source VPS monitoring solution that provides per-project visibility, custom metrics, and remote task execution capabilities.

## Features

- **Per-Project Visibility**: Monitor CPU, RAM, disk, network, and custom metrics for each project
- **Minimal Footprint**: Agent runs with ≤ 2% CPU and ≤ 100MB RAM
- **Pluggable Activity Checks**: Support for project-specific plugins
- **Task Management**: Execute remote tasks on VPS nodes
- **Self-Healing**: Built-in service management and auto-restart capabilities

## Project Structure

*   `/agent`: Go agent that runs on each VPS to collect metrics and execute tasks
*   `/api-gateway`: NestJS (Node.js/TypeScript) service acting as the central API
*   `/dashboard`: React (TypeScript, Vite, TailwindCSS) application for visualization
*   `/docs`: Project documentation and guides

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/yamancan/vps-screener.git
cd vps-screener
```

2. Start the development environment:
```bash
docker-compose up -d
```

3. Access the dashboard at `http://localhost:3000`

## Documentation

- [Installation Guide](docs/installation.md)
- [Plugin Development](docs/plugin-development.md)
- [API Documentation](docs/api.md)
- [Contributing Guide](CONTRIBUTING.md)

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- [GitHub Issues](https://github.com/yamancan/vps-screener/issues)
- [Discussions](https://github.com/yamancan/vps-screener/discussions)

## Authors

- [yamancan](https://github.com/yamancan)

## Acknowledgments

- Thanks to all contributors
- Inspired by various monitoring solutions
- Built with modern technologies 