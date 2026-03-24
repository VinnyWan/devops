#!/usr/bin/env node
import { execSync } from 'child_process'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const rootDir = path.resolve(__dirname, '..')
const swaggerPath = path.resolve(rootDir, '../backend/docs/openapi/openapi.json')
const outputFile = path.resolve(rootDir, 'src/types/generated/api.types.ts')

console.log('🚀 开始生成 TypeScript 类型定义...')
console.log(`📄 OpenAPI 文件: ${swaggerPath}`)
console.log(`📁 输出文件: ${outputFile}`)

try {
  execSync(`npx openapi-typescript "${swaggerPath}" -o "${outputFile}"`, {
    stdio: 'inherit',
    cwd: rootDir
  })
  console.log('✅ 类型定义生成完成！')
} catch (error) {
  console.error('❌ 生成失败:', error.message)
  process.exit(1)
}
