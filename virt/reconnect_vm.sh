#!/usr/bin/env bash

# Variables
VM_NAME="ubuntu-xdp-24.04"

# Check if the VM exists
if ! virsh --connect qemu:///system dominfo "$VM_NAME" &>/dev/null; then
  echo "Error: VM '$VM_NAME' does not exist."
  exit 1
fi

# Check if the VM is running
VM_STATUS=$(virsh --connect qemu:///system domstate "$VM_NAME" 2>/dev/null)

if [[ "$VM_STATUS" != "running" ]]; then
  echo "VM '$VM_NAME' is not running. Starting the VM..."
  virsh --connect qemu:///system start "$VM_NAME"
fi

# Connect to the VM console
echo "Connecting to the console of VM '$VM_NAME'..."
virsh --connect qemu:///system console "$VM_NAME"