# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## Project Overview

gkube is a Kubernetes management web platform (similar to Kuboard). It's a full-stack monorepo with a Go backend (Gin framework) and a Vue 3 SPA frontend (Element Plus UI). The platform manages multiple K8s clusters and provides CRUD operations on all major K8s resource types, plus a web terminal (xterm.js) and log viewer.

## Commands

### Backend (run from `backend/`)

```bash
go run main.go              # Start the API server (port 8080)
go run main.go migrate      # Auto-migrate database schema
go run main.go seed         # Seed default roles, permissions, and admin user
go build -o gkube .         # Build binary
go test ./...               # Run tests (none written yet)
```

### Frontend (run from `frontend/`)

```bash
npm install                 # Install dependencies
npm run dev                 # Vite dev server with proxy to backend :8080
npm run build               # Type-check (vue-tsc) + production build
npm run preview             # Preview production build
```

### Infrastructure (run from `backend/`)

```bash
docker-compose up -d        # Start MySQL, Elasticsearch
docker-compose up           # Start all services including the app
```

## Architecture

### Backend (`backend/`)

**Startup chain** (`cmd/root.go`): config.Init() → logger.Init() → database.Init() (MySQL/GORM) → es.Init() → HTTP server on :8080 → cluster health checker goroutine.

**Module pattern** — each domain under `app/<module>/`:
- `api/` — Gin HTTP handlers (receives request, calls service, returns response)
- `params/` — Request parameter structs (bound from query/body)
- `model/` — GORM database models
- `service/` — Business logic (used by some modules like cluster)

**Shared packages** under `pkg/`:
- `k8s/` — One sub-package per K8s resource type (deployment/, pod/, service/, etc.), each with functions that call `client-go`. Initialized in `pkg/k8s/init.go`.
- `middleware/` — CORS, JWT auth (`JWTAuth()`), RBAC (`RBAC(resource, action)`)
- `response/` — Standardized JSON response helpers
- `auth/` — JWT token generation/validation, bcrypt password hashing
- `database/` — GORM MySQL connection setup
- `redis/` — Redis client wrapper
- `es/` — Elasticsearch client for audit logging
- `logger/` — Logrus with lumberjack log rotation

**Routing** (`router/router.go`): All API routes under `/v1`. Public routes: `POST /v1/auth/login`, `POST /v1/auth/refresh`. All other routes require JWT. RBAC middleware is applied per resource group (user, role, cluster). K8s routes are under `/v1/k8s/` with per-resource prefixes (deployment/, pod/, service/, etc.).

**Configuration** (`config/config.yaml`): Server, MySQL, Redis, Elasticsearch, S3 (placeholder), audit settings. Loaded via Viper.

### Frontend (`frontend/`)

**Stack**: Vue 3 + TypeScript + Vite + Pinia + Element Plus + Vue Router.

**Key directories**:
- `src/api/` — Axios API clients: `auth.ts`, `cluster.ts`, `dashboard.ts`, `resource.ts` (generic K8s resource CRUD), `request.ts` (axios instance with interceptors)
- `src/stores/` — Pinia stores: `auth.ts` (login/token), `cluster.ts` (selected cluster)
- `src/views/` — Page components organized by K8s domain: `workload/`, `network/`, `storage/`, `config/`, `node/`, `namespace/`, `event/`, `terminal/`, `logviewer/`, `dashboard/`, `login/`
- `src/router/index.ts` — All SPA routes
- `src/components/` — Reusable components (YAML editor, layout shell, forms)

**Dev proxy** (`vite.config.ts`): `/api` → `http://localhost:8080` (strips `/api` prefix), `/v1` → `http://localhost:8080` (with WebSocket support for terminal).

### Multi-cluster Design

The platform supports multiple K8s clusters stored in MySQL. Each API request includes a cluster ID; the backend dynamically creates a `client-go` client for the target cluster using stored kubeconfig. The frontend lets users switch clusters via a selector in the header.

## Key Technology Stack

| Layer | Technology |
|-------|-----------|
| HTTP framework | Gin |
| ORM | GORM (MySQL) |
| K8s client | client-go |
| Config | Viper (YAML) |
| CLI | Cobra |
| Logging | Logrus + Lumberjack |
| Audit storage | Elasticsearch 7.x |
| WebSocket | gorilla/websocket |
| Object storage | MinIO (S3-compatible) |
| Frontend framework | Vue 3 + TypeScript |
| UI library | Element Plus |
| State management | Pinia |
| Build tool | Vite |
| Terminal | xterm.js |
| YAML editor | Monaco Editor |

## No Tests or Linting

There are no test files (`*_test.go`, `*.test.ts`, `*.spec.ts`) and no linting configuration in this project. The TypeScript config enables `noUnusedLocals`, `noUnusedParameters`, and `noFallthroughCasesInSwitch` as compiler-level checks.
