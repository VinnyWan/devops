import js from '@eslint/js'
import eslintConfigPrettier from 'eslint-config-prettier'
import pluginVue from 'eslint-plugin-vue'
import tseslint from 'typescript-eslint'
import vueParser from 'vue-eslint-parser'

export default tseslint.config(
  { ignores: ['dist/**', 'node_modules/**'] },
  js.configs.recommended,
  {
    files: ['scripts/**/*.mjs', '*.config.js', '*.config.ts'],
    languageOptions: {
      globals: {
        URL: 'readonly',
        console: 'readonly',
        process: 'readonly',
      },
    },
  },
  ...tseslint.configs.recommended,
  {
    files: ['**/*.ts'],
    languageOptions: {
      globals: {
        Blob: 'readonly',
        HTMLDivElement: 'readonly',
        URL: 'readonly',
        WebSocket: 'readonly',
        clearInterval: 'readonly',
        clearTimeout: 'readonly',
        document: 'readonly',
        localStorage: 'readonly',
        setInterval: 'readonly',
        setTimeout: 'readonly',
        window: 'readonly',
      },
    },
  },
  ...pluginVue.configs['flat/recommended'],
  {
    files: ['**/*.vue'],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tseslint.parser,
        extraFileExtensions: ['.vue'],
      },
      globals: {
        Blob: 'readonly',
        HTMLDivElement: 'readonly',
        URL: 'readonly',
        WebSocket: 'readonly',
        clearInterval: 'readonly',
        clearTimeout: 'readonly',
        document: 'readonly',
        localStorage: 'readonly',
        setInterval: 'readonly',
        setTimeout: 'readonly',
        window: 'readonly',
      },
    },
  },
  eslintConfigPrettier,
  {
    files: ['**/*.{ts,vue}'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
    },
  },
)
