# MEMORY.md - Long-Term Memory

This is Premchand's curated memory. Only significant events, decisions, and learnings go here.

---

## 👤 IDENTITY

- **Name:** Premchand
- **Role:** CEO of Formatho
- **Emoji:** 🏗️
- **Vibe:** Sharp, decisive, execution-focused
- **Mission:** Build Agent Orchestrator, launch 6 libraries, get revenue

---

## 🏢 COMPANY: FORMATHO

### Products
1. **Formatho.com** — Free web tools (maintenance mode)
2. **Agent Orchestrator** — Desktop app for AI agents (PRIMARY FOCUS)

### GitHub
- **Org:** https://github.com/formatho
- **Main repo:** https://github.com/formatho/agent-orchestrator
- **Account:** Ritavidhata

### Headquarters
- **Location:** `/Users/studio/sandbox/formatho/headquarters/`
- **Codebase:** `/Users/studio/sandbox/formatho/agent-orchestrator/`

---

## 📦 6 LIBRARIES STRATEGY

| # | Library | Purpose | Status |
|---|---------|---------|--------|
| 1 | go-llm-client | Unified LLM interface | ✅ COMPLETE |
| 2 | go-agent-pool | Agent lifecycle | ✅ COMPLETE |
| 3 | go-agent-skills | Skills + permissions | ✅ COMPLETE |
| 4 | go-todo-queue | Persistent TODO | 📋 Week 2 |
| 5 | go-cron-agents | Cron for agents | 📋 Week 2 |
| 6 | go-agent-config | Config management | 📋 Week 2 |

---

## 👤 FOUNDER INFO

- **Timezone:** Asia/Calcutta (GMT+5:30)
- **WhatsApp:** +971585903620
- **GitHub:** Ritavidhata
- **Tech-savvy, uses OpenClaw TUI**

---

## 🛠️ LOCAL SERVICES

- **ComfyUI:** localhost:8188 (image generation)
- **OpenClaw Gateway:** localhost:18789

---

## 📅 KEY DATES

- **2026-03-05:** Project launched, GitHub org created, first code pushed
- **2026-03-05:** go-llm-client v0.1.0 complete (OpenAI + Anthropic + Ollama + retry + tests)
- **2026-03-05:** Strategic pivot: Marketing deferred until working product (UI + backend)

---

## 🎯 STRATEGY UPDATE (March 5, 2026)

**Old Plan:** Launch each library separately on HN/Reddit
**New Plan:** Build complete product first, then market

**Why?**
- Better to launch working product than just libraries
- Users want solutions, not just infrastructure
- Higher impact launch with UI + backend together

**Marketing Timeline:**
- Phase 1: Build all 6 libraries (Weeks 1-5)
- Phase 2: Build UI + backend integration (Weeks 6-8)
- Phase 3: Polish + test (Week 9)
- Phase 4: Launch marketing (Week 10)

**Target:** Week 10 for marketing push (HN, Reddit, Twitter, Product Hunt)

---

## 🏗️ ARCHITECTURE PLANNING (March 5, 2026 - 10:20 PM)

### Planning Documents Created
1. **ARCHITECTURE.md** - Full stack overview
   - UI/UX design with 5 main screens
   - Frontend: Electron + React + TypeScript
   - Backend: Go + Fiber + WebSocket
   - Communication: REST API + WebSocket

2. **ELECTRON_PLAN.md** - Frontend implementation
   - Project setup (Vite + Electron)
   - Component structure (20+ components)
   - API client + WebSocket client
   - Build & distribution strategy

3. **BACKEND_PLAN.md** - Backend implementation
   - Go server with Fiber framework
   - REST API (25+ endpoints)
   - WebSocket for real-time updates
   - SQLite database with migrations

4. **PARALLEL_DEVELOPMENT_PLAN.md** - Library development strategy
   - 5 agent teams working in parallel
   - Wave 1: pool-dev + skills-dev (RUNNING)
   - Wave 2: todo-dev + cron-dev (QUEUED)
   - Wave 3: config-dev (QUEUED)

### Technology Stack
- **Frontend:** Electron, React 18, TypeScript, Vite, Tailwind CSS, Shadcn/ui
- **Backend:** Go, Fiber, gorilla/websocket, SQLite
- **Libraries:** Our 6 Go packages (1 done, 5 in progress)
- **Communication:** REST API + WebSocket (real-time)

### UI Screens Planned
1. **Dashboard** - Agent overview, activity feed, resource usage
2. **Agent Detail** - Status, controls, live logs, task history
3. **TODO Queue** - Priority queue, progress tracking, filters
4. **Cron Scheduler** - Job management, run history
5. **Configuration** - Global settings, LLM config, skill permissions

### Current Status
- ✅ go-llm-client: COMPLETE (6,000 lines, 3 providers)
- 🏃 go-agent-pool: BUILDING (pool-dev agent)
- 🏃 go-agent-skills: BUILDING (skills-dev agent)
- 📋 go-todo-queue: QUEUED (Wave 2)
- 📋 go-cron-agents: QUEUED (Wave 2)
- 📋 go-agent-config: QUEUED (Wave 3)

---

## ⚠️ PENDING

- [x] Install Go (`brew install go`) ✅ DONE
- [x] Implement OpenAI provider ✅ DONE
- [x] Add retry logic + tests ✅ DONE
- [x] Add Anthropic + Ollama providers ✅ DONE
- [ ] Build go-agent-pool (Week 2)
- [ ] Build remaining libraries (Weeks 3-5)
- [ ] Build UI + backend integration (Weeks 6-8)
- [ ] Launch marketing (Week 10)

---

*Last updated: 2026-03-05 (10:00 PM)*
*By Premchand 🏗️*
