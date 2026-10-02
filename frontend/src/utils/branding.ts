import { sanitizeUrl } from '@/utils/url'

export const DEFAULT_SITE_NAME = '哈呀哈基米ai'
export const DEFAULT_SITE_LOGO = '/hakimi-logo.png'

export function resolveSiteName(siteName: unknown): string {
  const value = typeof siteName === 'string' ? siteName.trim() : ''
  return !value || ['Sub2API', '悟狗ai'].includes(value) ? DEFAULT_SITE_NAME : value
}

export function resolveSiteLogo(siteLogo: unknown): string {
  const value = typeof siteLogo === 'string' ? siteLogo.trim() : ''
  const path = value.split(/[?#]/, 1)[0]
  return !value || ['/logo.svg', 'logo.svg', './logo.svg', '/wugou-logo.png', 'wugou-logo.png', './wugou-logo.png'].includes(path)
    ? DEFAULT_SITE_LOGO
    : value
}

export function updateFavicon(logoUrl: string): void {
  const sanitizedLogoUrl = sanitizeUrl(logoUrl, {
    allowRelative: true,
    allowDataUrl: true,
  })
  if (!sanitizedLogoUrl) {
    return
  }

  let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }

  const dataType = sanitizedLogoUrl.match(/^data:(image\/[a-z0-9.+-]+)/i)?.[1]
  const extension = sanitizedLogoUrl.split(/[?#]/, 1)[0].split('.').pop()?.toLowerCase()
  const imageTypes: Record<string, string> = {
    svg: 'image/svg+xml',
    png: 'image/png',
    jpg: 'image/jpeg',
    jpeg: 'image/jpeg',
    webp: 'image/webp',
    gif: 'image/gif',
    ico: 'image/x-icon',
  }
  const imageType = dataType || (extension ? imageTypes[extension] : undefined)
  if (imageType) {
    link.type = imageType
  } else {
    link.removeAttribute('type')
  }
  link.href = sanitizedLogoUrl
}
