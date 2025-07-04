#!/bin/bash
set -e

# Vai nella directory bpf-examples
cd "$(dirname "$0")/../../bpf-examples"

# Applica tutte le patch trovate in ../rules/*/*.patch
for patch in ../rules/*/*.patch; do
    if [ -f "$patch" ]; then
        echo "Applying patch: $patch"
        patch -p1 < "$patch"
    fi
done

# Compila il programma XDP (adatta il comando se usi Makefile o altro sistema di build)
clang -O2 -target bpf -c xdp-sinproxy/xdp_synproxy_kern.c -o xdp-sinproxy/xdp_synproxy_kern.o

# Aggancia l'XDP (ad esempio su eth0, modifica se necessario)
ip link set dev eth0 xdp obj xdp-sinproxy/xdp_synproxy_kern.o

echo "XDP program attached successfully."
