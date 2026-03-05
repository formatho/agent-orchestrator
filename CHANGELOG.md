# Changelog

All notable changes to Agent Orchestrator will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Initial release preparation

---

## [0.1.0] - 2026-03-06

### Added

#### Core Features
- **6 Modular Go Libraries**
  - `go-llm-client` - Unified LLM interface (OpenAI, Anthropic, Ollama)
  - `go-agent-pool` - Agent lifecycle management with resource limits
  - `go-agent-skills` - Skill execution engine with permissions
  - `go-todo-queue` - Persistent TODO management with SQLite
  - `go-cron-agents` - Cron scheduling for agents
  - `go-agent-config` - Multi-format configuration management

- **Backend API Server**
  - Go + Fiber framework
  - 44 REST API endpoints
  - WebSocket server for real-time updates
  - SQLite database with migrations
  - Health check and system status
  - Local LLM integration (LM Studio, Ollama)

- **Electron Application**
  - React 19 + TypeScript
  - Dark theme with Tailwind CSS
  - Dashboard with real-time metrics
  - Agent management with model selection
  - TODO management with priorities
  - Cron job scheduling
  - LLM model configuration
  - WebSocket real-time updates

- **Developer Experience**
  - Hot reload for development
  - TypeScript throughout
  - ESLint configuration
  - Build scripts for all platforms
  - GitHub Actions CI/CD
  - Comprehensive documentation

#### API Endpoints
- **Agents**: Create, Read, Update, Delete, Pause, Resume
- **TODOs**: Create, Read, Update, Delete, Start, Cancel
- **Cron**: Create, Read, Update, Delete, Pause, Resume, History
- **Config**: Get, Update, Test LLM connection
- **System**: Health check, Status

#### Supported Platforms
- **macOS**: x64 (Intel), arm64 (Apple Silicon)
- **Windows**: x64, ia32
- **Linux**: x64 (AppImage, DEB, RPM)

#### LLM Providers
- OpenAI (GPT-4, GPT-4-turbo, GPT-4o)
- Anthropic (Claude-3-opus, Claude-3-sonnet)
- Ollama (Local LLMs)
- LM Studio (Local LLMs)
- Custom providers via API

### Technical Details

**Backend:**
- Go 1.24
- Fiber v2.52.0
- SQLite with migrations
- WebSocket (gorilla/websocket)
- ~2,500 lines of code

**Frontend:**
- Electron 40
- React 19
- TypeScript 5.9
- Tailwind CSS 3.4
- React Query 5
- ~2,500 lines of code

**Libraries:**
- ~24,000 lines of Go code
- 200+ unit tests
- Comprehensive documentation

**Total Project:**
- ~29,000 lines of code
- MIT License
- 100% Open Source

### Tested With
- Local LLM: LM Studio (Qwen 3.5 35B)
- OpenAI API (GPT-4o)
- SQLite persistence
- WebSocket connections
- All CRUD operations

### Known Limitations
- Agent execution engine pending (library integration complete)
- Cron scheduler background process pending
- Code signing not configured (manual setup required)
- Auto-updates configured but not tested

---

## Release History

| Version | Date | Highlights |
|---------|------|------------|
| 0.1.0 | 2026-03-06 | Initial release |

---

## Roadmap

### v0.2.0 (Planned)
- [ ] Agent execution engine
- [ ] Cron scheduler background process
- [ ] More LLM providers
- [ ] Enhanced UI/UX

### v0.3.0 (Planned)
- [ ] Plugin system
- [ ] Team collaboration
- [ ] Cloud sync (optional)
- [ ] Mobile companion app

### v1.0.0 (Future)
- [ ] Stable API
- [ ] Production-ready
- [ ] Full documentation
- [ ] Community features

---

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

---

## License

MIT License - see [LICENSE](./LICENSE) for details.

---

*For more details on releases, see [RELEASE.md](./RELEASE.md)*
