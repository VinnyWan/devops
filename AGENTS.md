# Agents.md / CLAUDE.md（Claude Code 代理配置模板）

# 对话要求
- 全部对话必须以中文进行展示
- 所有的输出和改动必现简洁明了

## 项目规划
- backend目录为后端 golang 项目 gin的web框架项目
- frontend目录为前端 vue3 + element-plus 项目
- 后端backend在每次进行代码修改之后，都需要生成swagger文档，swagger的目录在devops/backend/docs/swagger中

## 技术栈详情

### 后端 (Backend)
- **语言**: Go 1.24.0
- **Web框架**: Gin
- **ORM**: GORM (支持 MySQL 和 SQLite)
- **API文档**: Swaggo (Swagger/OpenAPI)
- **认证**: OAuth2, OIDC, LDAP
- **K8s客户端**: client-go v0.31.4
- **配置管理**: Viper, Nacos
- **日志**: Zap + Lumberjack
- **缓存**: Redis (go-redis/v9)
- **权限**: Casbin

### 前端 (Frontend)
- **框架**: Vue 3.5.30
- **UI库**: Element Plus 2.13.6
- **状态管理**: Pinia 3.0.4
- **路由**: Vue Router 4.6.4
- **HTTP客户端**: Axios 1.14.0
- **构建工具**: Vite 8.0.1
- **自动导入**: unplugin-auto-import, unplugin-vue-components

## 项目结构

### 后端目录结构
```
backend/
├── cmd/server/          # 主程序入口
├── config/              # 配置文件和配置加载
├── internal/
│   ├── bootstrap/       # 初始化逻辑 (DB, Redis, K8s, Casbin)
│   ├── middleware/      # 中间件 (认证, 审计, 权限, CORS)
│   ├── modules/         # 业务模块
│   │   ├── user/        # 用户、角色、部门、权限、审计
│   │   ├── k8s/         # K8s资源管理 (集群、节点、工作负载)
│   │   ├── app/         # 应用管理
│   │   ├── alert/       # 告警中心
│   │   ├── cicd/        # CI/CD
│   │   ├── harbor/      # Harbor镜像仓库
│   │   ├── log/         # 日志中心
│   │   └── monitor/     # 监控中心
│   └── pkg/             # 公共工具包
├── routers/             # 路由定义
├── docs/                # API文档 (Swagger/OpenAPI)
└── scripts/             # 脚本工具
```

### 前端目录结构
```
frontend/
├── src/
│   ├── api/             # API请求封装
│   ├── components/      # 公共组件
│   │   ├── K8s/         # K8s相关组件
│   │   └── Layout/      # 布局组件
│   ├── views/           # 页面视图
│   │   ├── k8s/         # K8s管理页面
│   │   ├── System/      # 系统管理
│   │   ├── Dashboard/   # 仪表盘
│   │   └── Login/       # 登录
│   ├── stores/          # Pinia状态管理
│   ├── router/          # 路由配置
│   └── utils/           # 工具函数
└── public/              # 静态资源
```

## Workflow Orchestration（工作流编排）

### 1. Plan Node Default（默认计划节点）
- 任何非琐碎任务（3+ 步或架构决策）必须进入计划模式。
- 一旦出现偏差，立即停止并重新规划，不要继续硬推。
- 验证步骤也要使用计划模式，而非仅用于构建。
- 提前写详细规格，减少歧义。

### 2. Subagent Strategy（子代理策略）
- 大量使用子代理，保持主上下文窗口干净。
- 将研究、探索、并行分析全部外包给子代理。
- 复杂问题时，通过子代理投入更多算力。
- 每个子代理只专注一个方向。

### 3. Self-Improvement Loop（自我改进循环）
- 用户任何一次纠正后，立即更新 `tasks/lessons.md` 并记录模式。
- 写规则防止自己重复犯错。
- 无情迭代 lessons，直到错误率下降。
- 每次会话开始时，先复习项目相关 lessons。

### 4. Verification Before Done（完成前验证）
- 绝不在证明它能工作前标记任务完成。
- 必要时对比主分支与修改行为。
- 自问：“资深工程师会批准吗？”
- 运行测试、查日志、展示正确性。

### 5. Demand Elegance (Balanced)（要求优雅但平衡）
- 非琐碎改动时暂停：“有没有更优雅的方式？”
- 如果修复感觉 hacky：“基于我现在的一切知识，实现优雅方案。”
- 简单问题不要过度工程化。
- 每次呈现前先挑战自己的工作。

### 6. Autonomous Bug Fixing（自主 Bug 修复）
- 收到 bug 报告后直接修复，无需用户手把手。
- 指向日志、错误、失败测试，然后解决。
- 用户无需上下文切换。
- 自动修复失败的 CI 测试。

## Task Management（任务管理）
1. **先规划**：将计划写入 `tasks/todo.md`，使用可勾选清单。
2. **验证计划**：实现前先 check-in。
3. **跟踪进度**：每完成一项即标记。
4. **解释变更**：每步提供高层总结。
5. **记录结果**：在 `tasks/todo.md` 末尾添加 review 部分。
6. **捕捉教训**：纠正后更新 `tasks/lessons.md`。

## Core Principles（核心原则）
- **简洁优先**：每次变更尽量简单，只影响最小代码。
- **绝不偷懒**：找到根因，不用临时修复，坚持资深开发者标准。
- **最小影响**：只修改必要部分，避免引入新 bug。

## 开发规范

### 后端开发规范
1. **代码组织**：每个模块按 `api -> service -> repository -> model` 分层
2. **API注释**：所有API handler必须添加Swagger注释 (`@Summary`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router`)
3. **错误处理**：使用 `internal/pkg/obserr` 统一错误处理
4. **测试**：service层需要编写单元测试 (`*_test.go`)
5. **Swagger生成**：修改API后执行 `swag init` 或 `make swagger` 更新文档
6. **命名规范**：
   - 文件名：小写下划线 (`user_service.go`)
   - 结构体：大驼峰 (`UserService`)
   - 方法：大驼峰导出，小驼峰私有
   - 常量：全大写下划线 (`MAX_RETRY_COUNT`)

### 前端开发规范
1. **组件命名**：大驼峰 (`ClusterSelector.vue`)
2. **API封装**：所有后端请求封装在 `src/api/` 目录
3. **状态管理**：全局状态使用Pinia stores
4. **样式**：使用 Element Plus 主题变量，避免硬编码颜色
5. **路由**：懒加载页面组件
6. **类型安全**：尽量使用明确的类型定义

### Git提交规范
- `feat`: 新功能
- `fix`: Bug修复
- `refactor`: 重构
- `docs`: 文档更新
- `style`: 代码格式调整
- `test`: 测试相关
- `chore`: 构建/工具链相关

## 常用命令

### 后端
```bash
cd backend
go mod tidy                    # 整理依赖
go test ./...                  # 运行所有测试
swag init                      # 生成Swagger文档
go run cmd/server/main.go      # 启动服务
```

### 前端
```bash
cd frontend
npm install                    # 安装依赖
npm run dev                    # 启动开发服务器
npm run build                  # 构建生产版本
```

