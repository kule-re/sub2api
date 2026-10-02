import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import WugouLanding from '../WugouLanding.vue'

const localeState = vi.hoisted(() => ({ value: 'zh' as 'zh' | 'en' }))
vi.mock('@/i18n', () => ({
  availableLocales: [{ code: 'zh', name: '中文' }, { code: 'en', name: 'English' }],
  getLocale: () => localeState.value,
  setLocale: vi.fn(async (code: 'zh' | 'en') => { localeState.value = code }),
}))

let wrapper: ReturnType<typeof mount>
function render() {
  wrapper = mount(WugouLanding, {
    props: {
      siteName: 'Custom Brand', siteLogo: '/custom-logo.png', docUrl: '/docs',
      isAuthenticated: false, dashboardPath: '/dashboard',
      showModelPlazaEntry: true, registrationEnabled: true, isDark: false,
    },
    attachTo: document.body,
    global: { stubs: { RouterLink: RouterLinkStub, AnnouncementBell: true } },
  })
}

beforeEach(() => {
  localeState.value = 'zh'
  vi.spyOn(window, 'matchMedia').mockReturnValue({ matches: true } as MediaQueryList)
})
afterEach(() => {
  wrapper?.unmount()
  vi.restoreAllMocks()
  document.body.style.overflow = ''
})

async function selectLanguage(language: string, accessibleLabel: string) {
  await wrapper.get(`[aria-label="${accessibleLabel}"]`).trigger('click')
  const choice = wrapper.findAll('[role="menuitemradio"]').find(item => item.text().includes(language))
  expect(choice).toBeDefined()
  await choice!.trigger('click')
  await flushPromises()
}

describe('Wugou landing language selection', () => {
  it('changes all page sections, navigation, and accessible labels without changing API examples or branding', async () => {
    render()
    const codeExample = wrapper.get('#protocol-panel').html()
    expect(wrapper.get('h1').text()).toContain('统一 API 网关')
    expect(wrapper.text()).toContain('通过Custom Brand 接入常用 AI 应用与开发工具')

    await selectLanguage('English', '更改语言')
    expect(wrapper.get('h1').text()).toContain('One API gateway for')
    expect(wrapper.text()).toContain('Core features')
    expect(wrapper.text()).toContain('Get started in three steps')
    expect(wrapper.text()).toContain('your AI integration?')
    expect(wrapper.text()).toContain('All rights reserved.')
    expect(wrapper.text()).toContain('Create an API key on Custom Brand')
    expect(wrapper.text()).not.toMatch(/[一-龥]/)
    expect(wrapper.get('#protocol-panel').html()).toBe(codeExample)
    expect(wrapper.get('img').attributes('alt')).toBe('Custom Brand')
    expect(wrapper.find('[role="menu"]').exists()).toBe(false)
    expect(wrapper.get('[aria-label="Change language"]').exists()).toBe(true)

    await wrapper.get('[aria-label="Toggle navigation menu"]').trigger('click')
    const mobile = wrapper.get('#landing-mobile-menu')
    expect(mobile.attributes('aria-label')).toBe('Site navigation')
    expect(mobile.text()).toContain('Dashboard')
    expect(mobile.text()).toContain('Image generation')
    expect(mobile.text()).toContain('Sign in')

    await selectLanguage('中文', 'Change language')
    expect(wrapper.get('h1').text()).toContain('统一 API 网关')
    expect(wrapper.get('#landing-mobile-menu').text()).toContain('控制台')
    expect(wrapper.text()).toContain('三步快速上手')
    expect(wrapper.get('#protocol-panel').html()).toBe(codeExample)
  })

  it('renders the saved English locale on first load', () => {
    localeState.value = 'en'
    render()
    expect(wrapper.get('h1').text()).toContain('One API gateway for')
    expect(wrapper.get('[aria-label="Change language"]').exists()).toBe(true)
    expect(wrapper.text()).not.toMatch(/[一-龥]/)
  })
})
