<template>
  <AppLayout>
    <main class="wb-page">
      <header class="wb-hero">
        <span class="wb-label">WORKBUDDY × {{ app.siteName }} · WINDOWS 预览版</span>
        <h1>把本站模型，接入你的 WorkBuddy。</h1>
        <p>安装官方客户端，选择密钥和模型，运行一次配置助手。配置完成后，直接在 WorkBuddy 中使用。</p>
        <div class="wb-actions">
          <a class="wb-button primary" href="https://www.codebuddy.cn/work/" target="_blank" rel="noopener noreferrer">① 下载官方 WorkBuddy ↗</a>
          <a class="wb-button secondary" href="#wb-config">我已安装，开始配置 ↓</a>
        </div>
        <p class="wb-caption">首版适配 Windows x64 / 国内版 WorkBuddy 5.6.2。官方登录由用户完成，配置助手用完即退出。</p>
      </header>

      <div class="wb-grid">
        <section id="wb-config" class="wb-card">
          <span class="wb-step">02 / CONNECT</span>
          <h2>选择要接入的模型</h2>
          <p>建议在密钥管理中新建一把 WorkBuddy 专用密钥，设置适当额度，方便单独撤销。</p>
          <RouterLink to="/keys" class="wb-link">前往密钥管理 →</RouterLink>
          <form @submit.prevent="createPair">
            <label for="wb-key">本站 API 密钥</label>
            <select id="wb-key" v-model="keyId" :disabled="loading || busy">
              <option value="">{{ loading ? '正在加载密钥…' : '选择一把可用密钥' }}</option>
              <option v-for="key in keys" :key="key.id" :value="key.id">{{ key.name }} · {{ key.group?.name }}</option>
            </select>
            <p class="wb-caption">仅显示有效的 OpenAI 分组密钥。其他分组暂未适配。</p>
            <p v-if="!loading && keys.length === 0" class="wb-notice">暂无可用密钥，请先创建并绑定 OpenAI 分组。</p>
            <label for="wb-model">模型 ID</label>
            <input id="wb-model" v-model.trim="model" maxlength="200" placeholder="填写此密钥可调用的完整模型 ID" required :disabled="busy" autocomplete="off" />
            <p class="wb-caption">模型需支持流式工具调用。助手会实际测试后再写入配置；图片输入暂不开启。</p>
            <label class="wb-consent"><input v-model="consent" type="checkbox" :disabled="busy" />允许助手接收所选密钥，并发起一次可能计费的小型模型测试。</label>
            <button class="wb-button primary" type="submit" :disabled="!canPair">{{ busy ? '正在生成…' : '生成一次性配对码' }}</button>
          </form>
          <p v-if="error" class="wb-error" role="alert">{{ error }}</p>
        </section>

        <section class="wb-card">
          <span class="wb-step">03 / SET UP</span>
          <h2>运行配置助手</h2>
          <ol>
            <li>下载助手，退出 WorkBuddy（含系统托盘），双击运行。</li>
            <li>选择 <code>setup</code>，输入本站地址和下方配对码。</li>
            <li>核对模型和接口地址，确认后开始测试、备份和配置。</li>
            <li>打开 WorkBuddy，选择<strong>刚刚配置的本站模型</strong>。</li>
          </ol>
          <a v-if="helperAvailable" class="wb-button secondary" href="/downloads/sub2api-workbuddy-setup-windows-amd64.exe" download>下载 Windows 配置助手 ↓</a>
          <p v-else class="wb-notice">{{ helperChecking ? '正在检查助手下载…' : '本站尚未发布配置助手，请联系管理员完成构建。' }}</p>
          <div class="wb-site"><span>本站地址</span><code>{{ siteOrigin }}</code><button type="button" @click="copy(siteOrigin, 'site')">{{ copied === 'site' ? '已复制' : '复制' }}</button></div>
          <div v-if="pairCode && remaining > 0" class="wb-pair" aria-live="polite">
            <div class="wb-pair-title"><strong>一次性配对码</strong><span>{{ remaining }} 秒内有效</span></div>
            <code>{{ pairCode }}</code>
            <button class="wb-button secondary" type="button" @click="copy(pairCode, 'code')">{{ copied === 'code' ? '已复制' : '复制配对码' }}</button>
            <p class="wb-caption">配对码可兑换你的密钥，请勿转发。成功兑换后即失效；重新生成不会提前撤销旧码。</p>
            <p class="wb-caption">将连接：{{ endpoint }}</p>
          </div>
          <p v-else class="wb-notice">{{ expired ? '配对码已过期，请重新生成。' : '完成左侧选择后，配对码会显示在这里。' }}</p>
        </section>
      </div>
      <footer class="wb-footer">
        <details><summary>配置失败或想恢复原设置？</summary><p>助手会在测试通过后备份并写入配置。重新运行选择 restore 可恢复最近一次配置；如文件之后被修改，助手会停止自动恢复以保留新改动。备份位于 %USERPROFILE%\.workbuddy\sub2api-setup，其中可能包含密钥，请妥善保管。</p></details>
        <details><summary>其他版本、macOS 或没有兼容模型？</summary><p>可在 WorkBuddy 的「设置 → 模型 → 自定义」中手动填写本站接口、个人密钥和模型 ID。首版不会为未验证的版本自动改写配置，也不会安装额外的常驻切换工具。</p></details>
      </footer>
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores/app'
import { apiClient } from '@/api/client'
import { list } from '@/api/keys'
import type { ApiKey } from '@/types'

const app = useAppStore()
const keys = ref<ApiKey[]>([])
const keyId = ref<number | ''>('')
const model = ref('')
const consent = ref(false)
const loading = ref(true)
const busy = ref(false)
const error = ref('')
const pairCode = ref('')
const endpoint = ref('')
const deadline = ref(0)
const now = ref(Date.now())
const expired = ref(false)
const copied = ref('')
const helperAvailable = ref(false)
const helperChecking = ref(true)
const siteOrigin = window.location.origin
let active = true
const requests = new AbortController()
const remaining = computed(() => Math.max(0, Math.ceil((deadline.value - now.value) / 1000)))
const canPair = computed(() => !loading.value && !busy.value && helperAvailable.value && !!keyId.value && /^[a-zA-Z0-9][a-zA-Z0-9._:/-]{0,199}$/.test(model.value) && consent.value)
const timer = window.setInterval(() => {
  now.value = Date.now()
  if (pairCode.value && remaining.value === 0) { pairCode.value = ''; expired.value = true }
}, 1000)
watch([keyId, model], () => { pairCode.value = ''; copied.value = ''; expired.value = false })

async function copy(value: string, kind: string) {
  try { await navigator.clipboard.writeText(value); copied.value = kind }
  catch { error.value = '复制失败，请手动选择并复制。' }
}
async function createPair() {
  if (!canPair.value) return
  error.value = ''; busy.value = true; pairCode.value = ''; copied.value = ''
  try {
    const { data } = await apiClient.post<{ code: string; expires_in: number; endpoint: string }>('/workbuddy/pair', { key_id: keyId.value, model: model.value }, { signal: requests.signal })
    if (!active) return
    pairCode.value = data.code; endpoint.value = data.endpoint
    now.value = Date.now(); deadline.value = now.value + data.expires_in * 1000; expired.value = false
  } catch { if (active) error.value = '配对失败：请检查密钥状态、登录状态，以及管理员是否已配置 HTTPS API Base URL。' }
  finally { busy.value = false }
}

onMounted(async () => {
  const download = fetch('/downloads/sub2api-workbuddy-setup-windows-amd64.exe', { headers: { Range: 'bytes=0-1' }, signal: requests.signal })
    .then(async (res) => {
      // SPA fallback may return HTML with 200. Check executable magic, not just status.
      const reader = res.body?.getReader()
      if (!reader) return
      try {
        const { value } = await reader.read()
        helperAvailable.value = res.ok && !!value && value[0] === 0x4d && value[1] === 0x5a
      } finally { await reader.cancel() }
    }).catch(() => { helperAvailable.value = false }).finally(() => { helperChecking.value = false })
  try {
    let page = 1
    const all: ApiKey[] = []
    while (active) {
      const data = await list(page, 100, { status: 'active' }, { signal: requests.signal })
      all.push(...data.items)
      if (page >= data.pages || data.items.length === 0) break
      page++
    }
    if (active) keys.value = all.filter(k => k.group?.platform === 'openai' && k.group.status === 'active' && (!k.expires_at || Date.parse(k.expires_at) > Date.now()) && (k.quota <= 0 || k.quota_used < k.quota))
  } catch { if (active) error.value = '密钥加载失败，请刷新页面重试。' }
  finally { loading.value = false }
  await download
})
onUnmounted(() => { active = false; requests.abort(); window.clearInterval(timer); pairCode.value = '' })
</script>

<style scoped>
.wb-page{max-width:1080px;margin:0 auto;padding:16px 8px 48px;color:#162c34}.wb-hero{background:#102b2a;color:#effcf7;padding:48px;border-radius:24px;background-image:radial-gradient(ellipse at top right,#22584d,transparent 65%)}.wb-label,.wb-step{font-size:11px;letter-spacing:.12em;font-weight:700}.wb-label{color:#87d9b8}.wb-hero h1{font-size:clamp(28px,4vw,42px);line-height:1.25;letter-spacing:-.03em;margin:22px 0 18px;font-weight:700}.wb-hero p{color:#b6ccc5;max-width:660px;line-height:1.85}.wb-actions{display:flex;flex-wrap:wrap;gap:12px;margin:26px 0 16px}.wb-button{display:inline-flex;justify-content:center;align-items:center;padding:12px 18px;border-radius:10px;font-size:14px;font-weight:600;text-decoration:none;cursor:pointer;border:1px solid transparent;min-height:46px}.primary{background:#93e6bd;color:#10372c}.secondary{background:#f4f8f6;color:#234c40;border-color:#d9e7df}.wb-button:disabled{opacity:.45;cursor:not-allowed}.wb-button:hover:not(:disabled){filter:brightness(.96)}.wb-caption{font-size:12px!important;line-height:1.7;margin:10px 0;color:#63736b}.wb-grid{display:grid;grid-template-columns:1fr 1fr;gap:22px;margin-top:24px}.wb-card{background:white;border:1px solid #dfe7e2;border-radius:20px;padding:30px;min-width:0}.wb-step{color:#48836e}.wb-card h2{font-size:23px;font-weight:650;margin:16px 0 10px}.wb-card>p,.wb-card li{font-size:14px;line-height:1.85;color:#596c63}.wb-card ol{padding-left:22px;margin:24px 0}.wb-card li{padding:0 0 13px 5px}.wb-card form{margin-top:24px}.wb-card label{display:block;font-size:14px;font-weight:600;margin:20px 0 9px}.wb-card select,.wb-card input:not([type=checkbox]){width:100%;border:1px solid #cbd9d0;border-radius:9px;padding:12px;background:#fff;color:#19362a;font-size:14px;min-height:46px}.wb-consent{display:flex!important;gap:10px;font-size:12px!important;line-height:1.8;font-weight:400!important;margin:20px 0!important}.wb-consent input{margin-top:4px;flex-shrink:0;width:16px;height:16px;accent-color:#227b59}.wb-link{color:#217653;font-size:13px;text-decoration:underline;text-underline-offset:4px}.wb-notice{background:#f4f7f5;padding:16px;border-radius:10px;font-size:13px;line-height:1.7;margin:20px 0}.wb-site{display:flex;align-items:center;flex-wrap:wrap;gap:10px;margin:24px 0;font-size:12px}.wb-site code{overflow-wrap:anywhere}.wb-site button{margin-left:auto;color:#287452;font-weight:600}.wb-pair{background:#f0faf4;border:1px solid #c5e3ce;border-radius:12px;padding:18px}.wb-pair-title{display:flex;justify-content:space-between;gap:10px;font-size:13px}.wb-pair-title span{color:#557c62}.wb-pair>code{display:block;word-break:break-all;margin:18px 0;font-size:13px;letter-spacing:.03em}.wb-error{color:#b42318!important;margin:16px 0;line-height:1.7;font-size:13px}.wb-footer{margin:26px 0}.wb-footer details{border-bottom:1px solid #dce6de;padding:20px 4px;font-size:14px}.wb-footer summary{cursor:pointer;font-weight:600}.wb-footer p{line-height:1.9;color:#63736b;margin:14px 0;overflow-wrap:anywhere}:focus-visible{outline:3px solid #40a683;outline-offset:3px}:global(.dark) .wb-page{color:#e2ebe5}:global(.dark) .wb-card{background:#182a25;border-color:#3d5047}:global(.dark) .wb-card p,:global(.dark) .wb-card li{color:#b3c4ba}:global(.dark) .wb-notice,:global(.dark) .wb-pair{background:#243c31;color:#d7e9de}:global(.dark) .wb-footer p{color:#b3c4ba}@media(max-width:760px){.wb-grid{grid-template-columns:1fr}.wb-hero{padding:30px 22px}.wb-card{padding:24px 20px}.wb-page{padding:6px 0 28px}.wb-actions{flex-direction:column}.wb-pair-title{flex-wrap:wrap}}
</style>
