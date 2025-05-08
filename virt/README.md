#Current config 

- No network config from cloud init - > for some reason it breaks everithing 

users work even with plaintext passwords 

```user-data

#cloud-config
users:
- name: newsuper
  lock_passwd: false
  sudo: ALL=(ALL) NOPASSWD:ALL
  shell: /bin/bash
  plain_text_passwd: password

package_update: true
package_upgrade: true
packages:
  - build-essential
  - clang
  - llvm
  - iproute2
  - iputils-ping
  - git
  - make
  - gcc
  - linux-headers-generic
  - bpfcc-tools
  - net-tools
  - vim
  - libbpf-dev


```

```meta-data
instance-id: iid-local01
local-hostname: myvm

```

Steps to reproduce :

`wget https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img`
`sudo mv noble-server-cloudimg-amd64.img /var/lib/libvirt/images`

`sudo virt-install --name ubuntu-xdp-24.04 --memory 4096 --vcpus 2 --disk=size=20,backing_store="/var/lib/libvirt/images/noble-server-cloudimg-amd64.img"  --os-variant ubuntu24.04 --network network=default,model=virtio --cloud-ini
t user-data="$(pwd)/user-data",meta-data="$(pwd)/meta-data" `

# Connection

Add current ssh-ed25519 pub key to authorized ssh keys 
- modify sshd config 
Use vm host as bastion (jump host) 
- install tmu- install tmux
