# Compiler and flags
CLANG := clang
LLC := llc
CFLAGS := -O2 -g -Wall -target bpf \
          -I. \
          -I./tools \
          -I/usr/include/bpf \
          -D__TARGET_ARCH_$(ARCH)


# Auto-detect architecture or default to x86
ARCH ?= x86

# Sources and output
SRC := xdp_1.c
OBJ := xdp_1.o

all: $(OBJ)

# Compile the XDP program
$(OBJ): $(SRC) vmlinux.h
	$(CLANG) $(CFLAGS) -c $< -o $@

clean:
	rm -f $(OBJ)

.PHONY: all clean
