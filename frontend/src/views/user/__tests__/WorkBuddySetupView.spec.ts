import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import WorkBuddySetupView from '../WorkBuddySetupView.vue'

const api = vi.hoisted(() => ({ list: vi.fn(), post: vi.fn() }))
vi.mock('@/api/keys', () => ({ list: api.list }))
vi.mock('@/api/client', () => ({ apiClient: { post: api.post } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ siteName: 'Test site' }) }))
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<div><slot /></div>' } }))

function render() {
  return mount(WorkBuddySetupView, { global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
}
beforeEach(() => {
  vi.clearAllMocks()
  api.list.mockResolvedValue({ items: [
    { id: 1, name: 'My WorkBuddy', group: { name: 'OpenAI', platform: 'openai', status: 'active' }, quota: 0 },
    { id: 2, name: 'Wrong platform', group: { platform: 'anthropic', status: 'active' }, quota: 0 },
    { id: 3, name: 'Expired', group: { platform: 'openai', status: 'active' }, expires_at: '2020-01-01', quota: 0 }
  ], pages: 1 })
  api.post.mockResolvedValue({ data: { code: 'a'.repeat(64), expires_in: 300, endpoint: 'https://api.example.invalid/v1/chat/completions' } })
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(new Uint8Array([0x4d, 0x5a]))))
})
afterEach(() => { vi.unstubAllGlobals(); vi.useRealTimers() })

describe('WorkBuddy setup', () => {
  it('filters keys and requires consent before issuing a code without sending the API secret', async () => {
    const wrapper = render(); await flushPromises()
    expect(wrapper.findAll('option')).toHaveLength(2)
    await wrapper.get('select').setValue(1)
    await wrapper.get('#wb-model').setValue('test-model')
    await wrapper.get('form').trigger('submit'); expect(api.post).not.toHaveBeenCalled()
    await wrapper.get('[type=checkbox]').setValue(true)
    await wrapper.get('form').trigger('submit'); await flushPromises()
    expect(api.post).toHaveBeenCalledWith('/workbuddy/pair', { key_id: 1, model: 'test-model' }, expect.any(Object))
    expect(wrapper.text()).toContain('a'.repeat(64))
    await wrapper.get('#wb-model').setValue('other-model')
    expect(wrapper.text()).not.toContain('a'.repeat(64))
    wrapper.unmount()
  })

  it('rejects an HTML fallback masquerading as a successful download', async () => {
    vi.mocked(fetch).mockResolvedValue(new Response('<html>SPA fallback</html>'))
    const wrapper = render(); await flushPromises()
    expect(wrapper.text()).toContain('本站尚未发布配置助手')
    await wrapper.get('select').setValue(1)
    await wrapper.get('#wb-model').setValue('test-model')
    await wrapper.get('[type=checkbox]').setValue(true)
    await wrapper.get('form').trigger('submit'); expect(api.post).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('clears expired codes and keeps server failures actionable', async () => {
    vi.useFakeTimers()
    const wrapper = render(); await flushPromises()
    await wrapper.get('select').setValue(1)
    await wrapper.get('#wb-model').setValue('test-model')
    await wrapper.get('[type=checkbox]').setValue(true)
    await wrapper.get('form').trigger('submit'); await flushPromises()
    await vi.advanceTimersByTimeAsync(301000)
    expect(wrapper.text()).not.toContain('a'.repeat(64))
    expect(wrapper.text()).toContain('配对码已过期')
    api.post.mockRejectedValue(new Error('503'))
    await wrapper.get('form').trigger('submit'); await flushPromises()
    expect(wrapper.get('[role=alert]').text()).toContain('配对失败')
    wrapper.unmount()
  })
})
