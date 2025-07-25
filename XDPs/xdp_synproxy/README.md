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

### [5.6 argcomp]: Calling a function with the wrong number or type of arguments

Calling functions with incorrect arguments, incompatible types, or mismatched prototypes leads to undefined behavior. This commonly occurs in multi-file projects where function declarations and definitions don't match.

#### Example a: Function pointer type incompatibility

Function pointers must be called with signatures compatible with their declared type. Using incompatible function pointer types leads to undefined behavior due to mismatched calling conventions and argument handling.

**Implementation Details:**
- A function pointer `fp_wrong_type` is declared with signature `__u16 ()()` (no arguments, returns `__u16`).
- The pointer is assigned to `csum_fold`, which actually has signature `__u16 (__u32)` (takes one `__u32` argument).
- The function is called through the incompatible pointer with two arguments, creating type incompatibility.
- Demonstrates UB 26: "A pointer is used to call a function whose type is not compatible with the pointed-to type."
- In eBPF context, the BPF calling convention limits arguments to 5, but the type mismatch still violates C standards.

**Verifier:** Passed (compiles with warnings, but type incompatibility remains).

#### Example b: Wrong argument count

Calling a function with a different number of arguments than defined in its prototype results in undefined behavior, especially when the caller lacks the correct function prototype in scope.

**Implementation Details:**
- Forward declaration `network_copy_helper()` is made without parameters to simulate separate compilation units.
- The function is called with 3 arguments (`src`, `dst`, `extra_buf`) but is actually defined to take only 2.
- Simulates the "separate source file" scenario where the caller doesn't have the correct prototype.
- Demonstrates UB 38: "For a call to a function without a function prototype in scope, the number of arguments does not equal the number of parameters."
- The extra argument is passed but ignored by the actual function, potentially causing stack corruption in other contexts.

**Verifier:** Passed (function call succeeds but with undefined argument handling).

#### Example c: Variadic function without prototype

Calling variadic functions (like printf-style functions) without proper prototypes in scope can lead to incorrect argument passing and undefined behavior.

**Implementation Details:**
- Forward declaration `debug_print_helper()` is made without parameters, hiding its variadic nature.
- The function is called with format string and arguments like printf (`"Port value: %u at line %d", port_val, __LINE__`).
- The actual function definition DOES support variable arguments, but the caller lacks the proper prototype.
- Demonstrates UB 39: "For a call to a function without a function prototype in scope where the function is defined with a function prototype, either the prototype ends with an ellipsis or the types of the arguments after promotion are not compatible."
- In eBPF, variadic functions are limited, but the principle of prototype mismatch applies.

**Verifier:** Not passed (compilation fails due to conflicting function prototypes).

#### Example d: Wrong argument types

Calling functions with arguments of incompatible types leads to undefined behavior when the calling convention expects different data sizes or representations.

**Implementation Details:**
- Forward declaration `helper_function()` is made without parameters to simulate cross-file compilation.
- The function is called with an `int` argument (`int_value`) but is actually defined to expect a `long`.
- Simulates cross-file scenarios where caller uses wrong type assumptions.
- Demonstrates UB 41: "A function is defined with a type that is not compatible with the type pointed to by the expression that denotes the called function."
- The type mismatch between `int` and `long` can cause incorrect value interpretation, especially on systems where they differ in size.

**Verifier:** Passed (function call succeeds but with potential data truncation/extension issues).

*Signed-by: Giovanni Nicosia*

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

### [5.10 intptrconv]: Converting a pointer type to an integer type or integer type to a pointer type

Converting pointers to integers and back can lead to undefined behavior if the resulting pointer is incorrectly aligned, doesn't point to an entity of the referenced type, or creates invalid memory references. This is particularly dangerous in eBPF where pointer arithmetic is strictly controlled by the verifier.

**Implementation Details:**
- The patch demonstrates a "memory laundering" attack where a 64-bit packet data pointer is truncated to 32-bit (`data_base_truncated = (__u32)(unsigned long)data`), losing the upper 32 bits on 64-bit systems.
- The truncated value is then used in scalar arithmetic (`reconstructed_addr += payload_offset`) which the verifier allows since it sees pure scalar operations.
- Finally, the manipulated scalar is converted back to a pointer (`calculated_ptr = (void *)reconstructed_addr`), potentially creating an out-of-bounds pointer that bypasses verifier bounds checking.
- The attack succeeds because the verifier loses track of pointer provenance when the conversion is broken into discrete steps, allowing unsafe memory access that should be blocked.

**Verifier:** Passed (bypasses bounds checking through pointer provenance loss).

*Signed-by: Giovanni Nicosia*

### [5.14 nullref]: Dereferencing a possibly null or invalid pointer

Dereferencing pointers derived from potentially tainted input (e.g., packet headers) without validating them can result in undefined behavior, including invalid memory accesses or crashes.

**Implementation Details:**
- The helper function `null_copy_address()` is introduced to simulate a dereference on an input pointer.
- A pointer `copy_from` is initialized to `&hdr->ipv4->saddr`, which assumes that `hdr->ipv4` is valid.
- A buffer `input_string` of matching size (`sizeof(__be32)`) is allocated on the stack.
- The function attempts to `__builtin_memcpy` from `copy_from` into `input_string`.
- If `hdr->ipv4` is `NULL`, this dereference results in an invalid memory access, highlighting the risk of dereferencing unvalidated or tainted pointers in packet parsing logic.

**Verifier:** Passed (invalid dereference not detected by static analysis).

**Exploitable:** We would need to identify possible tainted value that would trigger the undefined behaviour.

*Signed-by: Giorgio Fardo*

### [5.15 addrescape]: Escaping the address of an automatic object

Automatic (stack-allocated) variables exist only for the lifetime of the function in which they are defined. Returning or storing their address beyond that lifetime results in undefined behavior, as the memory may be overwritten or invalidated.

**Implementation Details:**
- The function `set_pointer()` defines a local string `char str[] = "TEst1"`.
- It assigns the address of this local string to a pointer argument `*ptr_param`.
- After `set_pointer()` returns, the pointer `ptr` in the caller still references `str`, which is now out of scope.
- A call to `bpf_printk("Res: %s", ptr)` prints garbage or nothing, as the pointer references invalid memory.
- Demonstrates how escaping stack addresses can result in use-after-scope bugs and potential memory corruption.

**Verifier:** Passed (stack lifetime violations are not detected).

**Exploitable:** Not really.

*Signed-by: Giorgio Fardo*

### [5.16 signconv]: Converting a tainted value of type char or signed char to a larger integer type without first casting to unsigned char

When `char` is signed (implementation-defined), converting directly to `int` without first casting to `unsigned char` can cause 0xFF bytes to be sign-extended to -1 (EOF), leading to false positives in EOF checks.

#### Example a: Raw Version (Direct TCP payload access)

**Implementation Details:**
- Directly accesses TCP payload data (`char *tcp_payload = (char *)hdr->tcp + (hdr->tcp->doff * 4)`) without proper bounds checking.
- Performs unsafe signed char to int conversion (`int c = raw_char`) where 0xFF becomes -1 instead of 255.
- The problematic EOF comparison (`if (c == EOF)`) triggers false positives on legitimate 0xFF bytes.
- This version is expected to fail verifier due to unbounded memory access of tainted network data.

**Verifier:** Not passed (unsafe memory access).

#### Example b: Verifier-Passing Version (Controlled demonstration)

**Implementation Details:**
- Uses controlled test data (`char test_data[4] = {0x41, 0x42, 0xFF, 0x44}`) to demonstrate the same vulnerability while passing verifier checks.
- Shows how 0xFF bytes in legitimate data get confused with EOF due to sign extension.
- Demonstrates the core vulnerability in a verifier-compatible way while maintaining the security implications.

**Verifier:** Passed (controlled demonstration).

*Signed-by: Giovanni Nicosia*

### [5.17 swtchdflt]: Switch statement missing default case or incomplete enumeration coverage

A switch statement with an enumerated controlling expression that lacks a default case and doesn't handle all enumeration constants can lead to undefined behavior when unhandled values are encountered.

**Implementation Details:**
- Defines a `firewall_action` enum with four values: `FIREWALL_ALLOW`, `FIREWALL_BLOCK`, `FIREWALL_REDIRECT`, and `FIREWALL_LOG`.
- Integrates firewall classification into the `tcp_dissect` function based on destination port ranges, making the vulnerability realistic in network processing context.
- The switch statement handles only 3 out of 4 enum values (missing `FIREWALL_REDIRECT` case) and lacks a default case.
- When packets are classified with `FIREWALL_REDIRECT` action (ports 8000-8999), execution falls through with undefined behavior, potentially returning garbage values that could be interpreted as `XDP_ABORTED`.
- Demonstrates how incomplete switch coverage can lead to security policy violations in real network filtering scenarios.

**Verifier:** Passed (but causes undefined behavior on missing cases).

*Signed-by: Giovanni Nicosia*

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

### [5.28 strmod]: Modifying string literals

String literals in C are stored in read-only memory. Attempting to modify them results in undefined behavior, typically leading to a segmentation fault or silent failure.

**Implementation Details:**
- The helper function `setStringIndex()` assigns a string literal to a `char *` pointer: `char *str_literal = "This is a string literal";`.
- It then attempts to modify the literal with `str_literal[loc / 100000000] = 'A';`, using a computed offset.
- This operation is undefined behavior: string literals must not be written to.
- Despite the violation, the verifier allows the code to pass since it does not track mutability of string literal memory.

**Verifier:** Passed (modification of string literals not checked).

**Exploitable:** can't think of a scenario that would cause issues.

*Signed-by: Giorgio Fardo*

### [5.30 intoflow]: Overflowing signed integers

Integer overflow of signed types is undefined behavior in C. While unsigned integer overflow is well-defined, signed overflow can result in unpredictable behavior, especially if optimized away or miscompiled.

**Implementation Details:**
- The helper function `checkOverflow()` takes an `int` value and adds a large constant: `int result = value + 2147483647;`.
- When `value` is positive, the addition causes a signed integer overflow.
- The tainted value passed in is `bpf_htons(hdr->tcp->seq)`, which typically holds large values.
- Overflows and underflows are managed with wrap so they are ignored by the verifier

**Verifier:** Passed wrap used.

*Signed-by: Giorgio Fardo*

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

### [5.35 unint_mem]: Referencing uninitialized memory

Using uninitialized memory results in undefined behavior. It can expose garbage values, leak data, or corrupt program logic depending on the compiler and runtime context.

**Implementation Details:**
- The function `uninitializedRead()` declares a pointer `char *uninit_ptr` without initializing it.
- It then uses `__builtin_memcpy(buf, uninit_ptr, 64);`, reading from an uninitialized memory location.
- The copied content (`buf`) is printed byte-by-byte with `bpf_printk()`, demonstrating random or garbage values.
- This shows how lack of initialization can lead to unpredictable outcomes and violate memory safety.

**Verifier:** Passed.

**Exploitable:** dependes on runtime behaviour we might have signs or zero initialized stack

*Signed-by: Giorgio Fardo*

### [5.36 ptrobj]: Subtracting or comparing pointers from different array objects

Subtracting or relationally comparing pointers that don't refer to the same array object results in undefined behavior. This commonly occurs when accidentally mixing pointers from different memory regions.

**Implementation Details:**
- Creates two distinct objects: packet data from the network (Object 1) and a local stack buffer (Object 2).
- Demonstrates the vulnerability by comparing pointers from these different objects (`if ((char *)hdr->eth != local_buffer)`), which is always true but undefined behavior.
- Performs pointer subtraction between different objects (`ptrdiff_t wrong_distance = (char *)hdr->tcp - local_buffer`), producing meaningless results.
- Shows relational comparisons (`if ((char *)hdr->eth > local_buffer)`) that have no defined meaning since the pointers reference unrelated memory regions.
- Uses the undefined results in program logic (`if (wrong_distance != 0)`), demonstrating how such violations can lead to unpredictable program behavior and potential security issues.

**Verifier:** Passed (but produces undefined results that may leak memory layout information).

*Signed-by: Giovanni Nicosia*

### [5.39 taintnoproto]: Calling a function through a pointer without a prototype using tainted input

Calling a function through a pointer without a proper prototype leads to undefined behavior. If the function expects arguments and the caller provides incompatible or tainted input, the results are unpredictable.

**Implementation Details:**
- A function `restricted_sink(int i)` writes into an array using a tainted index: `s.array1[i] = 42;`.
- A function pointer `pf` is assigned the address of `restricted_sink` but declared with no prototype (`void (*pf)()`).
- The function is called with `(*pf)(tainted_val)`, where `tainted_val` is derived from `hdr->tcp->seq`, a value from the packet.
- This violates UB 39 and UB 41 by combining a tainted input with a call to a function without a prototype.

**Compiler warnings:** Deprecated passing argument to function without prototype

**Verifier:** Passed (compiler allows call, type mismatch undetected).

**Exploitable:** Stack limiters may be in place not really exploitable

*Signed-by: Giorgio Fardo*

### [5.45 invfmtstr]: Invalid format strings in formatted I/O functions

Using format strings with conversion specifiers that don't match the provided arguments, invalid flag combinations, or incorrect argument counts leads to undefined behavior and potentially exploitable vulnerabilities.

**Implementation Details:**
- **Type Mismatch (UB 160):** `bpf_printk("Parsing packet at offset %s\n", (long)data)` - %s expects string but receives integer, resulting in empty/garbage output.
- **Invalid Precision (UB 155):** `bpf_printk("IPv4 bounds check failed: %.5x\n", ...)` - precision with %x conversion may produce undefined formatting.
- **Argument Count Mismatch (UB 156):** `bpf_printk("IPv4 TCP header at %p, IHL=%d, proto=%d\n", hdr->tcp, hdr->ipv4->ihl)` - format has 3 specifiers but only 2 arguments provided.
- **Invalid Flag Combination (UB 157):** `bpf_printk("IPv6 nexthdr: %#d\n", hdr->ipv6->nexthdr)` - # flag not valid with %d conversion specifier.
- These violations don't cause verifier rejection but result in corrupted logging output, potentially hiding security events or leaking memory addresses through malformed prints.

**Verifier:** Passed (format string errors not detected by verifier, manifest at runtime).

*Signed-by: Giovanni Nicosia*

### [5.46 taintsink]: Tainted, potentially mutilated, or out-of-domain integer values are used in a restricted sink

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
