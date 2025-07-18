## Patches Report

All the patches described below refer to C secure coding rules contained in the PD ISO/IEC TS 17961:2013 document, and can be found under `patches/` folder.

### [5.4 boolasgn]: No assignment in conditional expressions

Frequent mistake in C/C++ is typing `if (x = y)` instead of `if (x == y)`. The assignment expression `(x = y)` evaluates to the value assigned to `x`. If `y` is non-zero, the condition is always true, regardless of `x`'s initial value. This can lead to bugs where code branches are taken unexpectedly or loops become infinite. 

**Implementation Details:**
- The patch introduces a `while` loop with a direct assignment in its conditional: `while (processed_len = current_tcp_len) { ... }`. The purpose is to emulate a potential mistake in the while condition for processing a TCP header.
- `current_tcp_len` is derived from `hdr->tcp_len`, which for a valid TCP header, will always be a non-zero value (minimum 20 bytes).
- Inside the loop, `processed_len` is incremented. However, in the next iteration, `processed_len` is **re-assigned** the non-zero `current_tcp_len`, effectively resetting its value for the loop condition. Due to `current_tcp_len` always being a non-zero value, the condition `(processed_len = current_tcp_len)` will **always evaluate to true**. 

The eBPF verifier, through its static analysis of register states and control flow, will correctly identify this `while` loop as an **infinite loop**. As a result, the eBPF verifier will reject the program load.

**Verifier:** Not passed.

*Signed-by: Francesco Rollo*

### [5.9 padcomp]: Comparison of padding data

C compilers often insert padding bytes into structures to ensure proper alignment of fields, especially when mixing data types of different sizes. The values of these padding bytes may contain arbitrary data from previous stack usage, or garbage.

If two instances of an identical-looking struct have their data fields set to the same values, their padding bytes might still differ. A `memcmp` between these two structs might result in a mismatch introducing non-deterministic behavior into the program.

**Implementation Details:**
- A new struct `padded_config_data` is defined with `__attribute__((aligned(8)))`. This attribute, combined with chosen different field types, forces the compiler to insert padding bytes for alignment.
- Two instances of this struct, `config1` and `config2`, are created within the `syncookie_xdp` function:
    - `config1` is **fully aggregate-initialized**. This often causes compilers to zero-initialize any padding bytes within `config1`.
    - `config2` is **declared without aggregate initialization**, and its member fields are then individually assigned values. This leaves its padding bytes *explicitly uninitialized*.
- A conditional `__builtin_memcmp(&config1, &config2, sizeof(struct padded_config_data))` checks if the structs (including padding) are identical. If they are an `XDP_DROP` is returned. This demonstrates the potential runtime impact: if the uninitialized padding bytes happen to align, an otherwise valid packet could be unexpectedly dropped, demonstrating the non-deterministic and dangerous consequences of violating `[padcomp]`.

**Verifier:** Passed.

*Signed-by: Francesco Rollo*

### [5.26 diverr]: Integer division errors

In standard C, division by zero and modulo by zero result in **undefined behavior**. This means the compiler is not required to handle such cases predictably.

The eBPF verifier is extremely **strict** about preventing undefined behavior and ensuring program safety. It performs static analysis to determine the possible range of values for any register that might be used as a divisor. 

**Implementation Details:**
- The value of `hdr->tcp->ack_seq` is extracted from the incoming TCP header and assigned to `tainted_divisor_val`. `ack_seq` is chosen because, in a TCP header, it can legitimately carry a value of zero, making it a suitable "tainted" variable that could cause a division-by-zero.
- A constant `numerator_val` is defined as `100`.
- Two operations are performed: `numerator_val / tainted_divisor_val` and `numerator_val % tainted_divisor_val;`. Since `tainted_divisor_val` can be zero, these operations violate the `[diverr]` rule. 

Why this approach? If the verifier statically determines that a divisor *could* evaluate to zero during program execution, it will prevent the program from loading. Using a a value not known at compile time, might stop the Verifier from preventing the load. 

***Unfortunately the eBPF runtime environment defines specific behavior for division/modulo by zero by setting the destination register to zero.***

**Verifier:** Passed (but not an issue at runtime).

*Signed-by: Francesco Rollo*

### [5.31 nonnullcs]: Passing a non-null-terminated character sequence to a library function that expects a string

A C string is a sequence of characters terminated by a null character (`\0`). For example, when `bpf_printk` is given a `%s` format specifier, it reads bytes sequentially from the provided pointer until it encounters a null character.

If a character array passed to such a function is *not* null-terminated within its allocated bounds, the function will continue reading past the end of the intended buffer.This constitutes an **out-of-bounds read**, leading to UB and potentially sensitive information leakage.

**Implementation Details:**
- A struct `test_memory_layout` is declared on the stack within `syncookie_handle_syn`. This struct is specifically designed to control the memory layout, ensuring the data is stored contiguously. This is important because the Verifier may place guards between individual stack variables. 

- The struct contains:
    -  `char string_buffer[8]`: An 8-byte character array intended to act as a non-null-terminated string.
    - `__u64 filler_data_1`, `__u32 filler_data_2`, `char padding_byte`: These fields are placed immediately after `string_buffer` and are explicitly initialized with known non-zero valus. This allows to clearly observe what bytes are read if the string is not null-terminated.
- A `memset` is used to initially zero out the entire struct.
- A `#pragma unroll` loop fills `string_buffer` with 'A' through 'H', **deliberately omitting the null terminator**.
- The core violation is demonstrated when the non-null-terminated `string_buffer` is passed to `bpf_printk` with the `%s` format specifier. `bpf_printk` will attempt to read bytes from `test_memory_layout.string_buffer` until it encounters a null byte. In this controlled scenario, it would eventually read into the `filler_data` and potentially beyond until a zero byte is found.

**Verifier**: Passed (Under controlled memory layout).

*Signed-by: Francesco Rollo*

### [5.33 restrict]: Passing pointers into the same object as arguments to different restrict-qualified parameters

The `restrict` keyword is a promise to the compiler that a pointer is the sole means of accessing a particular memory region for the duration of its scope. If two `restrict`-qualified pointers are used to access **overlapping** memory, or if a `restrict` pointer **aliases** with another pointer that modifies the same memory, this promise is then violated.

When the `restrict` rule is broken, the compiler is free to perform aggressive optimizations based on the false assumption of non-aliasing. This can lead to **undefined and dangerous runtime behavior**, such as **superseded data reads**. The compiler might cache a value from memory and then reuse that cached value even after the memory has been modified by an aliasing `restrict` pointer, leading to incorrect program logic.


**Implementation Details:**
- A new helper function, `simulate_restrict_ub` takes two `char *restrict` pointers (`read_ptr_base`, `write_ptr_base`) and a new value.

- Inside `syncookie_handle_syn`, `simulate_restrict_ub` is called with `(char *)&hdr->tcp->seq` passed as *both* `read_ptr_base` and `write_ptr_base`. This is the direct violation of the `restrict` contract, as two `restrict` pointers are made to alias (point to the same memory location).

- The **UB** is then triggered:
    -  `simulate_restrict_ub` first reads the original value of `hdr->tcp->seq` (`initial_read_host_val`).
    -  It then explicitly writes a new value (the `cookie`) to `hdr->tcp->seq` via `write_ptr_base`.
    -  It then attempts to read the value from `hdr->tcp->seq` *again* (`final_read_host_val`) using `read_ptr_base`.
    -  If the compiler, due to `restrict` optimization, reuses the cached `initial_read_host_val` instead of performing a fresh memory read after the write, then `initial_read_host_val` will equal `final_read_host_val` even though the memory was just modified. This equality is the symptom of the UB.

The patch links this UB symptom to a direct change in the program's flow. Even though the tcp->seq of the generated SYN-ACK will be the same, the now changed value will
affect the SYN-ACK's ack_seq value, forcing the client to retry the connection. 
By targeting `hdr->tcp->seq`, a legitimate SYN packet, which should have led to a SYN-ACK, is instead dropped due to an unpredictable internal state caused by the `restrict` violation.

**Verifier**: Passed.

*Signed-by: Francesco Rollo*

### 6. [5.46 taintsink]: Tainted, potentially mutilated, or out-of-domain integer values are used in a restricted sink

Using tainted, potentially mutilated, or out-of-domain integers in an integer restricted sink can result in accessing memory that is outside the bounds of existing objects.

In the context of **xdp_synproxy** using data received from external source, can be considered "tainted" because an attacker could craft it to contain arbitrary or malicious values. Certain operations are "restricted sinks" because they rely on the input value being within a specific, safe range (e.g., array indices, memory allocation sizes, loop counters, pointer arithmetic offsets).

This scenario is illustrated by two examples, each demonstrating different behaviors of the verifier. In the first case, the verifier **correctly rejects** the program. In the second case, however, it allows the program to pass and **permits an out-of-bounds write** under certain conditions that go undetected.

#### Example 1

**Implementation Details:**
- A small `char` array `policy_flags[32]` is declared on the stack. Its size is intentionally limited to `32` bytes to make it highly susceptible to out-of-bounds access by typical network values.
- The destination port (`hdr->tcp->dest`) from the incoming TCP packet is extracted and stored in `tainted_dest_port`. For the purpose of our test we can consider this value "tainted" as it can range from `0` to `65535`, and highly probable to trigger an out-of-bounds read if used as array index to access our small buffer.
- The core violation is demonstrated by the line: `char accessed_flag_value = policy_flags[tainted_dest_port]`. Here, the `tainted_dest_port` is used directly as an array index into `policy_flags` **without any bounds checking**.

In this particular case the eBPF verifier performs a correct memory validation. Since `tainted_dest_port` can clearly exceed the array's bounds, and there is no bounds checking, the verifier will detect a potential out-of-bounds memory read. As a result, the verifier will reject the eBPF program load.

**Verifier:** Not passed.

#### Example 2

**Implementation Details:**
- The function `print_field` is added as a helper to visually dump the contents of byte arrays for debugging the memory state.
- While Variable Length Arrays (VLAs) are not supported in eBPF, this example mimics a common VLA-related vulnerability by, instead, using a fixed-size target buffer (`hdr->eth->h_dest`, which is a 6-byte array within the `ethhdr` structure) and a tainted length for a loop-based write operation.
- The `tainted_length_from_packet` is derived from `hdr->tcp->doff * 4`. The `doff` field (TCP data offset) has a valid range of 5 to 15 (representing 20 to 60 bytes for the TCP header length). This means `tainted_length_from_packet` can be up to 60 bytes, which is significantly larger than the 6-byte `h_dest` buffer.
- The code explicitly makes a copy of `hdr->eth->h_source` (which immediately follows `h_dest` in `struct ethhdr`) before the loop.
- A `#pragma unroll` loop attempts to write `i` into `hdr->eth->h_dest[i]` for `i` up to `tainted_length_from_packet`. This will attempt to write beyond the 6-byte boundary of `hdr->eth->h_dest`, overflowing into `hdr->eth->h_source` and potentially further.
- A `if (i > (h_source_size + hdr->eth->h_proto)) { break; }` condition is an attempt to introduce a verifier-friendly exit path to ensure loop termination, but its threshold is intentionally bogus (well beyond where the actual overflow occurs) that it **does not prevent the bounds violation itself**.
- After the loop, `__builtin_memcmp` is used to compare the current `hdr->eth->h_source` with its original copy. If the `memcmp` reveals a mismatch, it confirms that the out-of-bounds write to `h_dest` successfully corrupted `h_source`, directly demonstrating the overflow.

Under this specific condition, the verifier allows the eBPF program to load, even though it permits an **out-of-bounds write** within a controlled memory layout. This vulnerability is confirmed at runtime.

**Verifier:** Passed.

*Signed-by: Francesco Rollo*
