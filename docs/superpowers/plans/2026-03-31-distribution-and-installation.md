# Buffer7 分发与安装实施计划 (Distribution and Installation)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现基于 GoReleaser 的自动化分发系统，提供 Homebrew Tap 和一键安装脚本。

**Architecture:** 使用 GoReleaser 作为核心构建和同步引擎。通过 GitHub Actions 触发发布，自动同步 Formula 到独立的 Homebrew Tap 仓库。提供 `install.sh` 脚本用于传统的下载安装。

**Tech Stack:** GoReleaser, GitHub Actions, Homebrew (Formula), Bash, gh CLI.

---

### Task 1: 初始化 Homebrew Tap 仓库

**Files:**
- Create: `/Users/gaojin/Documents/GitHub/homebrew-tap/README.md`
- Shell: `gh repo create buffer7/homebrew-tap --public --source=/Users/gaojin/Documents/GitHub/homebrew-tap --remote=origin --push`

- [ ] **Step 1: 初始化本地 git 仓库并提交 README**

```bash
cd /Users/gaojin/Documents/GitHub/homebrew-tap
git init
echo "# Homebrew Tap for Buffer7" > README.md
git add README.md
git commit -m "initial commit"
```

- [ ] **Step 2: 使用 gh 创建远程仓库并推送**

```bash
gh repo create buffer7/homebrew-tap --public --source=/Users/gaojin/Documents/GitHub/homebrew-tap --remote=origin --push
```

### Task 2: 配置 GoReleaser

**Files:**
- Create: `.goreleaser.yaml`

- [ ] **Step 1: 创建 .goreleaser.yaml 配置文件**

```yaml
# .goreleaser.yaml
before:
  hooks:
    - go mod tidy
builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}
      - "{{ if eq .Os \"darwin\" }}-linkmode=external{{ end }}"

archives:
  - format: tar.gz
    # 用 .zip 格式打包 windows 二进制文件
    format_overrides:
      - goos: windows
        format: zip
    name_template: "{{ .ProjectName }}-{{ .Os }}-{{ .Arch }}"

checksum:
  name_template: "checksums.txt"

snapshot:
  name_template: "{{ .Version }}-snapshot"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"

brews:
  - name: buffer7
    repository:
      owner: buffer7
      name: homebrew-tap
      token: "{{ .Env.GORELEASER_TOKEN }}"
    directory: Formula
    homepage: "https://github.com/buffer7/buffer7"
    description: "High-performance, zero-configuration software supply chain security tool."
    license: "MIT"
    install: |
      bin.install "buffer7"
```

- [ ] **Step 2: 验证配置文件**

运行: `goreleaser check` (如果本地安装了 goreleaser)
预期: 配置文件格式正确

- [ ] **Step 3: 提交配置**

```bash
git add .goreleaser.yaml
git commit -m "feat: add goreleaser configuration"
```

### Task 3: 更新 GitHub Actions 流程

**Files:**
- Modify: `.github/workflows/release.yml`

- [ ] **Step 1: 替换现有的 release.yml 逻辑为 GoReleaser 插件**

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GORELEASER_TOKEN: ${{ secrets.GORELEASER_TOKEN }}
```

- [ ] **Step 2: 提交 CI 更改**

```bash
git add .github/workflows/release.yml
git commit -m "ci: switch to goreleaser for releases"
```

### Task 4: 编写自动化安装脚本 (install.sh)

**Files:**
- Create: `install.sh`

- [ ] **Step 1: 编写安装脚本**

```bash
#!/bin/sh
set -e

# 设置变量
OWNER="buffer7"
REPO="buffer7"
BINARY_NAME="buffer7"

# 获取最新版本
LATEST_RELEASE_URL="https://api.github.com/repos/$OWNER/$REPO/releases/latest"
VERSION=$(curl -s $LATEST_RELEASE_URL | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Error: Could not find latest release version."
    exit 1
fi

echo "Installing $REPO $VERSION..."

# 检测平台
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Error: Unsupported architecture $ARCH"; exit 1 ;;
esac

# 下载 URL
FILENAME="${REPO}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$OWNER/$REPO/releases/download/$VERSION/$FILENAME"

# 下载并安装
TMP_DIR=$(mktemp -d)
curl -L "$DOWNLOAD_URL" -o "$TMP_DIR/$FILENAME"
tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"

echo "Moving binary to /usr/local/bin..."
sudo mv "$TMP_DIR/$BINARY_NAME" /usr/local/bin/
sudo chmod +x "/usr/local/bin/$BINARY_NAME"

# 清理
rm -rf "$TMP_DIR"

echo "Successfully installed $BINARY_NAME version $(buffer7 --version 2>/dev/null || echo $VERSION)"
```

- [ ] **Step 2: 提交脚本**

```bash
chmod +x install.sh
git add install.sh
git commit -m "feat: add installation script"
```

### Task 5: 验证与总结

- [ ] **Step 1: 推送 master 到远程**

```bash
git push origin master
```

- [ ] **Step 2: 生成测试 Tag 并推送**

```bash
git tag v0.0.1
git push origin v0.0.1
```

- [ ] **Step 3: 检查 GitHub Actions 状态**

访问: `https://github.com/buffer7/buffer7/actions`
预期: `goreleaser` 任务成功完成

- [ ] **Step 4: 验证 Homebrew Tap**

运行: `brew install buffer7/tap/buffer7`
预期: 安装成功且版本正确
