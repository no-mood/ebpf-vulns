# eBPF Vulnerability Testing Suite - ISO-IEC TS 17961-2013

This project provides a comprehensive testing environment for eBPF/XDP programs with **intentional vulnerability patches** based on the **ISO-IEC TS 17961-2013** standard. The suite includes automated VM provisioning, vulnerability injection, and verifier analysis tools.

## Overview

The project demonstrates security vulnerabilities in eBPF/XDP kernel code by implementing various rules from ISO-IEC TS 17961-2013. Each vulnerability is carefully crafted to show:
- How coding standard violations manifest in eBPF programs
- eBPF verifier behavior and limitations
- Compiler diagnostic capabilities
- Real-world security implications

## Project Structure

```
ebpf-tests-3-1/
├── virt/                    # VM management scripts
│   └── vmctl.sh            # VM creation, destruction, and connection
├── XDPs/
│   └── xdp_synproxy/       # XDP SYN proxy implementation (from Linux selftests)
│       ├── apply_rules     # Network rules configuration script
│       ├── start_session.sh # Tmux session setup for testing
│       └── patches/        # ISO-IEC TS 17961-2013 vulnerability patches
├── xvtlas/                 # XDP Verifier Launch Automation Suite
├── pretty-verifier/        # Python verifier output formatter
└── docs/                   # Documentation and references
```

## Quick Start

### 1. VM Environment Setup

The project uses a standardized Ubuntu VM environment for consistent testing:

```bash
cd virt/
./vmctl.sh create ~/.ssh/id_rsa.pub    # Create VM with your SSH key
./vmctl.sh connect                      # Connect to the VM
```

**VM Management Commands:**
- `./vmctl.sh create <ssh_pubkey_file>` - Create and configure new VM
- `./vmctl.sh destroy` - Destroy the VM
- `./vmctl.sh connect` - Connect/reconnect to existing VM

**VM Connection Details:**
- After VM creation, you may need to press **`ENTER`** a few times to reach the login prompt
- Default login credentials (as specified in `virt/user-data.yaml`):
  - Username: `user`
  - Password: `` (empty password)
**VM Connection Details:**
- After VM creation, you can connect either via:
  - `./vmctl.sh connect` - Direct console connection
  - `ssh user@<vm-ip>` - SSH connection using the pubkey configured during creation

The script uses cloud-config with a modified `user-data.yaml` file to provision an Ubuntu environment with all necessary development tools pre-installed. The SSH public key is automatically copied into the VM during the creation process.

### 2. Repository Setup (Inside VM)

```bash
# Clone the repository inside the VM (with submodules for pretty-verifier)
git clone --recurse-submodules <repository-url> ebpf-tests-3-1
cd ebpf-tests-3-1
```

### 3. XDP SynProxy Configuration

Configure the necessary network rules for XDP SynProxy operation:

```bash
cd XDPs/xdp_synproxy/
./apply_rules
```

This script configures:
- Network interface settings
- iptables rules for SYN proxy operation
- Kernel parameters for eBPF program loading

### 4. Testing Environment Setup

The project includes a testing script to start a tmux session with the necessary components:

```bash
cd XDPs/xdp_synproxy/
./start_session.sh
```

To close the session and clean the env : 

```bash
//From inside the tmux session : 
./kill_session.sh
```

**Network Topology:**
The current testing setup uses a simplified topology:
- **XDP SynProxy**: Running inside the VM on the network interface
- **Netcat Server**: Running inside the VM to test connections
- **Netcat Client**: Running on the host machine (outside VM) connecting to the server

**Design Decision:**
Initially, we followed the Linux kernel selftests approach using a 3-veth (virtual Ethernet) interface topology. However, virtual interfaces conflicted with the SYN cookie functionality, due to checksum problems, as noted in the Linux kernel source code comments. This simplified approach provides a realistic testing environment without these conflicts.

### 5. Vulnerability Patches

All vulnerability patches target `xdp_synproxy_kern.c`, an XDP-based SYN proxy implementation taken from the Linux kernel selftests. The `XDPs/xdp_synproxy/patches/` directory contains vulnerability patches for each applicable rule from ISO-IEC TS 17961-2013:

| Rule | Directory | Vulnerability Type |
|------|-----------|-------------------|
| 5.1 | `5_01_invalidptr/` | Creation of invalid pointers through out-of-bounds indexing |
| 5.4 | `5_4_boolasgn/` | Assignment in conditional expressions |
| 5.6a | `5_06a_argcomp/` | Calling functions with wrong number of arguments |
| 5.6b | `5_06b_argcomp/` | Calling functions with wrong argument types |
| 5.6c | `5_06c_argcomp/` | Calling functions with wrong argument structures |
| 5.6d | `5_06d_argcomp/` | Calling functions with wrong argument arrays |
| 5.9 | `5_9_padcomp/` | Comparison of padding data |
| 5.10 | `5_10_intptrconv/` | Pointer-to-integer conversion issues |
| 5.11 | `5_11_alignptr/` | Accessing memory through misaligned pointers |
| 5.13 | `5_13_objdec/` | Accessing objects through incompatible effective types |
| 5.14 | `5_14_nullref/` | Null pointer dereferencing and out-of-domain pointers |
| 5.15 | `5_15_addrescape/` | Address escaping of automatic variables |
| 5.16 | `5_16_signconv/` | Converting tainted values between signed/unsigned |
| 5.17 | `5_17_swtchdflt/` | Switch statements with incomplete enum coverage |
| 5.22 | `5_22_invptr/` | Using out-of-bounds pointers or array subscripts |
| 5.24 | `5_24_usrfmt/` | Including tainted or out-of-domain input in format strings |
| 5.26 | `5_26_diverr/` | Integer division errors |
| 5.28 | `5_28_strmod/` | Modifying string literals |
| 5.30 | `5_30_intoflow/` | Signed integer overflow |
| 5.31 | `5_31_nonnullcs/` | Non-null-terminated character sequences |
| 5.33 | `5_33_restrict/` | Pointers into the same object with restrict qualifier |
| 5.35 | `5_35_uninit_mem/` | Referencing uninitialized memory |
| 5.36 | `5_36_ptrobj/` | Pointer comparison/subtraction from different objects |
| 5.39 | `5_39_taintnoproto/` | Using tainted values as function pointers without prototypes |
| 5.45 | `5_45_invfmtstr/` | Invalid format strings |
| 5.46 | `5_46_taintsink_1/` | Tainted potentially mutilated non-character data (variant 1) |
| 5.46 | `5_46_taintsink_2/` | Tainted potentially mutilated non-character data (variant 2) |

Each vulnerability rule is implemented as a Git commit patch that modifies the base `xdp_synproxy_kern.c` file. These patches can be:
- **Applied manually**: Use `git apply` or `git am` to apply individual patches for manual testing
- **Applied automatically**: Use the XVTLAS tool which handles patch application, compilation, verification, and output export

The patches are designed to demonstrate specific ISO-IEC TS 17961-2013 rule violations while maintaining the core SYN proxy functionality.

### Rules Not Applicable to XDP/eBPF

The following ISO-IEC TS 17961-2013 rules are **not applicable** to XDP/eBPF environments due to fundamental limitations and architectural differences:

| Rule | Title | Category | Reason |
|------|-------|----------|---------|
| **5.2** | Accessing freed memory | Memory Management | No `free()` function or dynamic memory allocation |
| **5.3** | Accessing shared objects in signal handlers | Signal Handling | No signal handling mechanism in eBPF programs |
| **5.5** | Calling functions from signal handlers except abort, _Exit, signal | Signal Handling | No signal handling mechanism in eBPF programs |
| **5.7** | Calling signal from interruptible signal handlers | Signal Handling | No signal handling mechanism in eBPF programs |
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
| **5.37** | Tainted strings are passed to a string copying function | Format String | Limited string manipulation functions in eBPF context |
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


### 6. XVTLAS - XDP Verifier Launch Automation Suite

**XVTLAS** automates the entire process of:
- Compiling eBPF programs
- Applying vulnerability patches
- Loading programs with bpftool
- Analyzing verifier output
- Generating structured reports

#### Installation

```bash
cd xvtlas/
# See xvtlas/README.md for detailed compilation instructions
# The tool is written in Go and requires compilation
go build -o xvtlas .
```

#### Usage

```bash
# Apply and test a single vulnerability patch interactively
./xvtlas --run-single "./XDPs/xdp_synproxy/patches/5_45_invfmtstr/*.patch" \
         --base-file "./XDPs/xdp_synproxy/xdp_synproxy_kern.c"

# Run comprehensive testing on all patches
./xvtlas --export "./output/" \
         --kernel "6.8.58" \
         --patch-path "./XDPs/xdp_synproxy/patches/" \
         --base-file "./XDPs/xdp_synproxy/xdp_synproxy_kern.c" \
         --pretty "./pretty-verifier/pretty_verifier.py" \
         --save-logs \
         --verbose \
         --interactive

# Clean up after interactive session
./xvtlas --destroy

# Manual patch application (alternative approach)
cd XDPs/xdp_synproxy/
git apply patches/5_45_invfmtstr/0001-feat-5.45-invfmtstr-*.patch
make  # Compile manually
```

For detailed usage instructions, refer to `xvtlas/README.md`.


## Educational Purpose

This suite is designed for:
- **Security Research**: Understanding eBPF/XDP vulnerability patterns
- **Secure Coding Training**: Learning to avoid common pitfalls
- **Verifier Analysis**: Understanding eBPF verifier capabilities and limitations
- **Compiler Diagnostics**: Testing static analysis tool effectiveness

## Architecture

- **VM Environment**: Standardized Ubuntu cloud-config setup
- **XDP Target**: Real-world SYN proxy implementation
- **Patch System**: Git-based vulnerability injection
- **Automation**: Go-based testing orchestration
- **Analysis**: Python-based verifier output formatting

## Requirements

- Host system with KVM/QEMU support
- SSH key pair for VM access
- Go compiler (for XVTLAS)
- Python 3.x (for pretty-verifier)


## Developer Flow

For developers who want to contribute new vulnerability patches:

### Local Development Workflow

1. **Set up your preferred development environment** on the host machine
2. **Identify the target vulnerability** from ISO-IEC TS 17961-2013 standard
3. **Modify the source code** locally:
   ```bash
   # Edit the file in your preferred editor/IDE
   vim XDPs/xdp_synproxy/xdp_synproxy_kern.c

   # Focus on functions like tcp_dissect() for realistic vulnerability injection
   # Add clear comments explaining the vulnerability and expected behavior
   ```

4. **Transfer modified file to VM** using one of these methods:
   ```bash
   # Option 1: SCP copy
   scp XDPs/xdp_synproxy/xdp_synproxy_kern.c user@<vm-ip>:~/ebpf-tests-3-1/XDPs/xdp_synproxy/

   # Option 2: Mount via SSHFS (recommended for iterative development)
   mkdir ~/vm-mount
   sshfs user@<vm-ip>:~/ebpf-tests-3-1 ~/vm-mount
   # Now you can edit files directly in ~/vm-mount/
   ```

5. **Test inside the VM**:
   ```bash
   # Connect to VM
   ./virt/vmctl.sh connect

   # Compile with appropriate warning flags
   cd ~/ebpf-tests-3-1/XDPs/xdp_synproxy/
   make

   # Test the vulnerability behavior
   ./start_session.sh  # Start tmux testing environment
   ```

6. **Create patch** when satisfied with the vulnerability implementation:
   ```bash
   # Commit your changes
   git add xdp_synproxy_kern.c
   git commit -m "feat: 5.XX-rulename - Description of vulnerability"

   # Generate patch file
   git format-patch HEAD~1 -o patches/5_XX_rulename/
   ```

### Testing Your Vulnerability

- **Compiler Diagnostics**: Test with various warning flags (`-Wall`, `-Wformat`, `-Wpointer-arith`)
- **eBPF Verifier**: Load the program and observe verifier behavior
- **Runtime Behavior**: Use the tmux session to test actual packet processing
- **Documentation**: Update comments to explain expected vs actual behavior

## Contributing

When adding new vulnerability patches:
1. Follow the ISO-IEC TS 17961-2013 standard classification
2. Include detailed comments explaining the vulnerability
3. Document expected compiler/verifier behavior
4. Test in isolated VM environment
5. Update patch directory structure accordingly
6. Use the Developer Flow above for consistent development process

## References

- [ISO-IEC TS 17961-2013](https://www.iso.org/standard/61134.html) - C Static Analysis Standard
- [eBPF Documentation](https://ebpf.io/) - Extended Berkeley Packet Filter
- [XDP Documentation](https://www.kernel.org/doc/html/latest/networking/filter.html) - eXpress Data Path
- [BPF Verifier](https://www.kernel.org/doc/html/latest/bpf/verifier.html) - eBPF Program Verification
- [XDP SynProxy Examples](https://github.com/xdp-project/bpf-examples/tree/main/xdp-synproxy) - XDP-project BPF examples
- [Linux Kernel SynProxy Selftests](https://github.com/torvalds/linux/blob/v6.8/tools/testing/selftests/bpf/prog_tests/xdp_synproxy.c) - Linux kernel selftests

## License

TODO add license

## Authors

- Francesco Rollo
- Gianfranco Trad
- Giorgio Fardo
- Giovanni Nicosia

Developed for security research and educational purposes in the context of eBPF/XDP vulnerability analysis.

---
