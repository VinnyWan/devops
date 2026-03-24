<script setup lang="ts">
import { ref, watch } from 'vue'
import { NForm, NFormItem, NInput, NSelect, NRadioGroup, NRadio, NButton, NSpace } from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'

interface FieldConfig {
  name: string
  label: string
  type: 'text' | 'email' | 'password' | 'select' | 'radio' | 'textarea'
  required?: boolean
  options?: { label: string; value: any }[]
  rules?: any[]
}

interface Props {
  fields: FieldConfig[]
  initialData?: Record<string, any>
  onSubmit: (data: Record<string, any>) => Promise<void>
}

const props = defineProps<Props>()

const formRef = ref<FormInst | null>(null)
const formData = ref<Record<string, any>>({})
const loading = ref(false)

watch(() => props.initialData, (newData) => {
  if (newData) {
    formData.value = { ...newData }
  }
}, { immediate: true })

const rules: FormRules = {}
props.fields.forEach(field => {
  const fieldRules: any[] = []
  if (field.required) {
    fieldRules.push({ required: true, message: `${field.label}不能为空` })
  }
  if (field.type === 'email') {
    fieldRules.push({ type: 'email', message: '请输入有效的邮箱地址' })
  }
  if (field.rules) {
    fieldRules.push(...field.rules)
  }
  if (fieldRules.length > 0) {
    rules[field.name] = fieldRules
  }
})

const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
    loading.value = true
    await props.onSubmit(formData.value)
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  formData.value = props.initialData ? { ...props.initialData } : {}
}
</script>

<template>
  <NForm ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="100">
    <NFormItem v-for="field in fields" :key="field.name" :label="field.label" :path="field.name">
      <NInput v-if="field.type === 'text' || field.type === 'password'"
        v-model:value="formData[field.name]"
        :type="field.type"
        :placeholder="`请输入${field.label}`" />
      <NInput v-else-if="field.type === 'email'"
        v-model:value="formData[field.name]"
        type="text"
        :placeholder="`请输入${field.label}`" />
      <NInput v-else-if="field.type === 'textarea'"
        v-model:value="formData[field.name]"
        type="textarea"
        :placeholder="`请输入${field.label}`" />
      <NSelect v-else-if="field.type === 'select'"
        v-model:value="formData[field.name]"
        :options="field.options"
        :placeholder="`请选择${field.label}`" />
      <NRadioGroup v-else-if="field.type === 'radio'" v-model:value="formData[field.name]">
        <NRadio v-for="opt in field.options" :key="opt.value" :value="opt.value">
          {{ opt.label }}
        </NRadio>
      </NRadioGroup>
    </NFormItem>
    <NFormItem>
      <NSpace>
        <NButton type="primary" :loading="loading" @click="handleSubmit">提交</NButton>
        <NButton @click="handleReset">重置</NButton>
      </NSpace>
    </NFormItem>
  </NForm>
</template>
