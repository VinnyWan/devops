import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'

const nodeListPath = new URL('./src/views/k8s/NodeList.vue', import.meta.url)
const nodeDetailPath = new URL('./src/views/k8s/NodeDetail.vue', import.meta.url)

test('NodeList uses backend request field name for cordon/drain actions', async () => {
  const content = await readFile(nodeListPath, 'utf8')

  assert.match(content, /cordonNode\(\{\s*clusterName:\s*clusterName\.value,\s*name:\s*row\.name\s*\}\)/)
  assert.match(content, /drainNode\(\{\s*clusterName:\s*clusterName\.value,\s*name:\s*row\.name\s*\}\)/)
  assert.doesNotMatch(content, /nodeName:\s*row\.name/)
})

test('NodeDetail quick metrics uses filtered pod total to match Pods tab', async () => {
  const content = await readFile(nodeDetailPath, 'utf8')

  assert.match(content, /Pod 数<\/span>[\s\S]*?\{\{\s*podItems\.length\s*\}\}\s*\/\s*\{\{\s*formatValue\(detail\.podCapacity\)\s*\}\}/)
  assert.doesNotMatch(content, /detail\.podCount\s*\?\?\s*detail\.pods\?\.total/)
})
