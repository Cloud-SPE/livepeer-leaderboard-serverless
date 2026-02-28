#!/usr/bin/env bash
set -euo pipefail

base_url="${BASE_URL:-http://localhost:8084}"

echo "=== Health ==="
curl -sS "${base_url}/api/health" | jq .

echo "=== GPU Metrics (by orchestrator + pipeline) ==="
curl -sS "${base_url}/api/gpu/metrics?orchestrator_address=0x0abe02f6ef1fa8c29f9b3f9f170c6f3681fd3031&pipeline=streamdiffusion-sdxl-v2v&time_range=24h" | jq .

echo "=== GPU Metrics (unfiltered, last hour) ==="
curl -sS "${base_url}/api/gpu/metrics?time_range=1h" | jq .

echo "=== Network Demand (by gateway) ==="
curl -sS "${base_url}/api/network/demand?gateway=cloud-spe-ai-live-video-tester-mdw&pipeline=streamdiffusion-sdxl&model_id=streamdiffusion-sdxl&interval=15m" | jq .

echo "=== Network Demand (unfiltered) ==="
curl -sS "${base_url}/api/network/demand?interval=1h" | jq .

echo "=== SLA Compliance (by orchestrator + model) ==="
curl -sS "${base_url}/api/sla/compliance?orchestrator_address=0x5263e0ce3a97b634d8828ce4337ad0f70b30b077&model_id=meta-llama/Meta-Llama-3.1-8B-Instruct&period=24h" | jq .

echo "=== SLA Compliance (unfiltered, 7 days) ==="
curl -sS "${base_url}/api/sla/compliance?period=168h" | jq .

echo "=== Datasets (inference, good) ==="
curl -sS "${base_url}/api/datasets?workflow=inference&type=good" | jq .
