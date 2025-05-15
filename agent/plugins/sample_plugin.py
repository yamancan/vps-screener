#!/usr/bin/env python3

import json
import os
import random

# Example: Accessing an environment variable set by the agent
project_name = os.getenv("VPS_PROJECT_NAME", "unknown_project")

data = {
    "custom_metric_A": random.randint(1, 1000),
    "custom_status": "ok",
    "processed_for_project": project_name,
    "plugin_version": "1.0.1"
}

print(json.dumps(data)) 