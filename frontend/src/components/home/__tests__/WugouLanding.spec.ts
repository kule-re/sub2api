import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, RouterLinkStub } from '@vue/test-utils'
import WugouLanding from '../WugouLanding.vue'

vi.mock('@/i18n', () => ({
  availableLocales: [{ code: 'zh', name: '中文' }, { code: 'en', name: 'English' }],
  getLocale: () => 'zh',
  setLocale: vi.fn().mockResolvedValue(undefined),
}))

let wrapper: ReturnType<typeof mount>
const props = {
  siteName: '哈呀哈基米ai', siteLogo: '/hakimi-logo.png', docUrl: '',
  isAuthenticated: false, dashboardPath: '/dashboard',
  showModelPlazaEntry: true, registrationEnabled: true, isDark: false,
}
function render(overrides: Partial<typeof props> = {}) {
  wrapper = mount(WugouLanding, {
    props: { ...props, ...overrides }, attachTo: document.body,
    global: { stubs: { RouterLink: RouterLinkStub, AnnouncementBell: true } },
  })
  return wrapper
}
beforeEach(() => {
  vi.useFakeTimers()
  vi.spyOn(window, 'matchMedia').mockReturnValue({ matches: false } as MediaQueryList)
})
afterEach(() => {
  wrapper?.unmount()
  vi.useRealTimers()
  vi.restoreAllMocks()
  document.body.style.overflow = ''
})

describe('Wugou landing interactions', () => {
  it('switches protocol examples by click and keyboard without making model requests', async () => {
    const request = vi.spyOn(globalThis, 'fetch')
    render()
    await wrapper.get('#protocol-tab-claude').trigger('click')
    expect(wrapper.get('#protocol-panel').text()).toContain('/v1/messages')
    await wrapper.get('[role="tablist"]').trigger('keydown', { key: 'ArrowRight' })
    expect(wrapper.get('#protocol-panel').text()).toContain('/v1beta/models/{model}:generateContent')
    expect(wrapper.get('#protocol-tab-gemini').attributes('aria-selected')).toBe('true')
    expect(document.activeElement?.id).toBe('protocol-tab-gemini')
    vi.advanceTimersByTime(9000)
    expect(wrapper.get('#protocol-tab-gemini').attributes('aria-selected')).toBe('true')
    expect(request).not.toHaveBeenCalled()
  })

  it('keeps registration, admin navigation and private model access consistent with settings', () => {
    render({ isAuthenticated: true, dashboardPath: '/admin/dashboard', showModelPlazaEntry: false })
    const routes = wrapper.findAllComponents(RouterLinkStub).map(link => link.props('to'))
    expect(routes).toContain('/admin/dashboard')
    expect(routes).not.toContain('/register')
    expect(routes).not.toContain('/model-plaza')
    expect(routes).toContain('/batch-image')
    expect(routes).toContain('/workbuddy')
  })

  it('uses login when registration is disabled and preserves administrator documentation', () => {
    render({ registrationEnabled: false, docUrl: 'https://docs.example.test/start' })
    expect(wrapper.findAllComponents(RouterLinkStub).map(link => link.props('to'))).not.toContain('/register')
    expect(wrapper.find('a[href="https://docs.example.test/start"]').exists()).toBe(true)
    expect(wrapper.find('a[href*="qygate.com"]').exists()).toBe(false)
  })

  it('closes mobile navigation on Escape and restores scrolling and focus', async () => {
    document.body.style.overflow = 'auto'
    render()
    const button = wrapper.get('[aria-controls="landing-mobile-menu"]')
    await button.trigger('click')
    expect(wrapper.find('#landing-mobile-menu').exists()).toBe(true)
    expect(document.body.style.overflow).toBe('hidden')
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.find('#landing-mobile-menu').exists()).toBe(false)
    expect(document.body.style.overflow).toBe('auto')
    expect(document.activeElement).toBe(button.element)
  })
})
