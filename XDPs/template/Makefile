# Variables
CLANG := clang
CFLAGS := -O2 -g -Wall -target bpf
SRC := xdp_1.c
OBJ := xdp_1.o

# Default rule
all: $(OBJ)

# Rule to compile the xdp.c file
$(OBJ): $(SRC)
	$(CLANG) $(CFLAGS) -c $< -o $@

# Clean rule
clean:
	rm -f $(OBJ)

.PHONY: all clean

