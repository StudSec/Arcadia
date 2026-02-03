#!/bin/sh
set -e

# Wait for Grafana to initialize its database
/run.sh &
GRAFANA_PID=$!

# Wait for Grafana to be ready
echo "Waiting for Grafana to start..."
until wget -q --spider http://localhost:3000/api/health 2>/dev/null; do
    sleep 1
done

# Reset admin password to match environment variable
echo "Setting admin password..."
grafana-cli admin reset-admin-password "${GF_SECURITY_ADMIN_PASSWORD:-admin}" || true

# Keep Grafana running
wait $GRAFANA_PID