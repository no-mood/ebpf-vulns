# XVTlas - XDP Verifier Launch Automation Suite

**XVTlas** automates compilation, patching, loading, and verification of XDP eBPF programs. It leverages `bpftool`, `make`, and a Python "pretty verifier" script to provide structured loading and output for BPF/XDP developers.

---

## 🧰 Requirements

- Go (1.18+ recommended)
- `bpftool` (from `iproute2` package)
- `clang`, `llvm`, and `make` for compiling eBPF programs
- Python 3 with your `pretty_verifier.py` script
- PrettyVerifier

---

## ⚙️ Build

```bash
go build -o xvtlas .
```

This will produce an `xvtlas` binary in the current directory.

---

## 🚀 Usage

### Run on a directory of eBPF programs

```bash
sudo ./xvtlas \
  --export "./output/" \
  --kernel "6.8.58" \
  --path "./template_folders/" \
  --pretty "./pretty-verifier/pretty_verifier.py" \
  --verbose \
  --interactive
```

This compiles all programs found under `--path`, loads and verifies them using `bpftool`, and saves logs to `--export`.

---

### Run using patch-based processing

```bash
sudo ./XVTLAS_V1/xvtlas \
  --export "./output/" \
  --kernel "6.8.58" \
  --patch-path "./rules/" \
  --base-file "./linux-kernel/tools/testing/selftests/bpf/xdp_synproxy_kern.c" \
  --pretty "./pretty-verifier/pretty_verifier.py" \
  --verbose \
  --interactive
```

This will:
- Walk the patch folders under `--patch-path`
- Apply each folder's `patch.diff` to the `--base-file`
- Compile using the provided Makefile
- Load, verify, and unpin each resulting program
- Clean up the patched file (unless `--keep-patched` is set)

---

## 🧾 Command Line Options

| Flag             | Description                                                                 |
|------------------|-----------------------------------------------------------------------------|
| `--path`         | Root path for standard eBPF program directories (uses one Makefile per dir) |
| `--patch-path`   | Folder containing subfolders with `patch.diff`, `config.yaml`, and Makefile |
| `--base-file`    | Path to the original source file to which patches will be applied           |
| `--export`       | Directory to export logs, reports, and CSV outputs                          |
| `--pretty`       | Path to the Python pretty verifier script                                   |
| `--kernel`       | Target kernel version to associate with logs                                |
| `--verbose`      | Show detailed output from make, git, and verifier steps                     |
| `--interactive`  | Prompt on failure before continuing                                          |
| `--keep-patched` | Do not remove patched files after run (default is to remove)                |

---

## 🧪 Folder Structure (Standard Mode)

```
template_folders/
├── example1/
│   ├── main.c
│   ├── config.yaml
│   └── Makefile
├── example2/
│   ├── ...
```

---

## 📂 Folder Structure (Patch Mode)

```
tests
├── 1/
│   ├── config.yaml
│   ├── Makefile
│   └── patch.diff
├── 2/
│   ├── ...
```

Each folder contains a diff patch against the same `base-file`.

---

## 📌 Notes

- Programs are pinned under `/sys/fs/bpf/{program_name}` and unpinned automatically after each load (when using `--patch-path`)
- Logs are saved under the export path with `.log` and `.csv` formats
- `main.c` and `config.yaml` are required for each test folder

---

## 🆘 Help

```bash
./xvtlas --help
```

If no arguments are provided, help is printed and the tool exits with an error.

---
xvtlas --help`
