# Agent Orchestrator

[![CI](https://github.com/formatho/agent-orchestrator/actions/workflows/ci.yml/badge.svg)](https://github.com/formatho/agent-orchestrator/actions/workflows/ci.yml)

**Spin up AI workers. Let them run. Check results later.**

A local-first desktop application for managing autonomous AI agents. Built with Go + Electron.

---

## 🎯 What It Does

- **Create agents with text** — Describe what you want, spawn an agent
- **Agents work autonomously** — They use skills to complete TODOs
- **Cron scheduling** — Agents run on schedule, not just on-demand
- **100% local** — Your data stays on your machine

---

## 📦 Modular Libraries

Each component is a standalone Go library you can use independently:

| Library | Description |
|---------|-------------|
| [go-llm-client](./packages/llm-client) | Unified interface for OpenAI, Anthropic, Ollama |
| [go-agent-pool](./packages/agent-pool) | Concurrent agent lifecycle management |
| [go-agent-skills](./packages/agent-skills) | Skill system with permissions |
| [go-todo-queue](./packages/todo-queue) | Persistent TODO queue (SQLite) |
| [go-cron-agents](./packages/cron-agents) | Cron scheduler for agents |
| [go-agent-config](./packages/agent-config) | YAML/JSON config management |

---

## 🚀 Quick Start (Coming Soon)

```bash
# Install (not yet available)
go install github.com/formatho/agent-orchestrator/cmd/orchestrator@latest

# Or download desktop app
# https://github.com/formatho/agent-orchestrator/releases
```

---

## 🛠️ Development Status

| Component | Status |
|-----------|--------|
| `go-llm-client` | 🚧 In Progress |
| `go-agent-pool` | 📋 Planned |
| `go-agent-skills` | 📋 Planned |
| `go-todo-queue` | 📋 Planned |
| `go-cron-agents` | 📋 Planned |
| `go-agent-config` | 📋 Planned |
| Electron UI | 📋 Planned |

---

## 📋 Roadmap

See [ROADMAP.md](./ROADMAP.md) for week-by-week plan.

---

## 🤝 Contributing

MIT licensed. PRs welcome!

---

## 📫 Connect

- **Org:** https://github.com/formatho
- **Twitter:** [@heyformatho](https://twitter.com/heyformatho)
- **Website:** https://formatho.com

---

*Built by [Formatho](https://github.com/formatho)*
