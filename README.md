# Buffer7

[![Go Report Card](https://goreportcard.com/badge/github.com/jin10086/buffer7)](https://goreportcard.com/report/github.com/jin10086/buffer7)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[中文版 (Chinese Version)](README_CN.md)

**Buffer7** is a high-performance, zero-configuration software supply chain security tool. It establishes a local "Ephemeral Shadow Registry" and enforces a **7-day rule**: any software package version released for less than 7 days will be automatically filtered during the installation phase.

This "Time Window" strategy effectively mitigates risks from Zero-day vulnerabilities, malicious package racing (Dependency Confusion/Typosquatting), and unstable versions.

---

## ✨ Key Features

- **🛡️ 7-Day Enforcement**: Automatically hides versions released less than 7 days ago, ensuring only time-tested stable packages are installed.
- **🚀 Ephemeral Proxy**: No resident background process. Instantly starts a proxy by wrapping commands (e.g., `buffer7 npm install`) and destroys it upon completion.
- **🔌 Non-intrusive Design**: Does not modify global `~/.npmrc` or environment variables. Dynamically injects local configuration only during execution, ensuring safety and reliability.
- **🏎️ Extreme Performance**:
    - **Go Implementation**: Single binary, millisecond startup, extremely low memory footprint.
    - **Smart Redirection**: Direct 302 redirection for binary packages (.tgz/.whl) to original mirrors, saving local bandwidth.
- **🌐 Multi-protocol Support**: Native support for **Node.js (npm/yarn/pnpm)** and **Python (pip/poetry/pdm)**.
- **🔐 Transparent Forwarding**: Perfectly passes through `Authorization` and other headers, seamlessly supporting private repositories and internal mirrors.

---

## 🚀 Quick Start

### Installation

#### One-liner Command (Linux/macOS)
```bash
curl -fsSL https://raw.githubusercontent.com/jin10086/Buffer7/master/install.sh | sh
```

#### Manual Installation
1. Go to the [GitHub Releases](https://github.com/jin10086/Buffer7/releases) page.
2. Download the binary corresponding to your operating system and architecture.
3. Grant execution permissions: `chmod +x buffer7`.
4. Move the binary to a directory in your `PATH` (e.g., `/usr/local/bin`).

#### Build from Source
If you have a Go environment installed (1.16+):
```bash
git clone https://github.com/jin10086/buffer7.git
cd Buffer7
make build
# Binary will be generated in bin/buffer7
```

### Usage

Simply prefix your original installation command with `buffer7`:

#### Node.js (npm)
```bash
./bin/buffer7 npm install lodash
```

#### Python (pip)
```bash
./bin/buffer7 pip install requests
```

### How It Works
1. **Startup**: Buffer7 starts and occupies a random available local port.
2. **Interception**: It launches the package manager subprocess and injects `NPM_CONFIG_REGISTRY` or `PIP_INDEX_URL` pointing to the local port.
3. **Filtering**: When the package manager requests metadata, Buffer7 fetches the raw data, filters out non-compliant versions, and dynamically corrects the `latest` tag.
4. **Cleanup**: After installation, the subprocess exits, and Buffer7 automatically shuts down the proxy without leaving a trace.

---

## 🛠️ Architecture

Buffer7 uses a pluggable handler architecture:
- **NPM Handler**: Parses and filters JSON metadata.
- **PyPI Handler**: Handles complex versioning and releases for Python packages.
- **Time Engine**: Strictly aligns with UTC timestamps.

---

## 🧪 Testing

Buffer7 includes comprehensive tests, including unit tests, integration tests, and "real-world" End-to-End (E2E) tests.

- **E2E Tests**: Use real `npm` and `pip` binaries against a local mock registry to verify filtering and dependency resolution.
- **Upstream Override**: Use the `BUFFER7_UPSTREAM_REGISTRY` environment variable to point Buffer7 to a custom registry (useful for air-gapped environments or testing).

To run all tests:
```bash
make test
```
*Note: E2E tests will automatically skip if `npm` or `pip` is not found in your PATH.*

---

## 🤝 Contributing

We welcome any form of contribution!
1. Fork this repository.
2. Create a feature branch (`git checkout -b feat/amazing-feature`).
3. Run tests to ensure everything is fine (`make test`).
4. Commit your changes (`git commit -m 'feat: add some amazing feature'`).
5. Open a Pull Request.

---

## 📄 License

This project is open-sourced under the [MIT License](LICENSE).

Copyright (c) 2026 Buffer7 Authors.
