# Swagger 文档使用说明

## 快速开始

### 后端：重新生成文档

```bash
cd backend
make swag
```

这会自动执行：
1. 生成 Swagger 2.0 文档到 `docs/swagger/`
2. 转换为 OpenAPI 3.0 到 `docs/openapi/`

### 前端：更新类型定义

```bash
cd frontend
npm run generate:contract
```

生成的类型定义位于：`src/types/generated/api.types.ts`

## 访问 Swagger UI

启动后端服务后访问：
```
http://localhost:8000/swagger/index.html
```

## 文件位置

- **Swagger 2.0**: `backend/docs/swagger/swagger.json`
- **OpenAPI 3.0**: `backend/docs/openapi/openapi.json`
- **前端类型**: `frontend/src/types/generated/api.types.ts`

## 工作流程

1. 后端开发者在 API handler 中添加 Swagger 注解
2. 运行 `make swag` 生成文档
3. 前端运行 `npm run generate:contract` 更新类型
4. 前端使用生成的类型进行开发
