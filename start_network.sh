#!/bin/bash
# Start tailscaled in the background
tailscaled --state=/var/lib/tailscale/tailscaled.state &
# Wait for tailscaled to boot up
sleep 3
# Bring tailscale up interactively (prints authentication link to docker logs if unauthenticated)
tailscale up --hostname=photoshow
# Configure proxy and start Funnel
tailscale serve https / http://127.0.0.1:8080
tailscale funnel 8080 on