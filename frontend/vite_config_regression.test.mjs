import test from 'node:test'
import assert from 'node:assert/strict'
import config from './vite.config.js'

test('vite manualChunks uses function form for rolldown compatibility', () => {
  const manualChunks = config.build?.rollupOptions?.output?.manualChunks
  assert.equal(typeof manualChunks, 'function')
})

test('vite manualChunks groups vendor modules as expected', () => {
  const manualChunks = config.build?.rollupOptions?.output?.manualChunks
  assert.equal(typeof manualChunks, 'function')

  assert.equal(manualChunks('/project/node_modules/element-plus/es/index.mjs'), 'element-plus')
  assert.equal(manualChunks('/project/node_modules/vue/dist/vue.runtime.esm-bundler.js'), 'vue-vendor')
  assert.equal(manualChunks('/project/node_modules/vue-router/dist/vue-router.mjs'), 'vue-vendor')
  assert.equal(manualChunks('/project/node_modules/pinia/dist/pinia.mjs'), 'vue-vendor')
  assert.equal(manualChunks('/project/src/main.js'), undefined)
})
