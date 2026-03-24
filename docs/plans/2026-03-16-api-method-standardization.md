# API方法规范化实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将后端API中的查询接口从POST改为GET方法，更新路由定义、Swagger注释并重新生成文档

**Architecture:** 修改路由定义文件和API handler的Swagger注释，确保查询接口使用GET方法，增删改接口使用POST方法

**Tech Stack:** Go, Gin框架, Swaggo

---

## 背景

经过全面扫描，发现后端项目中只有1个接口需要修改：
- `/user/permissions` - 当前使用POST，应改为GET（这是获取用户权限列表的查询接口）

## 任务列表

### Task 1: 修改路由定义

**Files:**
- Modify: `backend/routers/v1/user.go:14`

**Step 1: 修改路由方法从POST改为GET**

将第14行：
```go
r.POST("/user/permissions", api.GetUserPermissions)
```

改为：
```go
r.GET("/user/permissions", api.GetUserPermissions)
```

**Step 2: 验证修改**

Run: `cd backend && go build`
Expected: 编译成功，无错误

---

### Task 2: 更新Swagger注释

**Files:**
- Modify: `backend/internal/modules/user/api/user.go:461`

**Step 1: 修改Swagger Router注释**

将第461行：
```go
// @Router /user/permissions [post]
```

改为：
```go
// @Router /user/permissions [get]
```

**Step 2: 验证修改**

Run: `cd backend && go build`
Expected: 编译成功，无错误

---

### Task 3: 重新生成Swagger文档

**Files:**
- Generate: `backend/docs/swagger.json`, `backend/docs/swagger.yaml`

**Step 1: 运行swag init生成文档**

Run: `cd backend && swag init -g cmd/main.go`
Expected: 成功生成docs/swagger.json和docs/swagger.yaml

**Step 2: 验证生成的文档**

Run: `grep -A 5 '/user/permissions' backend/docs/swagger.json`
Expected: 看到该接口的method为"get"而非"post"

---

### Task 4: 重新生成前端API

**Files:**
- Generate: `frontend/src/api/generated/用户管理.api.ts`

**Step 1: 运行前端API生成脚本**

Run: `cd frontend && node scripts/generate-from-swagger.mjs`
Expected: 成功生成API文件

**Step 2: 验证生成的API方法**

Run: `grep 'userPermissions' frontend/src/api/generated/用户管理.api.ts`
Expected: 看到userPermissionsGet方法（而非userPermissionsPost）

---

### Task 5: 集成测试

**Step 1: 启动后端服务**

Run: `cd backend && go run cmd/main.go`
Expected: 服务启动成功

**Step 2: 测试GET接口**

Run: `curl -X GET http://localhost:8080/api/v1/user/permissions -H "Authorization: Bearer <token>"`
Expected: 返回200状态码和权限列表数据

**Step 3: 验证POST方法不再可用**

Run: `curl -X POST http://localhost:8080/api/v1/user/permissions -H "Authorization: Bearer <token>"`
Expected: 返回404或405错误

---
