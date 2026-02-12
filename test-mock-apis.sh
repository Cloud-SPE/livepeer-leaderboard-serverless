#!/usr/bin/env bash
set -euo pipefail

base_url="${BASE_URL:-http://localhost:8080}"

curl -sS "${base_url}/api/gpu/metrics?o_wallet=0xabc123&gpu_id=gpu-1&region=us-west&workflow=inference&time_range=30m" | jq .

curl -sS "${base_url}/api/network/demand?gateway=public&region=us-east&workflow=streaming&interval=10m" | jq .

curl -sS "${base_url}/api/sla/compliance?orchestrator_id=orch-42&period=24h" | jq .

curl -sS "${base_url}/api/datasets?workflow=inference&type=good" | jq .
