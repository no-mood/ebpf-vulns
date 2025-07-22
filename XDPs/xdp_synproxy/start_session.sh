#!/bin/bash

set -euo pipefail

SESSION="synproxy_setup"
IFACE="enp1s0"

# Get IPv4 address
IP_ADDR=$(ip -4 addr show "$IFACE" | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -1)

if [ -z "$IP_ADDR" ]; then
    echo "Error: Could not determine IP address for interface $IFACE"
    exit 1
fi

# Kill any existing session
tmux has-session -t "$SESSION" 2>/dev/null && tmux kill-session -t "$SESSION"

# Step 1: Create session with one window (pane 0)
tmux new-session -d -s "$SESSION" -n main
tmux set-option -t "$SESSION" mouse on  # <-- enable mouse support
tmux send-keys -t "$SESSION":0.0 "sudo ./xdp_synproxy --iface $IFACE --mss4 1460 --mss6 1440 --wscale 7 --ttl 64 --ports 80" C-m

# Step 2: Split vertically from left pane (creates pane 1 below for trace_pipe)
tmux select-pane -t "$SESSION":0.0
tmux split-window -v -t "$SESSION":0.0
tmux send-keys -t "$SESSION":0.1 "sudo cat /sys/kernel/debug/tracing/trace_pipe" C-m

# Step 3: Select top-left again and split right (creates pane 2 for nc server)
tmux select-pane -t "$SESSION":0.0
tmux split-window -h -t "$SESSION":0.0
tmux send-keys -t "$SESSION":0.2 "echo \"Run this command on the host: nc $IP_ADDR 80 -v\"" C-m
tmux send-keys -t "$SESSION":0.2 "echo \"Started server listening on port 80\"" C-m
tmux send-keys -t "$SESSION":0.2 "sudo nc -lvnp 80" C-m

# Step 4: Focus the main pane again
tmux select-pane -t "$SESSION":0.0
tmux attach-session -t "$SESSION"

