# Buffer7 分发与安装设计文档 (Distribution and Installation)

**日期：** 2026-03-31
**目标：** 提供多种便捷的安装方式（Homebrew, 自动化安装脚本, GitHub Release），并实现基于 GoReleaser 的自动化发布流程。

## 1. 架构概览

本设计方案采用业界成熟的 **GoReleaser** 方案来替换现有的手动发布脚本。GoReleaser 将作为核心工具，负责多平台编译、打包、生成校验和以及更新 Homebrew Tap。

## 2. 发布流程 (GoReleaser)

- **触发条件**：推送符合 `v*` 格式的 Git Tag。
- **构建平台**：
  - `darwin/amd64`, `darwin/arm64`
  - `linux/amd64`, `linux/arm64`
  - `windows/amd64`
- **附件生成**：
  - `.tar.gz` (macOS/Linux)
  - `.zip` (Windows)
  - `checksums.txt` (SHA256 校验和)
- **GitHub Actions 集成**：更新 `.github/workflows/release.yml` 使用 `goreleaser/goreleaser-action`。

## 3. Homebrew Tap

- **Tap 仓库**：`github.com/buffer7/homebrew-tap`。
- **自动化同步**：GoReleaser 在发布新版本后，自动更新 `buffer7.rb` Formula 并推送到该仓库。
- **安装命令**：
  ```bash
  brew tap buffer7/tap
  brew install buffer7
  ```

## 4. 自动化安装脚本 (install.sh)

- **入口**：在项目根目录提供 `install.sh`。
- **逻辑**：
  1. 通过 GitHub API (`/repos/buffer7/buffer7/releases/latest`) 获取最新版本。
  2. 自动检测用户的 OS (darwin/linux) 和 Arch (amd64/arm64)。
  3. 下载对应的 `.tar.gz` 压缩包。
  4. 解压并将二进制文件移动到 `/usr/local/bin`。
  5. 检查安装结果。
- **一键安装命令**：
  ```bash
  curl -fsSL https://raw.githubusercontent.com/buffer7/buffer7/master/install.sh | sh
  ```

## 5. 任务分解 (Implementation Tasks)

1. **环境准备**：生成 GitHub PAT 并配置到项目的 Secret (`GORELEASER_TOKEN`)。
2. **本地配置**：在项目根目录创建 `.goreleaser.yaml`。
3. **CI 更新**：修改 GitHub Actions 配置文件以调用 GoReleaser。
4. **安装脚本编写**：编写并测试 `install.sh`。
5. **Homebrew 仓库初始化**：使用 `gh` 命令初始化 `homebrew-tap` 仓库。
6. **全链路验证**：推送一个测试 Tag 并验证 Release、Homebrew 以及安装脚本。

---

*请在执行前审核此设计，如果确认无误，我们将开始编写具体实现计划。*
