<script setup lang="ts">
definePageMeta({ layout: false })

const { setToken, api } = useWzap()
const tokenInput = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  loading.value = true
  error.value = ''
  setToken(tokenInput.value)

  try {
    await api('/sessions')
    navigateTo('/')
  } catch {
    error.value = 'Invalid token or API unreachable'
    setToken('')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-(--ui-bg)">
    <div class="w-full max-w-sm">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-(--ui-text-highlighted)">
          wzap
        </h1>
        <p class="text-sm text-(--ui-text-muted) mt-1">
          WhatsApp API Manager
        </p>
      </div>

      <UCard>
        <form class="space-y-4" @submit.prevent="handleLogin">
          <UFormField label="Admin Token">
            <UInput
              v-model="tokenInput"
              type="password"
              placeholder="Your API key"
              class="w-full"
            />
          </UFormField>

          <p v-if="error" class="text-sm text-red-500">
            {{ error }}
          </p>

          <UButton
            type="submit"
            block
            :loading="loading"
            color="primary"
          >
            Login
          </UButton>
        </form>
      </UCard>
    </div>
  </div>
</template>
