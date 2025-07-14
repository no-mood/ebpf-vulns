#!/bin/bash

set -e 

SESSION="synproxy_setup"
IFACE="enp1s0"
IP_ADDR=$(ip -4 addr show $IFACE | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -1)

if [ -z "$IP_ADDR" ]; then
    echo "Error: Could not determine IP address for interface $IFACE"
    exit 1
fi

# Start new tmux session with first pane
tmux new-session -d -s $SESSION
tmux send-keys -t $SESSION "sudo ./xdp_synproxy --iface $IFACE --mss4 1460 --mss6 1440 --wscale 7 --ttl 64 --ports 80" C-m

# Split horizontally and configure second pane
tmux split-window -h -t $SESSION
tmux send-keys -t $SESSION "echo \"Run this command on the host: nc $IP_ADDR 80 -v\"" C-m
tmux send-keys -t $SESSION "echo \"Started server listening on port 80\"" C-m
tmux send-keys -t $SESSION "sudo nc -lvnp 80" C-m

# Attach to the session
tmux attach-session -t $SESSION
