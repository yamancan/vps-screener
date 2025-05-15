Technical Blueprint

1 — Data Flow

┌─────────────┐   HTTPS/JWT   ┌──────────────┐  WebSocket/REST  ┌──────────────┐
│  Agent      │  ───────────▶│ API-Gateway   │ ───────────────▶ │  Dashboard   │
│ (per VPS)   │              │  (NestJS/Fastify)    │                 │  (React)     │
└─────────────┘◀─Task JSON───└──────────────┘◀─Browser poll────└──────────────┘

	•	Push only: metrics & logs buffered locally if offline.
	•	Task channel: small JSON jobs flow API → Agent; results flow back on the same pipe.

2 — Agent Internals (Go Implementation)

Package/Module	Responsibility
`main.go`		Entry point, main loop, signal handling.
`config/`		Loads and parses `config.yaml`.
`mapper/`		Maps PID ➜ Project based on config rules (systemd, Docker, user, etc.).
`collector/`	Collects system and per-project metrics using `gopsutil` and mapper.
`plugins/`	(Design TBD) Handles project-specific custom metrics. The `config.yaml` specifies plugin paths. The agent will need a mechanism to execute these, potentially as Go plugins compiled with the agent, or as external scripts/binaries invoked via a defined interface (e.g., expecting JSON output on stdout).
`sender/`		Sends collected metrics to the API Gateway.
`executor/`	Pulls queued tasks from the API, executes them securely (with resource limits TBD), and posts results.

Config (config.yaml):

api:  https://api.example.com
token: ${JWT}
interval: 30  # seconds
projects:
  ProjectA:
    match: {user: projAuser}
    plugin: plugins/projectA_plugin.py
  ProjectB:
    match: {container: project_b_*}

3 — Gateway & DB (PostgreSQL)

Table	Columns (key ones)
nodes	id, hostname, last_seen, version
projects	id, node_id, name
metrics	ts, project_id, cpu, ram, disk, net_in, net_out, custom JSONB
tasks	id, project_id, cmd, status, output, created_at

Main End-points (all JWT-auth):

Method	Path	Purpose
POST	/v1/metrics	Agent bulk upload
GET	/v1/tasks?node=	Agent pulls pending jobs
POST	/v1/tasks/{id}/result	Send execution result
POST	/v1/tasks	Dashboard/API user schedules job

4 — Dashboard Panels
	•	Node Grid: heartbeat badge, resource totals.
	•	Project Cards: plugin summary, real-time graphs (WebSocket + Chart.js or SWR as fallback).
	•	Task Console: live output via server-sent events.

⸻

Controlling & Monitoring Services on Each Node

A. Project Detection Strategies
	1.	systemd Tagging (lightest)

[Service]
ExecStart=/opt/projectA/bin/node
User=projAuser          # mapper picks this


	2.	Docker Labels

docker run -d --label project=ProjectB --name project_b_node ...


	3.	Kubernetes (future) — use namespace = project.

B. Service Health Checks
	•	systemd: systemctl is-active projA-node; plugin returns "running"/"failed".
	•	RPC: plugin hits http://localhost:26657/status → parse catching_up.
	•	Log scan: journalctl -u projB-worker -n 100 --no-pager | grep -qi error triggers alert.

C. Remote Actions from Dashboard

Action	How it Works
Restart service	Task {cmd: "systemctl restart projA-node"}
Pull new Docker image	docker compose -p project_b pull && up -d
Disk cleanup	rm -rf /var/log/projectA/*.gz

Each job result appears in the Task Console with exit code, stdout, stderr.

D. Adding a New Project in 3 Steps
	1.	Create mapping in config.yaml.
	2.	Write plugin (optional) with collect() returning custom dict.
	3.	Reload agent → systemctl restart vps-agent.

