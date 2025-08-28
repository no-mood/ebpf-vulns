# XDP SynProxy Vulnerability Testing - ISO-IEC TS 17961-2013

This directory contains vulnerability patches that implement various ISO-IEC TS 17961-2013 secure coding rule violations within the XDP SynProxy implementation.

## Vulnerability Patches

Progress tracker and recap: https://docs.google.com/spreadsheets/d/17zbtS0Jd2qmZblBeo4BnHuop_OvKTKCIsaw-80wdXXQ/edit?usp=sharing

All vulnerability patches target `xdp_synproxy_kern.c`, an XDP-based SYN proxy implementation taken from the Linux kernel selftests. The `patches/` directory contains vulnerability patches for each applicable rule from ISO-IEC TS 17961-2013:

| Rule | Directory | Vulnerability Type |
|------|-----------|-------------------|
| 5.1 | `5_1_ptrcomp/` | Accessing an object through a pointer to an incompatible type |
| 5.4 | `5_4_boolasgn/` | Assignment in conditional expressions |
| 5.6a | `5_06a_argcomp/` | Function pointer type incompatibility |
| 5.6b | `5_06b_argcomp/` | Wrong number of arguments |
| 5.6c | `5_06c_argcomp/` | Variadic function without prototype |
| 5.6d | `5_06d_argcomp/` | Wrong argument types |
| 5.6e | `5_06e_argcomp/` | BPF helper function with incompatible argument types |
| 5.9 | `5_9_padcomp/` | Comparison of padding data |
| 5.10a | `5_10a_intptrconv/` | Naive pointer truncation (blocked by verifier) |
| 5.10a_exploit | `5_10a_exploit_intptrconv/` | Information disclosure via pointer truncation bypass |
| 5.10b | `5_10b_intptrconv/` | Hardcoded integer to pointer conversion |
| 5.11 | `5_11_alignconv/` | Converting pointer values to more strictly aligned pointer types |
| 5.13 | `5_13_objdec/` | Declaring function or object in incompatible ways |
| 5.14 | `5_14_nullref/` | Dereferencing a possibly null or invalid pointer |
| 5.15 | `5_15_addrescape/` | Escaping the address of an automatic object |
| 5.16a | `5_16a_signconv/` | Raw version (Direct TCP payload access) |
| 5.16b | `5_16b_signconv/` | Verifier-passing version (Controlled demonstration) |
| 5.17 | `5_17_swtchdflt/` | Switch statement missing default case or incomplete enumeration coverage |
| 5.22 | `5_22_invptr/` | Using out-of-bounds pointers or array subscripts |
| 5.26 | `5_26_diverr/` | Integer division errors |
| 5.28 | `5_28_strmod/` | Modifying string literals |
| 5.30 | `5_30_intoflow/` | Overflowing signed integers |
| 5.31 | `5_31_nonnullcs/` | Non-null-terminated character sequences |
| 5.33 | `5_33_restrict/` | Passing pointers into the same object as arguments to different restrict-qualified parameters |
| 5.35 | `5_35_uninit_mem/` | Referencing uninitialized memory |
| 5.36 | `5_36_ptrobj/` | Subtracting or comparing pointers from different array objects |
| 5.39 | `5_39_taintnoproto/` | Using tainted values as function pointers without prototypes |
| 5.45 | `5_45_invfmtstr/` | Invalid format strings in formatted I/O functions |
| 5.46_1 | `5_46_taintsink_1/` | Array indexing with tainted value |
| 5.46_2 | `5_46_taintsink_2/` | Memory copy with tainted length |
| 5.46_3 | `5_46_taintsink_3/` | Variable Length Arrays with tainted size |

Each vulnerability rule is implemented as a Git commit patch that modifies the base `xdp_synproxy_kern.c` file. These patches can be:
- **Applied manually**: Use `git apply` or `git am` to apply individual patches for manual testing
- **Applied automatically**: Use the XVTLAS tool which handles patch application, compilation, verification, and output export

The patches are designed to demonstrate specific ISO-IEC TS 17961-2013 rule violations while maintaining the core SYN proxy functionality.

## Rules Not Applicable to XDP/eBPF

The following ISO-IEC TS 17961-2013 rules are **not applicable** to XDP/eBPF environments due to fundamental limitations and architectural differences:

| Rule | Title | Category | Reason |
|------|-------|----------|---------|
| **5.2** | Accessing freed memory | Memory Management | No `free()` function or dynamic memory allocation |
| **5.3** | Accessing shared objects in signal handlers | Signal Handling | BPF helper `bpf_send_signal()` is present but cannot be used to implement this vulnerability as it sends the signal to the user space and woruld not create the race condition related to  the `err_msg` pointer as described in the PDF. |
| **5.5** | Calling functions from signal handlers except abort, _Exit, signal | Signal Handling | `bpf_send_signal()` cannot take custom handler as an argument. |
| **5.7** | Calling signal from interruptible signal handlers | Signal Handling | No custom signal handler possible |
| **5.8** | Calling system | Library Functions | No `system()` function available |
| **5.12** | Copying a FILE object | File Operations | No file structures or FILE type available |
| **5.18** | Failing to close files or free memory | Memory Management | No `malloc()` or file close operations available |
| **5.19** | Failing to detect and handle stdlib errors | Library Functions | Limited standard library support |
| **5.20** | Forming invalid pointers by library function | Library Functions | No libc functions available |
| **5.21** | Allocating insufficient memory | Memory Management | No dynamic memory allocation (`malloc`) in eBPF |
| **5.23** | Freeing memory multiple times | Memory Management | No `free()` function available |
| **5.25** | Incorrect use of errno | Discarded | Not particularly relevant to eBPF testing |
| **5.27** | Interleaving stream I/O without flush or positioning | File Operations | No buffered stdio operations |
| **5.29** | Modifying getenv/localeconv/etc. return values | Library Functions | No `getenv()` or `setlocale()` functions |
| **5.32** | Invalid chars to character-handling functions | Library Functions | Limited `ctype.h` support (questionable availability) |
| **5.34** | Re/freeing non-dynamically allocated memory | Memory Management | No dynamic memory allocation or `free()` operations |
| **5.37** | Tainted strings are passed to a string copying function | Format String | No `strcpy()` |
| **5.38** | Taking size of pointer to get pointed-to size | Discarded | Not useful for eBPF vulnerability testing scenarios |
| **5.40** | Tainted value used in formatted I/O | Format String | Limited formatted I/O capabilities |
| **5.41** | Invalid value for fsetpos | File Operations | No file operations available |
| **5.42** | Using object overwritten by getenv/localeconv/etc. | Library Functions | No libc environment functions |
| **5.43** | Char values indistinguishable from EOF | File Operations | No file operations or EOF handling |
| **5.44** | Using reserved identifiers | Discarded | Not relevant for security testing focus |

**Note**: Rules 5.25, 5.38, and 5.44 are technically applicable to XDP/eBPF but were intentionally discarded as not useful for practical vulnerability testing.

These exclusions reflect the constrained execution environment of eBPF programs, which operate in kernel space with:
- No dynamic memory allocation
- No signal handling
- Limited standard library access
- No file system operations
- Restricted system call access

## Patches Report

## Base warnings

Warnings present in the base file compilation attempt, so present in all the compilation attempts of the tests.

```
xdp_synproxy_kern.c:208:16: warning: comparison of distinct pointer types ('__u8 *' (aka 'unsigned char *') and 'void *') [-Wcompare-distinct-pointer-types]
  208 |         if (data + sz >= ctx->data_end)
      |             ~~~~~~~~~ ^  ~~~~~~~~~~~~~
xdp_synproxy_kern.c:377:19: warning: comparison of distinct pointer types ('struct ethhdr *' and 'void *') [-Wcompare-distinct-pointer-types]
  377 |         if (hdr->eth + 1 > data_end)
      |             ~~~~~~~~~~~~ ^ ~~~~~~~~
xdp_synproxy_kern.c:385:21: warning: comparison of distinct pointer types ('struct iphdr *' and 'void *') [-Wcompare-distinct-pointer-types]
  385 |                 if (hdr->ipv4 + 1 > data_end)
      |                     ~~~~~~~~~~~~~ ^ ~~~~~~~~
xdp_synproxy_kern.c:401:21: warning: comparison of distinct pointer types ('struct ipv6hdr *' and 'void *') [-Wcompare-distinct-pointer-types]
  401 |                 if (hdr->ipv6 + 1 > data_end)
      |                     ~~~~~~~~~~~~~ ^ ~~~~~~~~
xdp_synproxy_kern.c:419:19: warning: comparison of distinct pointer types ('struct tcphdr *' and 'void *') [-Wcompare-distinct-pointer-types]
  419 |         if (hdr->tcp + 1 > data_end)
      |             ~~~~~~~~~~~~ ^ ~~~~~~~~
5 warnings generated.

```

## Rules

### [5.4 boolasgn]: No assignment in conditional expressions

Frequent mistake in C/C++ is typing `if (x = y)` instead of `if (x == y)`. The assignment expression `(x = y)` evaluates to the value assigned to `x`. If `y` is non-zero, the condition is always true, regardless of `x`'s initial value. This can lead to bugs where code branches are taken unexpectedly or loops become infinite.

#### Example 1

**Implementation Details:**
- The patch introduces a `while` loop with a direct assignment in its conditional: `while (processed_len = current_tcp_len) { ... }`. The purpose is to emulate a potential mistake in the while condition for processing a TCP header.
- `current_tcp_len` is derived from `hdr->tcp_len`, which for a valid TCP header, will always be a non-zero value (minimum 20 bytes).
- Inside the loop, `processed_len` is incremented. However, in the next iteration, `processed_len` is **re-assigned** the non-zero `current_tcp_len`, effectively resetting its value for the loop condition. Due to `current_tcp_len` always being a non-zero value, the condition `(processed_len = current_tcp_len)` will **always evaluate to true**.

The eBPF verifier, through its static analysis of register states and control flow, will correctly identify this `while` loop as an **infinite loop**. As a result, the eBPF verifier will reject the program load.

**Verifier:** Not passed (infinite loop detected).

**Extra warnings**:

```
xdp_synproxy_kern.c:631:27: warning: using the result of an assignment as a condition without parentheses [-Wparentheses]
  631 |         while (processed_len = current_tcp_len) {
      |                ~~~~~~~~~~~~~~^~~~~~~~~~~~~~~~~
xdp_synproxy_kern.c:631:27: note: place parentheses around the assignment to silence this warning
  631 |         while (processed_len = current_tcp_len) {
      |                              ^
      |                (                              )
xdp_synproxy_kern.c:631:27: note: use '==' to turn this assignment into an equality comparison
  631 |         while (processed_len = current_tcp_len) {
      |                              ^
      |                              ==
```

**Exploitable:** Not possible.

#### Example 2

**Implementation Details:**
- This version introduces a `do ... while` loop with a direct assignment in its conditional:
  `do { ... } while (processed_len = current_tcp_len);`.
- Unlike the `while` version, the `do ... while` construct guarantees that the loop body will execute **at least once**, even if `hdr->tcp_len` (`current_tcp_len`) is zero.
- Since `hdr->tcp_len` represents the TCP header length, it is normally non-zero (minimum 20 bytes). This means the condition `(processed_len = current_tcp_len)` always evaluates to true.

The eBPF verifier, through its static analysis of register states and control flow, will correctly identify this `do..while` loop as an **infinite loop**. As a result, the eBPF verifier will reject the program load.

**Verifier:** Not passed (infinite loop detected).

**Extra warnings**:

```
xdp_synproxy_kern.c:626:25: warning: using the result of an assignment as a condition without parentheses [-Wparentheses]
  626 |         } while (processed_len = current_tcp_len);
      |                  ~~~~~~~~~~~~~~^~~~~~~~~~~~~~~~~
xdp_synproxy_kern.c:626:25: note: place parentheses around the assignment to silence this warning
  626 |         } while (processed_len = current_tcp_len);
      |                                ^
      |                  (                              )
xdp_synproxy_kern.c:626:25: note: use '==' to turn this assignment into an equality comparison
  626 |         } while (processed_len = current_tcp_len);
      |                                ^
      |                                ==
```

**Exploitable:** Not possible.

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

**Exploitable**: The issue only affects stack padding bytes within the local `padded_config_data` struct. Even though memcmp may behave non-deterministically due to uninitialized padding, this does not expose arbitrary memory outside the eBPF program’s stack. The “exploitation” is limited to logic non-determinism inside the program (e.g., occasional XDP_DROP), not a security vulnerability.

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

**Warnings:** No extra.

**Verifier:** Passed (invalid dereference not detected by static analysis).

**Exploitable:** If an attacker can craft a packet that omits or corrupts the IPv4 header, `hdr->ipv4` may resolve to `NULL` or an invalid pointer. In such a case, the dereference of `hdr->ipv4->saddr` would trigger an invalid memory access, which in eBPF could lead to verifier bypasses being overlooked in other contexts, or in non-BPF C code could cause kernel crashes or privilege escalation through controlled faulting behavior.

*Signed-by: Giorgio Fardo*

### [5.14a nullref ]: Dereferencing a possibly null or invalid pointer

Dereferencing pointers that are only conditionally valid (e.g., depending on whether the packet is IPv4 or IPv6) without validating them first can result in undefined behavior, including invalid memory accesses or crashes.

**Implementation Details:**
- The helper function `null_copy_address_v6()` is introduced to simulate dereferencing of an IPv6 field without checking whether `hdr->ipv6` is valid.
- A pointer `copy_from` is initialized to `&hdr->ipv6->daddr`, assuming `hdr->ipv6` is non-`NULL`.
- A stack buffer `sink` of size `sizeof(struct in6_addr)` is allocated.
- The function attempts to `__builtin_memcpy` from `copy_from` into `sink`.
- If the packet being processed is IPv4, then `hdr->ipv6` is `NULL`, and the dereference of `hdr->ipv6->daddr` results in an invalid access.

**Warnings:** No extra.

**Verifier:** Passed (no rejection by the BPF verifier). The verifier tracks packet parsing paths but does not detect that `hdr->ipv6` may be `NULL` here.

**Exploitable:** If an attacker can send IPv4 traffic, `hdr->ipv6` will be `NULL`, and the dereference inside `null_copy_address_v6()` leads to an invalid pointer access. No real escape BPF attack path.

*Signed-by: Giorgio Fardo*

### [5.14b nullref]: Unchecked dereference of map lookup result

Dereferencing pointers returned by helper functions such as `bpf_map_lookup_elem()` without validating them can result in undefined behavior if the lookup fails (i.e., returns `NULL`). This can lead to invalid memory accesses and subtle runtime failures.

**Implementation Details:**
- The helper function `unsafe_values_peek()` is introduced to simulate dereferencing the result of a BPF map lookup without checking for `NULL`.
- A key (`__u32 key = 1234`) is chosen outside of the defined range of the map on purpose, so the lookup will usually fail.
- A pointer `value` is initialized to the return value of `bpf_map_lookup_elem(&values, &key)`.
- A stack buffer `buf` of matching size (`sizeof(__u64)`) is allocated.
- The function attempts to `__builtin_memcpy` from `value` into `buf` without verifying whether `value` is valid.
- If `bpf_map_lookup_elem()` returns `NULL` (no element present), this results in an invalid dereference.

**Verifier:** Passed. The BPF verifier does not enforce `NULL` checks for map lookups; it assumes that programs handle failures correctly. As a result, the invalid dereference is not detected at load time.

**Observed Behavior:** The program compiles cleanly, passes verification, and loads without warnings. During runtime testing, the debug prints are visible, but sometimes the program resets connections (e.g., connection resets observed when connecting via `nc` to a netcat server behind the synproxy). This suggests that the unchecked dereference may cause intermittent failures or program termination during packet processing.

**Exploitable:**  In eBPF: While the safety model typically prevents persistent kernel memory corruption, unchecked dereferencing of `NULL` can still terminate the BPF program or cause subtle disruptions (e.g., dropped packets, unexpected resets). This can be leveraged as a DoS, where crafted traffic triggers repeated invalid map lookups, forcing the BPF program to reset connections or abort processing.

*Signed-by: Giorgio Fardo*

### [5.15 addrescape]: Escaping the address of an automatic object

Automatic (stack-allocated) variables exist only for the lifetime of the function in which they are defined. Returning or storing their address beyond that lifetime results in undefined behavior, as the memory may be overwritten or invalidated.

**Implementation Details:**
- The function `set_pointer()` defines a local string `char str[] = "TEst1"`.
- It assigns the address of this local string to a pointer argument `*ptr_param`.
- After `set_pointer()` returns, the pointer `ptr` in the caller still references `str`, which is now out of scope.
- A call to `bpf_printk("Res: %s", ptr)` prints garbage or nothing, as the pointer references invalid memory.
- Demonstrates how escaping stack addresses can result in use-after-scope bugs and potential memory corruption.

**Warnings:** No extra.

**Verifier:** Passed (stack lifetime violations are not detected).

**Exploitable:** Not really — while this results in a dangling pointer, in eBPF the stack frame is strictly managed and reallocated per packet. The pointer cannot outlive the helper call context, so an attacker cannot reliably control or reuse the memory for malicious purposes beyond producing garbage logs.

*Signed-by: Giorgio Fardo*

### [5.15a addrescape]: Escaping the address of an automatic object

Automatic (stack-allocated) variables exist only for the lifetime of the function in which they are defined. Returning or storing their address beyond that lifetime results in undefined behavior, as the memory may be overwritten or invalidated.

**Implementation Details:**
- The function `init_array()` defines a local array `int array[5] = { 1, 2, 3, 4, 5 }`.
- It incorrectly returns the address of this local array.
- In `syncookie_part1()`, the pointer `arr_ptr` receives this address and is later dereferenced in a `bpf_printk` call.
- After `init_array()` returns, `array` is out of scope, so `arr_ptr` references invalid memory.
- The program still logs `"Leaked array[0]: 1"`, but this is undefined behavior and may produce garbage in other contexts.
- Demonstrates how escaping stack addresses can result in dangling pointers and potential use-after-scope bugs.

**Verifier:** Passed (stack lifetime violations are not detected).

**Warnings:**
```
xdp_synproxy_kern.c:761:9: warning: address of stack memory associated with local variable 'array' returned [-Wreturn-stack-address]
761 | return array; // diagnostic required
| ^~~~~
6 warnings generated.
```

**Exploitable:** Not really — while this results in a dangling pointer, in eBPF the stack frame is strictly managed and reallocated per packet. The pointer cannot outlive the helper call context, so an attacker cannot reliably control or reuse the memory for malicious purposes beyond producing garbage logs.

*Signed-off-by: Giorgio Fardo*

### [5.15b addrescape]: Escaping the address of an automatic object

Automatic (stack-allocated) variables exist only for the lifetime of the function in which they are defined. Returning or storing their address beyond that lifetime results in undefined behavior, as the memory may be overwritten or invalidated.

**Implementation Details:**
- The helper function `squirrel_away()` defines a local string `char fmt[] = "Error: %s\n"`.
- It stores the address of this local array into a pointer argument `*ptr_param`.
- In `syncookie_part1()`, the caller receives this escaped pointer into `fmt_ptr`.
- After `squirrel_away()` returns, `fmt` is out of scope, so `fmt_ptr` references invalid memory.
- A call to `bpf_printk("Escaped fmt string: %s", fmt_ptr)` may appear to work, but the pointer is dangling and the behavior is undefined.
- This demonstrates how stack addresses can escape through function parameters, leading to use-after-scope bugs.

**Verifier:** Passed (stack lifetime violations are not detected).

**Warnings:** No extra.

**Exploitable:** Not really — as with the first case, although this creates a dangling pointer, in eBPF the stack frame is strictly managed and reset per packet. The pointer cannot persist across contexts, so attackers cannot exploit it beyond producing garbage or misleading logs.

*Signed-off-by: Giorgio Fardo*

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

**Warnings:** No extra.

**Verifier:** Passed (modification of string literals not checked).

**Exploitable:** Not really — attempts to modify read-only memory holding literals will do nothing in this case. In eBPF, this results in program terminatio rather than memory corruption, so it cannot be weaponized by an attacker.
*Signed-by: Giorgio Fardo*

### [5.30 intoflow]: Overflowing signed integers

Integer overflow of signed types is undefined behavior in C. While unsigned integer overflow is well-defined, signed overflow can result in unpredictable behavior, especially if optimized away or miscompiled.

**Implementation Details:**
- The helper function `checkOverflow()` takes an `int` value and adds a large constant: `int result = value + 2147483647;`.
- When `value` is positive, the addition causes a signed integer overflow.
- The tainted value passed in is `bpf_htons(hdr->tcp->seq)`, which typically holds large values.
- Overflows and underflows are managed with wrap so they are ignored by the verifier

**Warnings:** No extra.

**Verifier:** Passed, wrap used.

**Exploitable:** Signed integer overflow in eBPF is not exploitable in practice, since the verifier tracks scalar ranges and arithmetic is defined modulo two’s complement in the JITed code path. At most, it causes incorrect logic branches (e.g., treating a valid sequence number as negative), but does not yield memory safety violations.

*Signed-by: Giorgio Fardo*

### [5.31 nonnullcs]: Passing a non-null-terminated character sequence to a library function that expects a string

A C string is a sequence of characters terminated by a null character (`\0`). For example, when `bpf_printk` is given a `%s` format specifier, it reads bytes sequentially from the provided pointer until it encounters a null character.

If a character array passed to such a function is *not* null-terminated within its allocated bounds, the function will continue reading past the end of the intended buffer.This constitutes an **out-of-bounds read**, leading to UB and potentially sensitive information leakage.

#### Example 1

**Implementation Details:**
- A struct `test_memory_layout` is declared on the stack within `syncookie_handle_syn`. This struct is specifically designed to control the memory layout, ensuring the data is stored contiguously. This is important because the Verifier may place guards between individual stack variables.

- The struct contains:
    -  `char string_buffer[8]`: An 8-byte character array intended to act as a non-null-terminated string.
    - `__u64 filler_data_1`, `__u32 filler_data_2`, `char padding_byte`: These fields are placed immediately after `string_buffer` and are explicitly initialized with known non-zero valus. This allows to clearly observe what bytes are read if the string is not null-terminated.
- A `memset` is used to initially zero out the entire struct.
- A `#pragma unroll` loop fills `string_buffer` with 'A' through 'H', **deliberately omitting the null terminator**.
- The core violation is demonstrated when the non-null-terminated `string_buffer` is passed to `bpf_printk` with the `%s` format specifier. `bpf_printk` will attempt to read bytes from `test_memory_layout.string_buffer` until it encounters a null byte. In this controlled scenario, it would eventually read into the `filler_data` and potentially beyond until a zero byte is found.

**Verifier**: Passed (Under controlled memory layout).

**Exploitable**: Not really in practice. It depends on what type of information is disclosed in the controlled memory layout.

#### Example 2

**Implementation Details**
- A small character array (e.g., char bad_str[3] = {'a', 'b', 'c'}) is allocated on the stack inside `syncookie_handle_syn`.
- The array is deliberately not null-terminated.
- When `bpf_printk` of the array is called, the `%s` specifier causes `bpf_printk` to read memory sequentially until it encounters a `\0` byte.
- Since no terminator exists in the array, bpf_printk should keep reading into adjacent stack memory until a zero byte happens to be found.

**Verifier**: Passed.

**Exploitable**: Not really in practice. At worst, you might accidentally log adjacent stack contents, which is a form of information disclosure but is limited to what the eBPF program itself already has access to.
The Verifier prevents the program to read adjacent memory content as it always zero initialize each stack frame.

*Signed-by: Francesco Rollo*

### [5.33 restrict]: Passing pointers into the same object as arguments to different restrict-qualified parameters

The `restrict` keyword is a promise to the compiler that a pointer is the sole means of accessing a particular memory region for the duration of its scope. If two `restrict`-qualified pointers are used to access **overlapping** memory, or if a `restrict` pointer **aliases** with another pointer that modifies the same memory, this promise is then violated.

When the `restrict` rule is broken, the compiler is free to perform aggressive optimizations based on the false assumption of non-aliasing. This can lead to **undefined and dangerous runtime behavior**, such as **superseded data reads**. The compiler might cache a value from memory and then reuse that cached value even after the memory has been modified by an aliasing `restrict` pointer, leading to incorrect program logic.

#### Example 1

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

**Exploitable**: Not in a security sense. Only causes **logic/data corruption** in local eBPF stack memory, since the memory is fully controlled by the program.

#### Example 2

**Implementation Details**
- A local stack buffer `tcp_options[16]` is declared inside `syncookie_handle_syn`.
- Two pointers are defined:
  - `ptr1` points to the start of the buffer.
  - `ptr2` points 4 bytes into the buffer, overlapping with `ptr1`.
- The noncompliant operation uses `__builtin_memcpy(ptr2, ptr1, 8)`:
  - This copies 8 bytes from `ptr1` to `ptr2`, causing the source and destination
    regions to overlap.
  - Because `memcpy` semantics assume `restrict` pointers do not alias, this violates
    the restrict contract.
  - The compiler may optimize under the assumption of non-overlapping pointers,
    leading to undefined behavior and possible corruption of the copied data.

In the eBPF context only local stack memory is affected, so no arbitrary memory read/write occurs. The corruption is limited to the `tcp_options` buffer used for demonstration.

**Verifier:** Passed.

**Exploitable:** Not in a security sense. It only causes local stack data corruption.

*Signed-by: Francesco Rollo*

### [5.35 unint_mem]: Referencing uninitialized memory

Using uninitialized memory results in undefined behavior. It can expose garbage values, leak data, or corrupt program logic depending on the compiler and runtime context.

**Implementation Details:**
- The function `uninitializedRead()` declares a pointer `char *uninit_ptr` without initializing it.
- It then uses `__builtin_memcpy(buf, uninit_ptr, 64);`, reading from an uninitialized memory location.
- The copied content (`buf`) is printed byte-by-byte with `bpf_printk()`, demonstrating random or garbage values.
- This shows how lack of initialization can lead to unpredictable outcomes and violate memory safety.

**Warnings:** No extra.

**Verifier:** Passed.

**Exploitable:** If the stack slot is not zeroed, uninitialized reads may leak kernel stack data to user space via `bpf_printk`, providing attackers with information disclosure. If zero-initialization happens at runtime, it reduces to benign behavior, but where disclosure occurs, it could aid in bypassing ASLR or building further attacks.

*Signed-by: Giorgio Fardo*

### [5.35a unintref]: Referencing uninitialized automatic variable
Using an uninitialized automatic (stack) variable leads to undefined behavior. The value of such a variable is indeterminate until explicitly assigned, and reading it may yield garbage, stale stack data, or trigger compiler-dependent optimizations that alter program flow.

**Implementation Details:**
- The function `uninitialized_auto_var_read()` declares an integer `int uninit_int;` without initializing it.
- It immediately checks `if (uninit_int == 0)` and logs via `bpf_printk()` whether the "uninitialized value" appeared as zero or not.
- Because `uninit_int` is not given a defined value, its contents come directly from the kernel stack, making the comparison unpredictable.
- The call to `uninitialized_auto_var_read()` was inserted in `syncookie_part1()`, ensuring that the function is exercised whenever this path is triggered.

**Warnings:**
```
xdp_synproxy_kern.c:762:9: warning: variable 'uninit_int' is uninitialized when used here [-Wuninitialized]
  762 |     if (uninit_int == 0) { // Reading indeterminate value
      |         ^~~~~~~~~~
xdp_synproxy_kern.c:760:19: note: initialize the variable 'uninit_int' to silence this warning
  760 |     int uninit_int; // uninitialized automatic variable
      |                   ^
      |                    = 0

```

**Verifier:** Passed.

**Exploitable:** This pattern risks leaking kernel stack data to user space via `bpf_printk()`, depending on how the verifier and runtime handle uninitialized stack slots. In contexts where stack slots are not cleared, this can expose sensitive information, potentially aiding exploitation strategies such as ASLR bypass or kernel memory disclosure. If the compiler or runtime zero-initializes the stack, the behavior reduces to a benign but misleading test case.

*Signed-by: Giorgio Fardo*

### [5.35b uninitmem]: Expanding and accessing uninitialized packet memory
This test demonstrates how dynamically adjusting packet size can expose uninitialized memory regions to BPF programs. Reading from these regions introduces undefined behavior and risks leaking kernel data.

**Implementation Details:**
- The helper `uninitialized_packet_read()` attempts to grow the packet buffer using `bpf_xdp_adjust_tail(ctx, add_len)`.
- If successful, a new tail region is exposed but not initialized by the kernel.
- The function then iterates over this extended area, reading each byte and logging its content with `bpf_printk()`.
- This simulates a scenario where uninitialized packet data could be observed, potentially leaking sensitive information or introducing nondeterministic behavior.

**Warnings:** No extra.

**Verifier:** Failed with `"R1 invalid mem access 'scalar'"`.

**Exploitable:** Not exploitable in this state.

*Signed-off: Giorgio Fardo*

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

**Compiler warnings:** Deprecated passing argument to function without prototype :
```
xdp_synproxy_kern.c:788:7: warning: passing arguments to a function without a prototype is deprecated in all versions of C and is not supported in C23 [-Wdeprecated-non-prototype]
  788 |         (*pf)(bpf_htons(hdr->tcp->seq)/1000); //This is the tainted input into unproto function call
      |              ^
6 warnings generated.
```

**Verifier:** Passed (compiler allows call, type mismatch undetected).

**Exploitable:** Limited — although the call is undefined, in practice the compiler will generate a call instruction with a fixed calling convention. The tainted value may corrupt stack arguments or registers, but within eBPF’s restricted environment the damage is confined and cannot be steered toward arbitrary memory writes. It primarily results in unpredictable logic, not exploitable memory corruption.

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

**Exploitable:** Not possible.

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

**Exploitable:**
- Memory safety exploitation (kernel R/W): No, not possible.
- Logic exploitation (attacker-controlled packet alteration): Yes, possible.

#### Example 3

**Implementation Details**

- Inside `syncookie_handle_syn`, a `__u32 tainted_vla_size` variable is declared and initialized with a value derived from `hdr->tcp->doff * 4`.
- The line `char vla_buffer[tainted_vla_size]` attempts to declare a Variable Length Array `vla_buffer` using this runtime-determined size.
- This program **will not pass verification**, regardless of the taintedness of `tainted_vla_size`. The eBPF verifier, as mentioned in the previous example, prohibits VLAs. The compilation succeeds, but the attempt to load such a BPF program into the kernel will result in a clear rejection message from the verifier.

**Verifier:** Not passed.

**Exploitable:** Not possible.

*Signed-by: Francesco Rollo*
