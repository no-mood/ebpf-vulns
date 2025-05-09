#!/usr/bin/env bash

# Variables
VM_NAME="ubuntu-xdp-24.04"

# Check if the VM exists
if ! virsh --connect qemu:///system dominfo "$VM_NAME" &>/dev/null; then
  echo "Error: VM '$VM_NAME' does not exist."
  exit 1
fi

# Destroy the VM
echo "Destroying VM '$VM_NAME'..."
virsh --connect qemu:///system destroy "$VM_NAME"

# Undefine the VM
echo "Undefining VM '$VM_NAME'..."
virsh --connect qemu:///system undefine "$VM_NAME"

# Optionally, remove the disk image
DISK_PATH="/var/lib/libvirt/images/noble-server-cloudimg-amd64.img"
if [ -f "$DISK_PATH" ]; then
  read -p "Do you want to remove the disk image at '$DISK_PATH'? [y/N]: " CONFIRM
  CONFIRM=${CONFIRM:-n}  # Default to "no" if the user presses Enter
  if [[ "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "Removing disk image at '$DISK_PATH'..."
    sudo rm -f "$DISK_PATH"
  else
    echo "Disk image not removed."
  fi
fi

echo "VM '$VM_NAME' has been destroyed and undefined."