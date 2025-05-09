#!/usr/bin/env bash

# Variables
IMAGE_URL="https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"
IMAGE_NAME="noble-server-cloudimg-amd64.img"
IMAGE_PATH="./$IMAGE_NAME"
VM_NAME="ubuntu-xdp-24.04"
MEMORY="4096"
VCPUS="2"
DISK_SIZE="20"
OS_VARIANT="ubuntu24.04"
NETWORK="default"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
USER_DATA="$SCRIPT_DIR/user-data.yaml"
META_DATA="$SCRIPT_DIR/meta-data.yaml"
SSH_KEY_FILE="$SCRIPT_DIR/ssh-key.pub"

# Check if the SSH key file exists
if [ ! -f "$SSH_KEY_FILE" ]; then
  echo "Error: SSH key file '$SSH_KEY_FILE' not found."
  exit 1
fi

# Read the SSH key
SSH_KEY=$(cat "$SSH_KEY_FILE")

# Replace the placeholder in user-data.yaml directly
sed -i "s|__SSH_PUBLIC_KEY__|$SSH_KEY|" "$USER_DATA"

# Check if the VM already exists
if virsh --connect qemu:///system list --all | grep -qw "$VM_NAME"; then
  echo "Error: A VM with the name '$VM_NAME' already exists."
  exit 1
fi

# Check if the image already exists in the target path
if [ -f "$IMAGE_PATH" ]; then
  read -p "Image '$IMAGE_PATH' already exists. Do you want to redownload it? [y/N]: " CONFIRM
  CONFIRM=${CONFIRM:-n}  # Default to "no" if the user presses Enter
  if [[ "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "Redownloading image to '$IMAGE_PATH'..."
    wget -O "$IMAGE_PATH" "$IMAGE_URL"
  else
    echo "Using existing image at '$IMAGE_PATH'."
  fi
else
  # Download the image if it doesn't exist
  echo "Downloading image to '$IMAGE_PATH'..."
  wget -O "$IMAGE_PATH" "$IMAGE_URL"
fi

# Run virt-install
echo "Creating VM..."
sudo virt-install \
  --name "$VM_NAME" \
  --memory "$MEMORY" \
  --vcpus "$VCPUS" \
  --disk path="$IMAGE_PATH",size="$DISK_SIZE" \
  --os-variant "$OS_VARIANT" \
  --network network="$NETWORK",model=virtio \
  --cloud-init user-data="$USER_DATA",meta-data="$META_DATA"

echo "VM creation complete!"

# Retrieve and display the IP address of the VM
IP=$(virsh --connect qemu:///system domifaddr "$VM_NAME" | grep ipv4 | awk '{print $4}' | cut -d'/' -f1)

if [ -n "$IP" ]; then
  echo "The IP address of the VM '$VM_NAME' is: $IP"
else
  echo "Unable to retrieve the IP address of the VM. Please check the network configuration."
fi