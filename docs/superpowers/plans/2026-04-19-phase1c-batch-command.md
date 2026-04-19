# Batch Command Execution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow users to select multiple hosts, type a command, and execute it simultaneously on all selected hosts with per-host output streaming and audit recording.

**Architecture:** Backend opens concurrent SSH connections (via existing SFTP/SSH infrastructure) to each selected host, executes the command, and streams results back via WebSocket. Each execution creates a `TerminalSession` record for audit. Frontend provides a host selector dialog, command input, and parallel output panel.

**Tech Stack:** Go (SSH/WebSocket), Vue 3 + Element Plus (frontend)

---

## File Structure

| Action | Path | Responsibility |
|--------|------|---------------|
| Create | `backend/internal/modules/cmdb/service/batch_command.go` | SSH execution, result collection |
| Create | `backend/internal/modules/cmdb/api/batch_command.go` | WebSocket handler for streaming |
| Modify | `backend/internal/modules/cmdb/api/common.go` | Add batch command service singleton |
| Modify | `backend/routers/v1/cmdb.go` | Register batch command route |
| Create | `frontend/src/api/cmdb/batch_command.js` | WebSocket URL builder |
| Create | `frontend/src/views/Cmdb/BatchCommand.vue` | Batch command page |
| Modify | `frontend/src/router/index.js` | Add batch command route |

---

### Task 1: Batch Command Backend Service

**Files:**
- Create: `backend/internal/modules/cmdb/service/batch_command.go`

- [ ] **Step 1: Create the batch command service**

Create `backend/internal/modules/cmdb/service/batch_command.go`:

```go
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
	"devops-platform/internal/modules/cmdb/terminal"
	"devops-platform/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BatchCommandRequest 批量命令请求
type BatchCommandRequest struct {
	HostIDs []uint `json:"hostIds" binding:"required"`
	Command string `json:"command" binding:"required"`
	Timeout int    `json:"timeout"` // seconds, default 30
}

// HostResult 单个主机的执行结果
type HostResult struct {
	HostID   uint   `json:"hostId"`
	HostName string `json:"hostName"`
	HostIP   string `json:"hostIp"`
	Status   string `json:"status"` // running, success, failed, timeout
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
}

// BatchCommandService 批量命令服务
type BatchCommandService struct {
	db         *gorm.DB
	hostRepo   *repository.HostRepo
	credRepo   *repository.CredentialRepo
	termSvc    *TerminalService
}

func NewBatchCommandService(db *gorm.DB) *BatchCommandService {
	return &BatchCommandService{
		db:       db,
		hostRepo: repository.NewHostRepo(db),
		credRepo: repository.NewCredentialRepo(db),
		termSvc:  NewTerminalService(db),
	}
}

// ExecuteOnHosts 在多个主机上执行命令，结果通过 channel 流式返回
func (s *BatchCommandService) ExecuteOnHosts(ctx context.Context, tenantID uint, req BatchCommandRequest, resultCh chan<- HostResult) {
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30
	}

	var wg sync.WaitGroup
	// Limit concurrency to 10
	sem := make(chan struct{}, 10)

	for _, hostID := range req.HostIDs {
		wg.Add(1)
		go func(hid uint) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := s.executeOnSingleHost(ctx, tenantID, hid, req.Command, timeout)
			resultCh <- result
		}(hostID)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()
}

func (s *BatchCommandService) executeOnSingleHost(ctx context.Context, tenantID, hostID uint, command string, timeoutSeconds int) HostResult {
	result := HostResult{
		HostID: hostID,
		Status: "running",
	}

	// Get host
	host, err := s.hostRepo.GetByIDInTenant(tenantID, hostID)
	if err != nil {
		result.Status = "failed"
		result.Error = "主机不存在"
		return result
	}
	result.HostName = host.Hostname
	result.HostIP = host.Ip

	// Get credential
	if host.CredentialID == nil {
		result.Status = "failed"
		result.Error = "主机未绑定凭据"
		return result
	}
	cred, err := s.credRepo.GetByIDInTenant(tenantID, *host.CredentialID)
	if err != nil {
		result.Status = "failed"
		result.Error = "凭据不存在"
		return result
	}

	// Decrypt credential password
	credPassword := ""
	if cred.Password != "" {
		credPassword, err = s.credRepo.DecryptPassword(cred.Password)
		if err != nil {
			result.Status = "failed"
			result.Error = "凭据解密失败"
			return result
		}
	}

	// Create SSH client
	sshClient, err := terminal.NewSSHClient(host.Ip, host.Port, cred.Username, credPassword, cred.PrivateKey)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("SSH 连接失败: %v", err)
		return result
	}
	defer sshClient.Close()

	// Execute command with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	output, err := terminal.ExecuteCommand(execCtx, sshClient, command)
	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			result.Status = "timeout"
			result.Error = "执行超时"
		} else {
			result.Status = "failed"
			result.Error = fmt.Sprintf("执行失败: %v", err)
		}
		result.Output = output
		return result
	}

	result.Status = "success"
	result.Output = output
	return result
}

// CreateBatchAuditRecords 为批量命令创建审计记录
func (s *BatchCommandService) CreateBatchAuditRecords(tenantID, userID uint, username string, req BatchCommandRequest, results []HostResult) {
	for _, r := range results {
		session := model.TerminalSession{
			TenantID:     tenantID,
			UserID:       userID,
			Username:     username,
			HostID:       r.HostID,
			HostIP:       r.HostIP,
			HostName:     r.HostName,
			CredentialID: 0, // batch command doesn't use a specific credential session
			Status:       "closed",
			CloseReason:  "batch_command",
			Duration:     0,
		}
		if err := s.db.Create(&session).Error; err != nil {
			logger.Log.Error("创建批量命令审计记录失败", zap.Error(err))
		}
	}
}
```

- [ ] **Step 2: Add helper functions to terminal package**

Add to `backend/internal/modules/cmdb/terminal/ssh.go` (at the end of the file):

```go
// NewSSHClient creates an SSH client connection without PTY (for command execution)
func NewSSHClient(host string, port int, username, password, privateKey string) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method available")
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	return ssh.Dial("tcp", addr, config)
}

// ExecuteCommand runs a command on an SSH client and returns the output
func ExecuteCommand(ctx context.Context, client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	type result struct {
		output []byte
		err    error
	}
	done := make(chan result, 1)

	go func() {
		out, e := session.CombinedOutput(command)
		done <- result{output: out, err: e}
	}()

	select {
	case <-ctx.Done():
		session.Close()
		return "", ctx.Err()
	case r := <-done:
		return string(r.output), r.err
	}
}
```

Add import for `"context"` to the ssh.go file if not already present.

- [ ] **Step 3: Verify compilation**

Run: `cd backend && go build ./internal/modules/cmdb/...`
Expected: compiles with no errors

- [ ] **Step 4: Commit**

```bash
cd backend
git add internal/modules/cmdb/service/batch_command.go internal/modules/cmdb/terminal/ssh.go
git commit -m "feat(cmdb): add batch command service with concurrent SSH execution"
```

---

### Task 2: Batch Command Backend API & Route

**Files:**
- Create: `backend/internal/modules/cmdb/api/batch_command.go`
- Modify: `backend/internal/modules/cmdb/api/common.go`
- Modify: `backend/routers/v1/cmdb.go`

- [ ] **Step 1: Create batch command API handler (WebSocket streaming)**

Create `backend/internal/modules/cmdb/api/batch_command.go`:

```go
package api

import (
	"encoding/json"
	"time"

	"devops-platform/internal/modules/cmdb/service"

	"github.com/gorilla/websocket"
	"github.com/gin-gonic/gin"
)

var batchUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func BatchCommandConnect(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}

	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uint)

	usernameValue, _ := c.Get("username")
	username, _ := usernameValue.(string)

	ws, err := batchUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	// Read the batch command request from WebSocket
	_, msg, err := ws.ReadMessage()
	if err != nil {
		return
	}

	var req service.BatchCommandRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		ws.WriteJSON(gin.H{"type": "error", "message": "请求格式错误"})
		return
	}

	svc := getBatchCommandService()
	if svc == nil {
		ws.WriteJSON(gin.H{"type": "error", "message": "服务未初始化"})
		return
	}

	// Execute commands and stream results
	resultCh := make(chan service.HostResult, len(req.HostIDs))
	ctx := c.Request.Context()

	svc.ExecuteOnHosts(ctx, tenantID, req, resultCh)

	var allResults []service.HostResult
	for result := range resultCh {
		allResults = append(allResults, result)
		ws.WriteJSON(gin.H{
			"type": "host_result",
			"data": result,
		})
	}

	// Send completion signal
	ws.WriteJSON(gin.H{
		"type":   "complete",
		"total":  len(req.HostIDs),
		"success": countSuccess(allResults),
		"failed":  len(allResults) - countSuccess(allResults),
	})

	// Create audit records in background
	go svc.CreateBatchAuditRecords(tenantID, userID, username, req, allResults)

	// Keep connection alive briefly for client to read final message
	time.Sleep(100 * time.Millisecond)
}

func countSuccess(results []service.HostResult) int {
	count := 0
	for _, r := range results {
		if r.Status == "success" {
			count++
		}
	}
	return count
}
```

- [ ] **Step 2: Add singleton to common.go**

Add to the `var` block in `backend/internal/modules/cmdb/api/common.go`:

```go
batchCmdSvcInstance *service.BatchCommandService
```

Add `batchCmdSvcInstance = nil` to the `SetDB` function.

Add getter function:

```go
func getBatchCommandService() *service.BatchCommandService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if batchCmdSvcInstance != nil {
		return batchCmdSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	batchCmdSvcInstance = service.NewBatchCommandService(cmdbDB)
	return batchCmdSvcInstance
}
```

- [ ] **Step 3: Register route**

Add to `backend/routers/v1/cmdb.go`:

```go
terminalConnectPerm := middleware.RequirePermission("cmdb:terminal", "connect")
```

(This already exists — reuse it.)

Add route inside the block:

```go
// 批量命令
g.GET("/terminal/batch", terminalConnectPerm, api.BatchCommandConnect)
```

- [ ] **Step 4: Verify compilation**

Run: `cd backend && go build ./...`
Expected: compiles

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/modules/cmdb/api/batch_command.go internal/modules/cmdb/api/common.go routers/v1/cmdb.go
git commit -m "feat(cmdb): add batch command WebSocket API with streaming results"
```

---

### Task 3: Batch Command Frontend

**Files:**
- Create: `frontend/src/api/cmdb/batch_command.js`
- Create: `frontend/src/views/Cmdb/BatchCommand.vue`
- Modify: `frontend/src/router/index.js`

- [ ] **Step 1: Create API helper**

Create `frontend/src/api/cmdb/batch_command.js`:

```js
import { getTerminalWsBaseUrl } from './terminal'

const joinBasePath = (basePath, path) => {
  const normalizedBase = (basePath || '').replace(/\/$/, '')
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${normalizedBase}${normalizedPath}`
}

export const getBatchCommandWsUrl = () => joinBasePath(getTerminalWsBaseUrl(), '/cmdb/terminal/batch')
```

- [ ] **Step 2: Create BatchCommand page**

Create `frontend/src/views/Cmdb/BatchCommand.vue`:

```vue
<template>
  <div class="page-container">
    <div class="page-header">
      <h3>批量命令</h3>
    </div>

    <el-row :gutter="16">
      <!-- Left: Host selection + command input -->
      <el-col :span="10">
        <el-card shadow="never">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span>选择主机</span>
              <el-button type="primary" size="small" @click="showHostSelector = true">
                添加主机 (已选 {{ selectedHosts.length }})
              </el-button>
            </div>
          </template>

          <div class="selected-hosts" v-if="selectedHosts.length">
            <el-tag
              v-for="host in selectedHosts"
              :key="host.id"
              closable
              @close="removeHost(host)"
              style="margin: 2px;"
            >
              {{ host.hostname || host.ip }}
            </el-tag>
          </div>
          <el-empty v-else description="请选择主机" :image-size="40" />
        </el-card>

        <el-card shadow="never" style="margin-top: 16px;">
          <template #header><span>命令</span></template>
          <el-input
            v-model="command"
            type="textarea"
            :rows="6"
            placeholder="输入要执行的命令..."
            :disabled="executing"
            style="font-family: monospace;"
          />
          <div style="margin-top: 12px; display: flex; justify-content: space-between; align-items: center;">
            <el-input-number v-model="timeout" :min="5" :max="300" :step="5" size="small" style="width: 160px;" />
            <span style="font-size: 12px; color: #909399; margin-left: 8px;">超时(秒)</span>
            <el-button
              type="primary"
              @click="executeBatch"
              :loading="executing"
              :disabled="!command || !selectedHosts.length"
            >
              {{ executing ? '执行中...' : '执行' }}
            </el-button>
          </div>
        </el-card>
      </el-col>

      <!-- Right: Results -->
      <el-col :span="14">
        <el-card shadow="never">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span>执行结果</span>
              <span v-if="results.length" style="font-size: 12px; color: #909399;">
                {{ successCount }}/{{ results.length }} 成功
              </span>
            </div>
          </template>
          <div class="results-panel">
            <div v-for="r in results" :key="r.hostId" class="result-item">
              <div class="result-header">
                <el-tag :type="r.status === 'success' ? 'success' : r.status === 'running' ? 'warning' : 'danger'" size="small">
                  {{ r.hostName || r.hostIp }}
                </el-tag>
                <span class="result-status">{{ statusText(r.status) }}</span>
              </div>
              <pre class="result-output">{{ r.output || r.error || '等待中...' }}</pre>
            </div>
            <el-empty v-if="!results.length && !executing" description="执行命令后查看结果" :image-size="60" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Host Selector Dialog -->
    <el-dialog v-model="showHostSelector" title="选择主机" width="70%">
      <el-table
        ref="hostTableRef"
        :data="hosts"
        @selection-change="handleSelectionChange"
        stripe
        max-height="400"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column prop="hostname" label="主机名" width="150" />
        <el-table-column prop="ip" label="IP" width="130" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'info'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="osName" label="系统" />
      </el-table>
      <template #footer>
        <el-button @click="showHostSelector = false">取消</el-button>
        <el-button type="primary" @click="confirmHostSelection">确认 ({{ tempSelection.length }})</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getHostList } from '@/api/cmdb/host'
import { getBatchCommandWsUrl } from '@/api/cmdb/batch_command'

const command = ref('')
const timeout = ref(30)
const executing = ref(false)
const selectedHosts = ref([])
const tempSelection = ref([])
const results = ref([])
const showHostSelector = ref(false)
const hosts = ref([])
const hostTableRef = ref(null)

const successCount = computed(() => results.value.filter(r => r.status === 'success').length)

const statusText = (s) => {
  const map = { running: '执行中...', success: '成功', failed: '失败', timeout: '超时' }
  return map[s] || s
}

const fetchHosts = async () => {
  try {
    const res = await getHostList({ page: 1, pageSize: 1000 })
    hosts.value = res.data?.list || res.data || []
  } catch (e) {
    ElMessage.error('获取主机列表失败')
  }
}

const handleSelectionChange = (selection) => {
  tempSelection.value = selection
}

const confirmHostSelection = () => {
  selectedHosts.value = [...tempSelection.value]
  showHostSelector.value = false
}

const removeHost = (host) => {
  selectedHosts.value = selectedHosts.value.filter(h => h.id !== host.id)
}

const executeBatch = () => {
  if (!command.value || !selectedHosts.value.length) return

  executing.value = true
  results.value = selectedHosts.value.map(h => ({
    hostId: h.id,
    hostName: h.hostname || h.ip,
    hostIp: h.ip,
    status: 'running',
    output: '',
    error: ''
  }))

  const ws = new WebSocket(getBatchCommandWsUrl())

  ws.onopen = () => {
    ws.send(JSON.stringify({
      hostIds: selectedHosts.value.map(h => h.id),
      command: command.value,
      timeout: timeout.value
    }))
  }

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data)

    if (msg.type === 'host_result') {
      const idx = results.value.findIndex(r => r.hostId === msg.data.hostId)
      if (idx >= 0) {
        results.value[idx] = { ...results.value[idx], ...msg.data }
      }
    } else if (msg.type === 'complete') {
      executing.value = false
      ElMessage.success(`执行完成: ${msg.success} 成功, ${msg.failed} 失败`)
      ws.close()
    } else if (msg.type === 'error') {
      executing.value = false
      ElMessage.error(msg.message)
      ws.close()
    }
  }

  ws.onerror = () => {
    executing.value = false
    ElMessage.error('WebSocket 连接失败')
  }

  ws.onclose = () => {
    executing.value = false
  }
}

onMounted(fetchHosts)
</script>

<style scoped>
.page-container { padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }

.selected-hosts { max-height: 120px; overflow-y: auto; }

.results-panel { max-height: 65vh; overflow-y: auto; }
.result-item { margin-bottom: 12px; border: 1px solid #ebeef5; border-radius: 4px; overflow: hidden; }
.result-header { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; background: #f5f7fa; }
.result-status { font-size: 12px; color: #909399; }
.result-output { margin: 0; padding: 10px 12px; font-family: 'Consolas', 'Monaco', monospace; font-size: 12px; background: #1e1e1e; color: #d4d4d4; max-height: 200px; overflow-y: auto; white-space: pre-wrap; word-break: break-all; }
</style>
```

- [ ] **Step 3: Add route**

In `frontend/src/router/index.js`, add after the `cmdb/files` route:

```js
{
  path: 'cmdb/batch-command',
  component: () => import('../views/Cmdb/BatchCommand.vue')
},
```

- [ ] **Step 4: Test batch command**

Run: `cd frontend && npm run dev`

Test flow:
1. Open `/cmdb/batch-command`
2. Click "添加主机", select multiple hosts, confirm
3. Type a command (e.g., `hostname && uptime`)
4. Click "执行"
5. Verify results stream in for each host
6. Verify audit records created in terminal session list

- [ ] **Step 5: Commit**

```bash
cd frontend
git add src/api/cmdb/batch_command.js src/views/Cmdb/BatchCommand.vue src/router/index.js
git commit -m "feat(cmdb): add batch command page with host selection and streaming results"
```

---

## Self-Review Checklist

1. **Spec coverage:** Select multiple hosts — Task 3. Execute command on all — Task 1+2. Real-time output via WebSocket — Task 2+3. Per-host results panel — Task 3. Timeout configuration — Task 1+3. Audit records — Task 1. All covered.

2. **Placeholder scan:** No TBD/TODO found. All code complete.

3. **Type consistency:** `HostResult` struct fields match between service and API handler. Frontend `results` array uses matching field names (`hostId`, `hostName`, `hostIp`, `status`, `output`, `error`). WebSocket message types (`host_result`, `complete`, `error`) consistent between backend and frontend.
