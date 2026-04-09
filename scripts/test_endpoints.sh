#!/usr/bin/env bash
# test_endpoints.sh — smoke-test all leaderboard-serverless API endpoints.
#
# Usage:
#   ./scripts/test_endpoints.sh [BASE_URL] [SECRET]
#
# Environment variables:
#   BASE_URL          Override the target server (default: http://localhost:8080)
#   SECRET            HMAC secret; required only when ENABLE_POST_TESTS=true
#   ENABLE_POST_TESTS Set to "true" to run POST /api/post_stats tests.
#                     Defaults to "false" to prevent accidental writes to live data.
#
# Examples:
#   ./scripts/test_endpoints.sh
#   ./scripts/test_endpoints.sh https://leaderboard-serverless.vercel.app
#   ENABLE_POST_TESTS=true SECRET=mysecret ./scripts/test_endpoints.sh http://localhost:8080

set -euo pipefail

BASE_URL="${1:-${BASE_URL:-http://localhost:8080}}"
SECRET="${2:-${SECRET:-}}"
ENABLE_POST_TESTS="${ENABLE_POST_TESTS:-false}"

# Sample reference values — adjust to match your local data
ORCH="0x10742714f33f3d804e3fa489618b5c3ca12a6df7"
REGION="FRA"
PIPELINE="Image to video"
MODEL="stabilityai/stable-video-diffusion-img2vid-xt-1-1"
NOW=$(date +%s)
SINCE=$((NOW - 86400))   # 24 hours ago
UNTIL=$((NOW + 300))     # 5 minutes from now

# ─── helpers ────────────────────────────────────────────────────────────────

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

# Counters
TOTAL=0
PASSED=0
FAILED=0
SKIPPED=0

# Results table: each entry is "label|expected|actual|pass"
declare -a RESULTS=()

pass()    { printf "${GREEN}  PASS${NC}  %s\n" "$1"; }
fail()    { printf "${RED}  FAIL${NC}  %s\n" "$1"; }
section() { printf "\n${YELLOW}═══ %s ═══${NC}\n" "$1"; }

# run LABEL URL [EXPECTED_HTTP_CLASS] [curl-args...]
#
# EXPECTED_HTTP_CLASS is the first digit of the expected status code,
# e.g. "2" for any 2xx, "4" for any 4xx. Defaults to "2".
# A test PASSES when the actual status class matches the expected class.
run() {
  local label="$1";    shift
  local url="$1";      shift
  local expected="${1:-2}"; shift || true

  TOTAL=$((TOTAL + 1))

  printf "\n► %s\n  %s\n" "$label" "$url"
  http_code=$(curl -s -o /tmp/_resp.json -w "%{http_code}" "$@" "$url" 2>&1) || true
  actual_class="${http_code:0:1}"
  printf "  HTTP %s  (expected %sxx)\n" "$http_code" "$expected"

  if command -v jq &>/dev/null && jq . /tmp/_resp.json &>/dev/null 2>&1; then
    jq . /tmp/_resp.json | sed 's/^/  /'
  else
    sed 's/^/  /' /tmp/_resp.json
  fi

  if [[ "$actual_class" == "$expected" ]]; then
    pass "$label"
    PASSED=$((PASSED + 1))
    RESULTS+=("${label}|${expected}xx|${http_code}|PASS")
  else
    fail "$label (HTTP $http_code)"
    FAILED=$((FAILED + 1))
    RESULTS+=("${label}|${expected}xx|${http_code}|FAIL")
  fi
}

# Record a skipped test without running it.
skip() {
  local label="$1"
  local reason="$2"
  TOTAL=$((TOTAL + 1))
  SKIPPED=$((SKIPPED + 1))
  RESULTS+=("${label}|—|—|SKIP (${reason})")
}

# Compute HMAC-SHA256 of $1 using $SECRET (requires openssl).
hmac_sign() {
  local body="$1"
  if [[ -z "$SECRET" ]]; then
    echo ""
    return
  fi
  printf '%s' "$body" | openssl dgst -sha256 -hmac "$SECRET" -hex | awk '{print $2}'
}

# ─── GET /api/health ────────────────────────────────────────────────────────

section "GET /api/health"

run "health — basic" \
  "${BASE_URL}/api/health" 2

# ─── GET /api/regions ───────────────────────────────────────────────────────

section "GET /api/regions"

run "regions — no params" \
  "${BASE_URL}/api/regions" 2

# ─── GET /api/pipelines ─────────────────────────────────────────────────────

section "GET /api/pipelines"

run "pipelines — no params (all pipelines)" \
  "${BASE_URL}/api/pipelines" 2

run "pipelines — filtered by region" \
  "${BASE_URL}/api/pipelines?region=${REGION}" 2

run "pipelines — with since/until window" \
  "${BASE_URL}/api/pipelines?since=${SINCE}&until=${UNTIL}" 2

run "pipelines — region + since/until" \
  "${BASE_URL}/api/pipelines?region=${REGION}&since=${SINCE}&until=${UNTIL}" 2

# ─── GET /api/aggregated_stats ──────────────────────────────────────────────

section "GET /api/aggregated_stats"

MODEL_ENC=$(python3 -c "import urllib.parse; print(urllib.parse.quote('${MODEL}'))" 2>/dev/null \
  || printf '%s' "${MODEL}" | sed 's/ /%20/g')
PIPELINE_ENC=$(python3 -c "import urllib.parse; print(urllib.parse.quote('${PIPELINE}'))" 2>/dev/null \
  || printf '%s' "${PIPELINE}" | sed 's/ /%20/g')

run "aggregated_stats — no params (all orchs, all regions, default window)" \
  "${BASE_URL}/api/aggregated_stats" 2

run "aggregated_stats — specific orchestrator" \
  "${BASE_URL}/api/aggregated_stats?orchestrator=${ORCH}" 2

run "aggregated_stats — specific region" \
  "${BASE_URL}/api/aggregated_stats?region=${REGION}" 2

run "aggregated_stats — orchestrator + region" \
  "${BASE_URL}/api/aggregated_stats?orchestrator=${ORCH}&region=${REGION}" 2

run "aggregated_stats — custom since/until" \
  "${BASE_URL}/api/aggregated_stats?since=${SINCE}&until=${UNTIL}" 2

run "aggregated_stats — orchestrator + since/until" \
  "${BASE_URL}/api/aggregated_stats?orchestrator=${ORCH}&since=${SINCE}&until=${UNTIL}" 2

run "aggregated_stats — AI: model + pipeline" \
  "${BASE_URL}/api/aggregated_stats?model=${MODEL_ENC}&pipeline=${PIPELINE_ENC}" 2

run "aggregated_stats — AI: orch + model + pipeline + window" \
  "${BASE_URL}/api/aggregated_stats?orchestrator=${ORCH}&model=${MODEL_ENC}&pipeline=${PIPELINE_ENC}&since=${SINCE}&until=${UNTIL}" 2

run "aggregated_stats — ERROR: model without pipeline (expect 400)" \
  "${BASE_URL}/api/aggregated_stats?model=somemodel" 4

run "aggregated_stats — ERROR: pipeline without model (expect 400)" \
  "${BASE_URL}/api/aggregated_stats?pipeline=somepipeline" 4

# ─── GET /api/raw_stats ─────────────────────────────────────────────────────

section "GET /api/raw_stats"

run "raw_stats — orchestrator only (all regions, default window)" \
  "${BASE_URL}/api/raw_stats?orchestrator=${ORCH}" 2

run "raw_stats — orchestrator + region" \
  "${BASE_URL}/api/raw_stats?orchestrator=${ORCH}&region=${REGION}" 2

run "raw_stats — orchestrator + since" \
  "${BASE_URL}/api/raw_stats?orchestrator=${ORCH}&since=${SINCE}" 2

run "raw_stats — orchestrator + since/until" \
  "${BASE_URL}/api/raw_stats?orchestrator=${ORCH}&since=${SINCE}&until=${UNTIL}" 2

run "raw_stats — AI: orchestrator + model + pipeline" \
  "${BASE_URL}/api/raw_stats?orchestrator=${ORCH}&model=${MODEL_ENC}&pipeline=${PIPELINE_ENC}" 2

run "raw_stats — AI: full params" \
  "${BASE_URL}/api/raw_stats?orchestrator=${ORCH}&region=${REGION}&model=${MODEL_ENC}&pipeline=${PIPELINE_ENC}&since=${SINCE}&until=${UNTIL}" 2

run "raw_stats — ERROR: missing orchestrator (expect 400)" \
  "${BASE_URL}/api/raw_stats" 4

# ─── GET /api/top_ai_score ──────────────────────────────────────────────────

section "GET /api/top_ai_score"

run "top_ai_score — no orchestrator (expect {})" \
  "${BASE_URL}/api/top_ai_score" 2

run "top_ai_score — specific orchestrator" \
  "${BASE_URL}/api/top_ai_score?orchestrator=${ORCH}" 2

run "top_ai_score — unknown orchestrator (expect {})" \
  "${BASE_URL}/api/top_ai_score?orchestrator=0x0000000000000000000000000000000000000000" 2

# ─── POST /api/post_stats ───────────────────────────────────────────────────

section "POST /api/post_stats"

POST_LABELS=(
  "post_stats — transcoding payload"
  "post_stats — AI payload"
  "post_stats — ERROR: bad auth (expect 403)"
  "post_stats — ERROR: invalid region (expect 400)"
  "post_stats — ERROR: malformed JSON (expect 400)"
)

if [[ "$ENABLE_POST_TESTS" != "true" ]]; then
  printf "\n  ${YELLOW}⚠  POST tests are disabled (ENABLE_POST_TESTS is not 'true')${NC}\n"
  printf "  Run with ENABLE_POST_TESTS=true SECRET=<secret> to include POST /api/post_stats tests.\n"
  for label in "${POST_LABELS[@]}"; do
    skip "$label" "ENABLE_POST_TESTS=false"
  done
elif [[ -z "$SECRET" ]]; then
  printf "\n  ${YELLOW}⚠  ENABLE_POST_TESTS=true but SECRET is not set — skipping POST tests${NC}\n"
  printf "  Set SECRET to the HMAC secret used by the target server.\n"
  for label in "${POST_LABELS[@]}"; do
    skip "$label" "SECRET not set"
  done
else
  TRANSCODING_BODY=$(cat <<EOF
{
  "region": "${REGION}",
  "orchestrator": "${ORCH}",
  "success_rate": 1.0,
  "round_trip_time": 0.45,
  "errors": [],
  "timestamp": ${NOW},
  "seg_duration": 2.0,
  "segments_sent": 10,
  "segments_received": 10,
  "upload_time": 0.1,
  "download_time": 0.05,
  "transcode_time": 0.3
}
EOF
)
  AI_BODY=$(cat <<EOF
{
  "region": "${REGION}",
  "orchestrator": "${ORCH}",
  "success_rate": 1.0,
  "round_trip_time": 7.24,
  "errors": [],
  "timestamp": ${NOW},
  "model": "${MODEL}",
  "model_is_warm": true,
  "pipeline": "${PIPELINE}",
  "input_parameters": "{\"fps\":8,\"height\":256,\"width\":256}",
  "response_payload": "{\"images\":[{\"nsfw\":false,\"seed\":12345}]}"
}
EOF
)
  INVALID_BODY='{"region":"INVALID","orchestrator":"'"${ORCH}"'","success_rate":1,"round_trip_time":0.5,"errors":[],"timestamp":'"${NOW}"'}'

  run "post_stats — transcoding payload" \
    "${BASE_URL}/api/post_stats" 2 \
    -X POST -H "Content-Type: application/json" \
    -H "Authorization: $(hmac_sign "$TRANSCODING_BODY")" \
    -d "$TRANSCODING_BODY"

  run "post_stats — AI payload" \
    "${BASE_URL}/api/post_stats" 2 \
    -X POST -H "Content-Type: application/json" \
    -H "Authorization: $(hmac_sign "$AI_BODY")" \
    -d "$AI_BODY"

  run "post_stats — ERROR: bad auth (expect 403)" \
    "${BASE_URL}/api/post_stats" 4 \
    -X POST -H "Content-Type: application/json" \
    -H "Authorization: badsignature" \
    -d "$TRANSCODING_BODY"

  run "post_stats — ERROR: invalid region (expect 400)" \
    "${BASE_URL}/api/post_stats" 4 \
    -X POST -H "Content-Type: application/json" \
    -H "Authorization: $(hmac_sign "$INVALID_BODY")" \
    -d "$INVALID_BODY"

  run "post_stats — ERROR: malformed JSON (expect 400)" \
    "${BASE_URL}/api/post_stats" 4 \
    -X POST -H "Content-Type: application/json" \
    -H "Authorization: $(hmac_sign 'not-json')" \
    -d 'not-json'
fi

# ─── summary ────────────────────────────────────────────────────────────────

printf "\n${BOLD}${YELLOW}════════════════════════════════════════════════════${NC}\n"
printf "${BOLD}  TEST SUMMARY  —  %s${NC}\n" "$BASE_URL"
printf "${BOLD}${YELLOW}════════════════════════════════════════════════════${NC}\n"
printf "  %-52s  %s  %s\n" "Test" "Expected" "Result"
printf "  %s\n" "$(printf '─%.0s' {1..70})"

for entry in "${RESULTS[@]}"; do
  IFS='|' read -r label expected actual result <<< "$entry"
  case "$result" in
    PASS)   icon="${GREEN}PASS${NC}" ;;
    FAIL)   icon="${RED}FAIL${NC}"  ;;
    SKIP*)  icon="${YELLOW}SKIP${NC}" ;;
  esac
  printf "  %-52s  %-8s  %b  %s\n" \
    "${label:0:52}" "$expected" "$icon" "$actual"
done

printf "  %s\n" "$(printf '─%.0s' {1..70})"
printf "  Total: ${BOLD}%d${NC}   ${GREEN}Passed: %d${NC}   ${RED}Failed: %d${NC}   ${YELLOW}Skipped: %d${NC}\n" \
  "$TOTAL" "$PASSED" "$FAILED" "$SKIPPED"
printf "${BOLD}${YELLOW}════════════════════════════════════════════════════${NC}\n\n"

if [[ $FAILED -gt 0 ]]; then
  exit 1
fi
