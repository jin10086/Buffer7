# Buffer7

[![Go Report Card](https://goreportcard.com/badge/github.com/buffer7/buffer7)](https://goreportcard.com/report/github.com/buffer7/buffer7)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[English Version](README.md)

**Buffer7** 是一个高性能、零配置的软件供应链安全工具。它通过在本地建立一个“瞬时影子注册表（Ephemeral Shadow Registry）”，强制执行 **7 天规则**：任何发布时间不足 7 天的软件包版本，都会在安装阶段被自动过滤。

这种“时间窗口”策略能有效规避零日漏洞（Zero-day）、恶意包抢跑（Dependency Confusion/Typosquatting）以及不稳定版本带来的风险。

---

## ✨ 核心特性

- **🛡️ 7 天强制规则**：自动隐藏所有发布未满 7 天的版本，确保只安装经受过时间考验的稳定包。
- **🚀 瞬时代理 (Ephemeral Proxy)**：无需常驻后台。通过包装命令（如 `buffer7 npm install`）即时启动代理，命令结束即销毁。
- **🔌 零侵入设计**：不修改全局 `~/.npmrc` 或环境变量。仅在执行期间动态注入局部配置，安全可靠。
- **🏎️ 极致性能**：
    - **Go 语言实现**：单二进制文件，毫秒级启动，极低内存占用。
    - **智能重定向**：二进制包（.tgz/.whl）直接 302 重定向至原始镜像站，不占本地带宽。
- **🌐 多协议支持**：原生支持 **Node.js (npm/yarn/pnpm)** 和 **Python (pip/poetry/pdm)**。
- **🔐 透明转发**：完美透传 `Authorization` 等 Header，无缝支持私有仓库及公司内网源。

---

## 🚀 快速上手

### 安装

确保你已安装 Go 环境（1.16+），运行：

```bash
git clone https://github.com/buffer7/buffer7.git
cd Buffer7
make build
# 二进制文件将生成在 bin/buffer7
```

### 使用方法

只需在原有的安装命令前加上 `buffer7` 即可：

#### Node.js (npm)
```bash
./bin/buffer7 npm install lodash
```

#### Python (pip)
```bash
./bin/buffer7 pip install requests
```

### 它是如何工作的？
1. **启动**：Buffer7 启动并随机占用一个本地空闲端口。
2. **拦截**：它启动包管理器子进程，并注入 `NPM_CONFIG_REGISTRY` 或 `PIP_INDEX_URL` 指向本地端口。
3. **过滤**：当包管理器请求元数据时，Buffer7 抓取原始数据，剔除不合规版本，并动态修正 `latest` 标签。
4. **清理**：安装完成后，子进程退出，Buffer7 自动关闭代理，不留痕迹。

---

## 🛠️ 架构设计

Buffer7 采用插件化处理器架构：
- **NPM Handler**: 解析并过滤 JSON 元数据。
- **PyPI Handler**: 处理 Python 包发布的复杂版本关系。
- **Time Engine**: 严格对齐 UTC 时间戳。

---

## 🤝 贡献指南

我们欢迎任何形式的贡献！
1. Fork 本仓库。
2. 创建特性分支 (`git checkout -b feat/amazing-feature`)。
3. 运行测试确保一切正常 (`make test`)。
4. 提交你的更改 (`git commit -m 'feat: add some amazing feature'`)。
5. 开启一个 Pull Request。

---

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE) 开源。

Copyright (c) 2026 Buffer7 Authors.
