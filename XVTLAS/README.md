# Introduction
**XVTLAS** tool is a eBPF verifier test automation suite, designed to automate test retrival, compilation and attachment of eBPF/XDP programs
that works in conjunction withe **Prettyverifier** and tracing points to retrieve information on successfull passing of the eBPF verifier.
...

---

# 🛠️ Installation Guide

This section covers different ways to install **XVTLAS** on your system.

## 🚀 Quick Install (Recommended)

If you have  >= **Go 1.18+** installed, you can install **XVTLAS** directly using:

```sh
go install <add_repo_url>
```

Once installed, ensure your Go binaries are in the system PATH:

```sh
export PATH=$PATH:$(go env GOPATH)/bin
```

Verify the installation:

```sh
xvtlas -help
```

---

## 📦 Install from Source (Manual Method)

If you prefer to build from source:

```sh
git clone <add_repo_url>
cd yourrepo
go build -o xvtlas .
```

Then move the binary to a system-wide location:

```sh
sudo mv xvtlas /usr/local/bin/
```

Now you can run:

```sh
xvtlas --help
```

---

## ✅ Verify Installation

After installation, check that **YourCLI** is correctly installed:

```sh
xvtlas -version #TODO add version command
xvtlas -help
```

---

## 🔧 Troubleshooting

- **Command Not Found?**  
  Ensure that `/usr/local/bin` is in your `PATH`:

  ```sh
  export PATH=$PATH:/usr/local/bin
  ```

- **Permission Denied?**  
  Try running:

  ```sh
  chmod +x xvtlas
  ```

- **Older Go Version?**  
  Upgrade Go with:

  ```sh
  go install golang.org/dl/go1.21.0@latest && go1.21.0 download
  ```

---

# Usage 

... #TODO

---


