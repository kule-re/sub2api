import { beforeEach, describe, expect, it } from 'vitest'
import { resolveSiteLogo, resolveSiteName, updateFavicon } from '@/utils/branding'

describe('site branding defaults', () => {
  it.each([undefined, null, '', '   ', 'Sub2API', ' Sub2API ', '悟狗ai', ' 悟狗ai '])('replaces legacy site name %s', (name) => {
    expect(resolveSiteName(name)).toBe('哈呀哈基米ai')
  })

  it.each([
    undefined, null, '', '/logo.svg', '/logo.svg?v=1', './logo.svg#icon',
    '/wugou-logo.png', 'wugou-logo.png', './wugou-logo.png',
    '/wugou-logo.png?v=1', './wugou-logo.png#icon', ' /wugou-logo.png?v=1#icon ',
  ])('replaces legacy logo %s', (logo) => {
    expect(resolveSiteLogo(logo)).toBe('/hakimi-logo.png')
  })

  it('preserves administrator-defined branding', () => {
    expect(resolveSiteName('My gateway')).toBe('My gateway')
    expect(resolveSiteName('悟狗ai Custom')).toBe('悟狗ai Custom')
    expect(resolveSiteLogo('https://example.com/logo.svg?v=2')).toBe('https://example.com/logo.svg?v=2')
    expect(resolveSiteLogo('https://example.com/wugou-logo.png?v=2')).toBe('https://example.com/wugou-logo.png?v=2')
    expect(resolveSiteLogo('/custom/wugou-logo.png')).toBe('/custom/wugou-logo.png')
    expect(resolveSiteLogo('data:image/png;base64,abc')).toBe('data:image/png;base64,abc')
  })
})

describe('updateFavicon', () => {
  beforeEach(() => {
    document.head.innerHTML = '<link rel="icon" href="/logo.svg">'
  })

  it('replaces the default favicon with the configured logo', () => {
    updateFavicon('https://example.com/custom-logo.png')

    const link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
    expect(link?.href).toBe('https://example.com/custom-logo.png')
    expect(link?.type).toBe('image/png')
  })

  it.each([
    ['/hakimi-logo.png?v=1', 'image/png'],
    ['/custom.svg?v=2#icon', 'image/svg+xml'],
    ['/custom.jpeg', 'image/jpeg'],
    ['data:image/png;base64,abc', 'image/png'],
  ])('sets the correct MIME for %s', (logo, mime) => {
    updateFavicon(logo)
    expect(document.querySelector<HTMLLinkElement>('link[rel="icon"]')?.type).toBe(mime)
  })

  it('lets the browser determine MIME for extensionless custom logo URLs', () => {
    updateFavicon('https://example.com/logo')
    expect(document.querySelector<HTMLLinkElement>('link[rel="icon"]')?.hasAttribute('type')).toBe(false)
  })

  it('ignores unsafe logo URLs', () => {
    updateFavicon('javascript:alert(1)')

    const link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
    expect(link?.getAttribute('href')).toBe('/logo.svg')
  })
})
