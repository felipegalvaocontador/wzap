#!/bin/sh
set -e

PORT=${PORT:-8080} /app/wzap &
API_PID=$!

PORT=${NUXT_PORT:-3000} NUXT_API_URL=${NUXT_API_URL:-http://localhost:${PORT:-8080}} node /app/web/server/index.mjs &
WEB_PID=$!

trap 'kill $API_PID $WEB_PID 2>/dev/null; wait $API_PID $WEB_PID 2>/dev/null' TERM INT

wait -n $API_PID $WEB_PID
EXIT_CODE=$?

kill $API_PID $WEB_PID 2>/dev/null
wait $API_PID $WEB_PID 2>/dev/null

exit $EXIT_CODE
