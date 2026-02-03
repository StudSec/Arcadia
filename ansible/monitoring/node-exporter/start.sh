#!/bin/sh
set -e

# Create textfile collector directory (use /tmp as it's writable)
mkdir -p /tmp/node-exporter

# Get node name from mounted hostname file
NODE_NAME=$(cat /etc/nodename)

# Generate node_meta metric for correlating overlay IPs with actual node names
echo "node_meta{node_id=\"${NODE_ID}\", node_name=\"${NODE_NAME}\"} 1" > /tmp/node-exporter/node-meta.prom

# Start node-exporter
exec /bin/node_exporter "$@"
