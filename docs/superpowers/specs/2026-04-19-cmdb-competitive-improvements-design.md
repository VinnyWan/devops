# CMDB Competitive Improvement Design

**Date:** 2026-04-19
**Status:** Approved
**Approach:** Auditable Ops Platform — ops-first with deep audit, multi-tenant SaaS

## Problem Statement

The CMDB module needs to become the platform users prefer over AutoOps (天枢AutoOps). AutoOps is a feature-rich but architecturally shallow single-tenant platform. Our advantages are multi-tenant architecture, terminal audit recording, host-level RBAC, and clean code structure. This design closes the feature gaps while leveraging our architectural edge.

## Design Principles

1. **Ops-first, audit as support** — Dashboard and daily workflows feel like a general ops platform; audit trails are always present but not the hero
2. **Depth over breadth** — Each feature is polished and complete rather than check-the-box
3. **Multi-tenant native** — Every new feature is tenant-scoped from day one
4. **Phased delivery** — Ship value incrementally, not all at once

---

## Phase 1: Polish & Core Differentiation

### 1.1 Operations Dashboard

Replace the current placeholder Dashboard page with a rich ops dashboard.

**Stats Cards (top row):**
- Hosts Total (with online count)
- K8s Clusters (with running pods count)
- Cloud Instances (with last sync time)
- Active Terminals (with online user count)
- Tasks Today (with success rate)
- Alerts (with critical count)

**Host Status Distribution (ECharts donut chart):**
- Online / Warning / Offline breakdown
- Click-through to filtered host list

**Resource Trends (ECharts bar chart, 7-day):**
- Daily task execution count
- Daily terminal session count
- Stacked to show operational volume over time

**Unified Activity Feed:**
- Chronological feed aggregating events from: K8s operations, CMDB changes, task executions, file operations, cloud syncs, alert triggers
- Color-coded by source module
- Click-through to detail pages

**My Hosts Quick Access:**
- Per-user personalized grid showing 4-8 most-recently-accessed hosts
- Each card shows: hostname, IP, online status, recent session count
- One-click to open terminal or file browser
- Persisted per-user in Redis or DB

**Quick Action Buttons:**
- Open Terminal, Run Task, File Browser, View Audit
- Each opens the corresponding page with appropriate context

**Frontend:** Vue 3 page at `views/Dashboard/index.vue`, uses ECharts for charts, Pinia for user preferences. Backend adds a new `dashboard` API endpoint that aggregates stats from host, terminal, task, and alert tables.

### 1.2 Terminal Enhancements

#### 1.2.1 Multi-Tab Terminal

Users can open multiple SSH sessions as tabs within a single browser window.

**Frontend:**
- Tab bar component above the xterm.js terminal area
- Each tab shows: hostname (truncated), close button
- "+" button to open new connection dialog
- Keyboard shortcuts: Ctrl+Shift+T (new tab), Ctrl+Tab (next tab), Ctrl+W (close tab)
- Each tab has its own WebSocket connection and terminal instance
- Tab state (hostname, session ID) persisted in component state

**Backend:**
- No changes needed — each tab creates a new WebSocket connection through the existing terminal bridge
- Each session independently recorded in asciicast v2 format

#### 1.2.2 Batch Command Execution

Select multiple hosts, type one command, execute on all simultaneously.

**Frontend:**
- "Batch Command" button on host list page
- Host multi-select (checkboxes or group selector)
- Command input area with syntax highlighting
- Results panel showing per-host output in parallel
- Progress indicator (X/Y hosts completed)

**Backend:**
- New API endpoint: `POST /api/v1/cmdb/terminal/batch`
- Accepts: list of host IDs, command string, timeout
- Opens SSH connections to each host, executes command, captures output
- Streams results via WebSocket as each host completes
- Creates individual TerminalSession records for audit

#### 1.2.3 Command Snippet Library

Save, share, and quick-paste common commands.

**Data Model:**
- `CommandSnippet` table: ID, tenant ID, name, content, tags, creator, visibility (personal/team/public), created/updated timestamps

**Frontend:**
- Snippet panel (collapsible sidebar in terminal view)
- Search/filter by name or tags
- Click to insert into active terminal
- CRUD for managing snippets

**Backend:**
- Standard CRUD API under `/api/v1/cmdb/snippets/`
- Tenant-scoped, supports personal and shared visibility levels

#### 1.2.4 Smart Recording Tags

Label and search terminal recordings by tags and command content.

**Data Model:**
- Add `tags` (JSON array) and `command_summary` (text) fields to `TerminalSession`
- New `SessionTag` table: ID, tenant ID, session ID, tag (e.g., "incident-fix", "routine-check", "deploy")

**Features:**
- User can tag active or completed sessions with labels
- Session list supports filtering by tags
- Search recordings by command content (parse asciicast events for command strings)
- Export recording as shareable link (with expiry)

---

## Phase 2: Host Intelligence & Automation

### 2.1 Host Management Improvements

#### 2.1.1 SSH Auto-Discovery

When adding a host, automatically connect via SSH to gather system information.

**Backend:**
- `SyncHostInfo(hostID)` function: SSH to host, run `hostname`, `cat /etc/os-release`, `nproc`, `free -m`, `df -h`
- Parse output and populate host fields: OS, CPU cores, memory, disk
- Triggered automatically after host creation if credentials are provided
- Also available as manual "Refresh Info" action

#### 2.1.2 Excel/CSV Import & Export

Bulk import hosts from spreadsheet files.

**Backend:**
- Import: `POST /api/v1/cmdb/hosts/import` accepts multipart file (xlsx/csv)
  - Parse rows: hostname, SSH IP, port, username, group, OS, tags
  - Validate and create hosts in batch
  - Return import report: X created, Y skipped (duplicate), Z errors
- Export: `GET /api/v1/cmdb/hosts/export` generates xlsx from current host list
  - Apply current filters (group, status, search)
- Use `excelize/v2` library for Excel handling

#### 2.1.3 Asset Topology View

Visual network diagram showing host relationships and cloud topology.

**Frontend:**
- New page or toggle view in Host List
- Uses a graph visualization library (e.g., AntV G6 or vis-network)
- Shows: hosts as nodes, grouped by group hierarchy (color-coded)
- VPC/subnet boundaries shown as containers
- Click node to open host detail
- Filters: by group, by cloud provider, by status

**Backend:**
- New API: `GET /api/v1/cmdb/hosts/topology`
- Returns: groups tree, hosts with group assignments, cloud resource relationships (VPC/subnet)

#### 2.1.4 Multi-Cloud Support

Extend cloud sync beyond Tencent Cloud.

**Supported Providers:**
- Alibaba Cloud (ECS instances, VPC, security groups)
- AWS (EC2 instances, VPC, security groups) — future

**Backend:**
- Provider abstraction: interface with provider-specific implementations
- `CloudProvider` interface: `SyncInstances()`, `SyncVPC()`, `SyncSecurityGroups()`
- Tencent Cloud: existing implementation
- Alibaba Cloud: new implementation using Alibaba Cloud SDK
- Cloud account model already supports multiple providers — extend enum

### 2.2 Ansible Task Engine

Full Ansible playbook execution engine integrated into the platform.

#### 2.2.1 Core Architecture

**Components:**
- **Task Template** — reusable script/playbook definition
- **Task Execution** — a single run of a template on target hosts
- **Task Work** — per-host execution record within a task execution
- **Task Queue** — Redis-backed queue with priority levels
- **Scheduler** — cron-based recurring task execution

**Data Models:**
- `TaskTemplate`: ID, tenant ID, name, type (shell/python/ansible), content, variables (JSON), timeout, creator, tags
- `TaskExecution`: ID, tenant ID, template ID, targets (host IDs), status (pending/running/success/failed/partial), trigger (manual/scheduled), started/completed timestamps, output log path
- `TaskWork`: ID, execution ID, host ID, status, exit code, output, started/completed timestamps
- `TaskSchedule`: ID, tenant ID, template ID, cron expression, targets, enabled, last_run, next_run

#### 2.2.2 Ansible Playbook Execution

**Features:**
- Upload playbook ZIP files or connect to Git repository
- Parse playbook YAML for task listing
- Host group assignment with per-group variables
- Global variables support
- Real-time log streaming via WebSocket
- Per-host sub-task tracking (TaskWork records)
- Task timeout and retry configuration

**Backend:**
- Ansible binary invoked as subprocess
- Output captured and streamed via WebSocket
- Redis task queue with configurable concurrency per host
- Execution logs persisted to disk (`./data/task-logs/`)

#### 2.2.3 Cron Scheduling

- Global scheduler using `robfig/cron/v3`
- CRUD for scheduled tasks with cron expressions
- Start/stop/pause/resume lifecycle
- Execution history linked to schedule
- Configurable retry on failure

#### 2.2.4 Real-Time Output

- WebSocket endpoint for live task output streaming
- Frontend shows output as it arrives (like terminal)
- Per-host output tabs in batch execution view

#### 2.2.5 Task History

- Searchable list of all task executions
- Filter by: template, status, date range, trigger type
- Per-host drill-down showing individual results
- Full audit trail of who ran what, when, on which hosts

---

## Phase 3: Monitoring & Integration Hub

### 3.1 Monitoring

#### 3.1.1 Prometheus Integration (Primary)

- Configure Prometheus server URL in platform settings
- Query Prometheus HTTP API for host metrics: CPU, memory, disk, network, load
- Display metrics on host detail page and dashboard
- Alert rules based on Prometheus queries with configurable thresholds

**Backend:**
- `MonitorService` with Prometheus HTTP client
- Metric caching in Redis (5-minute TTL)
- Alert rule evaluation loop (runs every configurable interval)

#### 3.1.2 SSH Agentless Monitoring (Fallback)

For hosts without Prometheus node_exporter:
- Periodically SSH to hosts and collect metrics via standard commands
- Store snapshots in DB with timestamps
- Show basic trends (no high-resolution data)

**Backend:**
- Scheduler runs SSH metric collection at configurable interval (default 5 min)
- `HostMetric` table: host ID, timestamp, CPU%, memory%, disk%, load_avg
- Retention policy: aggregate to hourly after 7 days, daily after 30 days

### 3.2 Notification & Alert Routing

#### 3.2.1 Alert Rules

- `AlertRule` model: ID, tenant ID, name, metric (cpu/memory/disk/custom), condition (>, <, ==), threshold, duration, severity, notification channels (JSON)
- Evaluate rules against latest metrics on schedule
- Deduplication: same alert won't fire again within cooldown period

#### 3.2.2 Notification Channels

- **Feishu (Lark):** Webhook URL + optional App credentials for rich cards
- **DingTalk:** Robot webhook + optional keyword filters
- **WeChat Work:** Corp ID + Agent ID + Secret for corp messages
- **Email:** SMTP configuration for email alerts

**Backend:**
- `NotificationChannel` model: ID, tenant ID, type, config (JSON), enabled
- `NotificationService` with provider-specific implementations
- Alert routing: rules → channels → delivery with retry

#### 3.2.3 Escalation Rules

- Multi-level escalation: notify group A, if unacknowledged for X minutes, escalate to group B
- `EscalationPolicy` model: levels with delay and channel overrides

### 3.3 Database Asset Management

Track and manage database instances as first-class assets.

**Data Model:**
- `DatabaseAsset` table: ID, tenant ID, name, type (mysql/postgresql/redis/mongodb/elasticsearch), host, port, version, group, tags, accounts (JSON), description, status
- `DatabaseAccount` table: ID, database asset ID, username, auth type, permissions, description

**Features:**
- CRUD for database instances
- Group and tag organization
- SQL query console for supported types (MySQL, PostgreSQL)
- Query execution recording for audit
- Connection test

**Backend:**
- New `database` module under `internal/modules/cmdb/` or separate module
- Database driver connections for query execution (go-sql-driver/mysql, lib/pq)
- Query timeout enforcement
- Result set size limits

### 3.4 SSL Certificate Management

Track certificate expiry across domains and hosts.

**Data Model:**
- `SSLCertificate` table: ID, tenant ID, domain, issuer, issued date, expiry date, days remaining, auto-discovered, alert thresholds, status

**Features:**
- Manual certificate entry
- Auto-crawl: periodic TLS handshake to configured domains to discover certs
- Expiry alerts via notification channels
- Dashboard widget showing upcoming expirations

### 3.5 Application Management

Track application lifecycle with CI/CD integration.

**Data Model:**
- `Application` table: ID, tenant ID, name, type, description, business line, environment, language, git repo, build config, deploy config, domains (JSON), host associations, owner
- `ReleaseRecord` table: ID, application ID, version, status, trigger type, operator, started/completed, changelog

**Features:**
- Application CRUD with rich metadata
- Jenkins integration: trigger builds, check status, parameterized builds
- Release workflow: request → approval → deploy → verify
- Release history and rollback support

---

## Competitive Advantages (Post-Implementation)

| Feature | Our Platform | AutoOps |
|---------|-------------|---------|
| Multi-tenant SaaS | Built-in, tenant-scoped | Single-tenant only |
| Terminal audit recording | Asciicast v2, full replay | Not implemented |
| Host-level RBAC | Per-group hierarchical | Basic role-based |
| Terminal experience | Multi-tab, batch commands, snippets | Single session only |
| Code quality | Clean architecture, layered | 157K single files |
| Monitoring | Prometheus + SSH agentless | Custom agent (Linux only) |
| Task automation | Full Ansible + lightweight runner | Ansible only |
| Database assets | Tracked with SQL console | Basic inventory |
| SSL certificates | Auto-crawl + expiry alerts | Not implemented |
| Notifications | Feishu/DingTalk/WeChat/Email | Not implemented |
| Cloud providers | Tencent + Alibaba + (AWS future) | Tencent + Alibaba + Baidu |

## Implementation Priority

1. **Phase 1** (Dashboard + Terminal) — Highest impact, immediate user-facing value
2. **Phase 2** (Host Intel + Ansible) — Closes the biggest functional gaps
3. **Phase 3** (Monitoring + Integration) — Completes the platform story

Each phase delivers standalone value and can be shipped independently.
