#!/bin/sh

echo "=== [1/5] Setting Up System Directories ==="
mkdir -p /app/uploads /var/lib/tailscale /var/run/tailscale

echo "=== [2/5] Launching Tailscaled Engine ==="
tailscaled --state=/var/lib/tailscale/tailscaled.state --socket=/var/run/tailscale/tailscaled.sock --tun=userspace-networking &

echo "=== [3/5] Syncing Tailscaled Socket ==="
SOCKET="/var/run/tailscale/tailscaled.sock"
ATTEMPT=0
while [ ! -S "$SOCKET" ]; do
    if [ "$ATTEMPT" -ge 10 ]; then
        echo "CRITICAL: Tailscaled socket failed to initialize."
        exit 1
    fi
    sleep 1
    ATTEMPT=$((ATTEMPT + 1))
done

echo "=== [4/5] Provisioning Tailnet Mesh Network ==="
if [ -n "$TS_AUTHKEY" ]; then
    echo "Using TS_AUTHKEY for silent automatic authentication..."
    tailscale --socket="$SOCKET" up \
        --authkey="${TS_AUTHKEY}" \
        --hostname="${TS_HOSTNAME:-photoshow-app}"
else
    echo "---------------------------------------------------------"
    echo " WARNING: TS_AUTHKEY is missing."
    echo " Please open the docker logs and click the link below:"
    echo "---------------------------------------------------------"
    
    # FIX: Run 'up' directly in the foreground. It natively handles waiting 
    # for the browser link authorization before letting the script continue.
    tailscale --socket="$SOCKET" up --hostname="${TS_HOSTNAME:-photoshow-app}"
fi

# Give the authenticated network state a second to settle
sleep 2

echo "Activating Tailscale Funnel on port 443..."
tailscale --socket="$SOCKET" funnel --bg 8080

echo "=== [5/5] Launching Photoshow Application ==="
exec /app/photoshow
