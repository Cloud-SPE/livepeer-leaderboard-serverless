#!/usr/bin/env bash
set -euo pipefail

base_url="${BASE_URL:-http://localhost:8080}"

uri_encode() {
  jq -rn --arg v "$1" '$v|@uri'
}

fetch_body() {
  local url="$1"
  local tmp_body
  local status
  tmp_body="$(mktemp)"
  status="$(curl -sS -o "${tmp_body}" -w '%{http_code}' "${url}")"
  if [[ "${status}" -lt 200 || "${status}" -ge 300 ]]; then
    echo "Request failed: ${url} (HTTP ${status})" >&2
    cat "${tmp_body}" >&2
    rm -f "${tmp_body}"
    exit 1
  fi
  cat "${tmp_body}"
  rm -f "${tmp_body}"
}

ensure_json() {
  local label="$1"
  local payload="$2"
  if ! printf '%s\n' "${payload}" | jq -e . >/dev/null 2>&1; then
    echo "Invalid JSON response for ${label}" >&2
    printf '%s\n' "${payload}" >&2
    exit 1
  fi
}

echo "=== Health ==="
health_json="$(fetch_body "${base_url}/api/health")"
ensure_json "health" "${health_json}"
printf '%s\n' "${health_json}" | jq .

gpu_base_json="$(fetch_body "${base_url}/api/gpu/metrics?time_range=24h")"
network_base_json="$(fetch_body "${base_url}/api/network/demand?interval=24h")"
sla_base_json="$(fetch_body "${base_url}/api/sla/compliance?period=720h")"
ensure_json "gpu/metrics base" "${gpu_base_json}"
ensure_json "network/demand base" "${network_base_json}"
ensure_json "sla/compliance base" "${sla_base_json}"

echo "=== Base Check: GPU Metrics (unfiltered, max window 24h) ==="
printf '%s\n' "${gpu_base_json}" \
  | jq '{count: (.metrics | length), first_window_start: (.metrics[0].window_start // null), sample: (.metrics[0] // null)}'

echo "=== Base Check: Network Demand (unfiltered, widest interval 24h) ==="
printf '%s\n' "${network_base_json}" \
  | jq '{count: (.demand | length), first_window_start: (.demand[0].window_start // null), sample: (.demand[0] // null)}'

echo "=== Base Check: SLA Compliance (unfiltered, max period 30d) ==="
printf '%s\n' "${sla_base_json}" \
  | jq '{count: (.compliance | length), first_window_start: (.compliance[0].window_start // null), sample: (.compliance[0] // null)}'

gpu_orch="$(printf '%s\n' "${gpu_base_json}" | jq -r '.metrics | map(.orchestrator_address) | map(select(. != null and . != "")) | first // empty')"
gpu_pipeline_id="$(printf '%s\n' "${gpu_base_json}" | jq -r '.metrics | map(.pipeline_id) | map(select(. != null and . != "")) | first // empty')"
gpu_org="$(printf '%s\n' "${gpu_base_json}" | jq -r '.metrics | map(.org) | map(select(. != null and . != "")) | first // empty')"

network_gateway="$(printf '%s\n' "${network_base_json}" | jq -r '.demand | map(.gateway) | map(select(. != null and . != "")) | first // empty')"
network_pipeline_id="$(printf '%s\n' "${network_base_json}" | jq -r '.demand | map(.pipeline_id) | map(select(. != null and . != "")) | first // empty')"
network_model_id="$(printf '%s\n' "${network_base_json}" | jq -r '.demand | map(.model_id) | map(select(. != null and . != "")) | first // empty')"
network_org="$(printf '%s\n' "${network_base_json}" | jq -r '.demand | map(.org) | map(select(. != null and . != "")) | first // empty')"

sla_orch="$(printf '%s\n' "${sla_base_json}" | jq -r '.compliance | map(.orchestrator_address) | map(select(. != null and . != "")) | first // empty')"
sla_model_id="$(printf '%s\n' "${sla_base_json}" | jq -r '.compliance | map(.model_id) | map(select(. != null and . != "")) | first // empty')"
sla_pipeline_id="$(printf '%s\n' "${sla_base_json}" | jq -r '.compliance | map(.pipeline_id) | map(select(. != null and . != "")) | first // empty')"
sla_org="$(printf '%s\n' "${sla_base_json}" | jq -r '.compliance | map(.org) | map(select(. != null and . != "")) | first // empty')"

echo "=== GPU Metrics (by orchestrator + pipeline_id) ==="
if [[ -n "${gpu_orch}" || -n "${gpu_pipeline_id}" ]]; then
  gpu_filtered_url="${base_url}/api/gpu/metrics?time_range=24h"
  if [[ -n "${gpu_orch}" ]]; then
    gpu_filtered_url="${gpu_filtered_url}&orchestrator_address=$(uri_encode "${gpu_orch}")"
  fi
  if [[ -n "${gpu_pipeline_id}" ]]; then
    gpu_filtered_url="${gpu_filtered_url}&pipeline_id=$(uri_encode "${gpu_pipeline_id}")"
  fi
  echo "Using derived filters: orchestrator_address=${gpu_orch:-<none>} pipeline_id=${gpu_pipeline_id:-<none>}"
  gpu_filtered_json="$(fetch_body "${gpu_filtered_url}")"
  ensure_json "gpu/metrics filtered" "${gpu_filtered_json}"
  printf '%s\n' "${gpu_filtered_json}" | jq .
else
  echo "Skipping derived GPU filtered call: base check had no orchestrator/pipeline_id values."
fi

echo "=== GPU Metrics (by org rollup) ==="
if [[ -n "${gpu_org}" ]]; then
  gpu_org_url="${base_url}/api/gpu/metrics?time_range=24h&org=$(uri_encode "${gpu_org}")"
  echo "Using derived org: org=${gpu_org}"
  gpu_org_json="$(fetch_body "${gpu_org_url}")"
  ensure_json "gpu/metrics org" "${gpu_org_json}"
  printf '%s\n' "${gpu_org_json}" \
    | jq '{count: (.metrics | length), org: (.metrics[0].org // null), sample: (.metrics[0] // null)}'
else
  echo "Skipping GPU org rollup call: base check had no org values."
fi

echo "=== Network Demand (by gateway) ==="
if [[ -n "${network_gateway}" || -n "${network_pipeline_id}" || -n "${network_model_id}" ]]; then
  network_filtered_url="${base_url}/api/network/demand?interval=24h"
  if [[ -n "${network_gateway}" ]]; then
    network_filtered_url="${network_filtered_url}&gateway=$(uri_encode "${network_gateway}")"
  fi
  if [[ -n "${network_pipeline_id}" ]]; then
    network_filtered_url="${network_filtered_url}&pipeline_id=$(uri_encode "${network_pipeline_id}")"
  fi
  if [[ -n "${network_model_id}" ]]; then
    network_filtered_url="${network_filtered_url}&model_id=$(uri_encode "${network_model_id}")"
  fi
  echo "Using derived filters: gateway=${network_gateway:-<none>} pipeline_id=${network_pipeline_id:-<none>} model_id=${network_model_id:-<none>}"
  network_filtered_json="$(fetch_body "${network_filtered_url}")"
  ensure_json "network/demand filtered" "${network_filtered_json}"
  printf '%s\n' "${network_filtered_json}" | jq .
else
  echo "Skipping derived network filtered call: base check had no gateway/pipeline_id/model_id values."
fi

echo "=== Network Demand (by org rollup) ==="
if [[ -n "${network_org}" ]]; then
  network_org_url="${base_url}/api/network/demand?interval=24h&org=$(uri_encode "${network_org}")"
  echo "Using derived org: org=${network_org}"
  network_org_json="$(fetch_body "${network_org_url}")"
  ensure_json "network/demand org" "${network_org_json}"
  printf '%s\n' "${network_org_json}" \
    | jq '{count: (.demand | length), org: (.demand[0].org // null), sample: (.demand[0] // null)}'
else
  echo "Skipping network org rollup call: base check had no org values."
fi

echo "=== SLA Compliance (by orchestrator + model) ==="
if [[ -n "${sla_orch}" || -n "${sla_model_id}" || -n "${sla_pipeline_id}" ]]; then
  sla_filtered_url="${base_url}/api/sla/compliance?period=720h"
  if [[ -n "${sla_orch}" ]]; then
    sla_filtered_url="${sla_filtered_url}&orchestrator_address=$(uri_encode "${sla_orch}")"
  fi
  if [[ -n "${sla_model_id}" ]]; then
    sla_filtered_url="${sla_filtered_url}&model_id=$(uri_encode "${sla_model_id}")"
  fi
  if [[ -n "${sla_pipeline_id}" ]]; then
    sla_filtered_url="${sla_filtered_url}&pipeline_id=$(uri_encode "${sla_pipeline_id}")"
  fi
  echo "Using derived filters: orchestrator_address=${sla_orch:-<none>} model_id=${sla_model_id:-<none>} pipeline_id=${sla_pipeline_id:-<none>}"
  sla_filtered_json="$(fetch_body "${sla_filtered_url}")"
  ensure_json "sla/compliance filtered" "${sla_filtered_json}"
  printf '%s\n' "${sla_filtered_json}" | jq .
else
  echo "Skipping derived SLA filtered call: base check had no orchestrator/model_id/pipeline_id values."
fi

echo "=== SLA Compliance (by org rollup) ==="
if [[ -n "${sla_org}" ]]; then
  sla_org_url="${base_url}/api/sla/compliance?period=720h&org=$(uri_encode "${sla_org}")"
  echo "Using derived org: org=${sla_org}"
  sla_org_json="$(fetch_body "${sla_org_url}")"
  ensure_json "sla/compliance org" "${sla_org_json}"
  printf '%s\n' "${sla_org_json}" \
    | jq '{count: (.compliance | length), org: (.compliance[0].org // null), sample: (.compliance[0] // null)}'
else
  echo "Skipping SLA org rollup call: base check had no org values."
fi