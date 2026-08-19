export type FingerprintCapabilityMode = 'configurable' | 'seed' | 'platform_limited' | 'unconfirmed' | 'removed'

export interface FingerprintCapabilityItem {
  id: string
  name: string
  mode: FingerprintCapabilityMode
  coverage: string
  args: string[]
}

export interface FingerprintPersona {
  id: string
  name: string
  region: string
  platform: 'windows' | 'macos' | 'linux'
  brand: 'Chrome' | 'Edge' | 'Opera' | 'Vivaldi'
  lang: string
  timezone: string
  resolution: string
  hardwareConcurrency: string
  brandVersion?: string
  platformVersion?: string
}

export const FINGERPRINT_CAPABILITIES: FingerprintCapabilityItem[] = [
  {
    id: 'seed',
    name: '核心种子',
    mode: 'configurable',
    coverage: '启用本地 Chrom-144 已接入的指纹伪装，并保持同实例稳定',
    args: ['--fingerprint'],
  },
  {
    id: 'identity',
    name: '浏览器身份',
    mode: 'configurable',
    coverage: '控制 UA、navigator.platform、UA-CH 的品牌和平台画像',
    args: ['--fingerprint-brand', '--fingerprint-brand-version', '--fingerprint-platform', '--fingerprint-platform-version'],
  },
  {
    id: 'locale',
    name: '语言与时区',
    mode: 'configurable',
    coverage: '控制 navigator.language、navigator.languages、Accept-Language 和 Intl 时区',
    args: ['--lang', '--accept-lang', '--timezone'],
  },
  {
    id: 'hardware',
    name: 'CPU',
    mode: 'configurable',
    coverage: 'CPU 核心数已实测生效；设备内存、色深和触控点本地 Chrom-144 实测无效，不再作为运行参数传递',
    args: ['--fingerprint-hardware-concurrency'],
  },
  {
    id: 'canvas-audio',
    name: 'Canvas / Audio',
    mode: 'configurable',
    coverage: 'Canvas ImageData 噪声已实测生效；Audio 独立参数本地 Chrom-144 实测无效，不再作为运行参数传递',
    args: ['--fingerprinting-canvas-image-data-noise', '--disable-spoofing=canvas,audio'],
  },
  {
    id: 'fonts-clientrects',
    name: '字体 / ClientRects',
    mode: 'configurable',
    coverage: 'ClientRects 噪声已实测生效；字体列表独立参数本地 Chrom-144 实测无效，不再作为运行参数传递',
    args: ['--fingerprinting-client-rects-noise', '--disable-spoofing=font,clientrects'],
  },
  {
    id: 'webgl',
    name: 'WebGL 图像与 GPU',
    mode: 'unconfirmed',
    coverage: 'WebGL Vendor / Renderer 独立参数本地 Chrom-144 实测无效；GPU 伪装按种子和 disable-spoofing 控制',
    args: ['--disable-spoofing=gpu'],
  },
  {
    id: 'webrtc',
    name: 'WebRTC 防泄漏',
    mode: 'configurable',
    coverage: 'WebRTC 防泄漏开关可配置；DNT 和媒体设备数量本地 Chrom-144 实测无效，不再作为运行参数传递',
    args: ['--disable-non-proxied-udp', '--webrtc-ip-handling-policy'],
  },
  {
    id: 'gpu-legacy',
    name: '旧 GPU 指定参数',
    mode: 'removed',
    coverage: '当前适配矩阵不再把旧 GPU vendor / renderer 参数作为独立运行参数',
    args: ['--fingerprint-gpu-vendor', '--fingerprint-gpu-renderer', '--disable-gpu-fingerprint'],
  },
]

export const PLATFORM_VERSION_OPTIONS = [
  { value: '10.0.0', label: 'Windows 10 / 10.0.0' },
  { value: '15.0.0', label: 'Windows 11 / 15.0.0' },
  { value: '12.7.6', label: 'macOS 12 Monterey / 12.7.6' },
  { value: '13.7.4', label: 'macOS 13 Ventura / 13.7.4' },
  { value: '14.7.1', label: 'macOS 14 Sonoma / 14.7.1' },
  { value: '15.2.0', label: 'macOS 15 Sequoia / 15.2.0' },
  { value: '15.5.0', label: 'macOS 15 Sequoia / 15.5.0' },
  { value: '15.6.1', label: 'macOS 15 Sequoia / 15.6.1' },
  { value: '26.0.0', label: 'macOS 26 Tahoe / 26.0.0' },
  { value: '26.4.1', label: 'macOS 26 Tahoe / 26.4.1' },
]

export const FINGERPRINT_PERSONAS: FingerprintPersona[] = [
  {
    id: 'us-win-chrome-office',
    name: '美国 Windows Chrome 办公',
    region: 'US',
    platform: 'windows',
    brand: 'Chrome',
    lang: 'en-US',
    timezone: 'America/New_York',
    resolution: '1920,1080',
    hardwareConcurrency: '8',
    platformVersion: '10.0.0',
  },
  {
    id: 'jp-mac-chrome-light',
    name: '日本 macOS 15 Sequoia Chrome 轻办公',
    region: 'JP',
    platform: 'macos',
    brand: 'Chrome',
    lang: 'ja-JP',
    timezone: 'Asia/Tokyo',
    resolution: '1440,900',
    hardwareConcurrency: '8',
    platformVersion: '15.2.0',
  },
  {
    id: 'cn-mac-sonoma-design',
    name: '中国 macOS 14 Sonoma Chrome 设计',
    region: 'CN',
    platform: 'macos',
    brand: 'Chrome',
    lang: 'zh-CN',
    timezone: 'Asia/Shanghai',
    resolution: '2560,1440',
    hardwareConcurrency: '10',
    platformVersion: '14.7.1',
  },
  {
    id: 'us-mac-ventura-edu',
    name: '美国 macOS 13 Ventura Chrome 教育',
    region: 'US',
    platform: 'macos',
    brand: 'Chrome',
    lang: 'en-US',
    timezone: 'America/New_York',
    resolution: '1440,900',
    hardwareConcurrency: '8',
    platformVersion: '13.7.4',
  },
  {
    id: 'hk-mac-sequoia-business',
    name: '香港 macOS 15 Sequoia Chrome 商务',
    region: 'HK',
    platform: 'macos',
    brand: 'Chrome',
    lang: 'zh-HK',
    timezone: 'Asia/Hong_Kong',
    resolution: '1920,1080',
    hardwareConcurrency: '10',
    platformVersion: '15.6.1',
  },
  {
    id: 'us-mac-tahoe-office',
    name: '美国 macOS 26 Tahoe Chrome 办公',
    region: 'US',
    platform: 'macos',
    brand: 'Chrome',
    lang: 'en-US',
    timezone: 'America/Los_Angeles',
    resolution: '1920,1080',
    hardwareConcurrency: '12',
    platformVersion: '26.4.1',
  },
  {
    id: 'gb-mac-sonoma-office',
    name: '英国 macOS 14 Sonoma Chrome 办公',
    region: 'GB',
    platform: 'macos',
    brand: 'Chrome',
    lang: 'en-GB',
    timezone: 'Europe/London',
    resolution: '1920,1080',
    hardwareConcurrency: '8',
    platformVersion: '14.7.1',
  },
  {
    id: 'sg-mac-tahoe-business',
    name: '新加坡 macOS 26 Tahoe Chrome 商务',
    region: 'SG',
    platform: 'macos',
    brand: 'Chrome',
    lang: 'en-SG',
    timezone: 'Asia/Singapore',
    resolution: '2560,1440',
    hardwareConcurrency: '10',
    platformVersion: '26.0.0',
  },
  {
    id: 'au-mac-monterey-light',
    name: '澳洲 macOS 12 Monterey Chrome 轻办公',
    region: 'AU',
    platform: 'macos',
    brand: 'Chrome',
    lang: 'en-AU',
    timezone: 'Australia/Sydney',
    resolution: '1440,900',
    hardwareConcurrency: '8',
    platformVersion: '12.7.6',
  },
  {
    id: 'uk-win-edge-enterprise',
    name: '英国 Windows Edge 企业',
    region: 'GB',
    platform: 'windows',
    brand: 'Edge',
    lang: 'en-GB',
    timezone: 'Europe/London',
    resolution: '1920,1080',
    hardwareConcurrency: '8',
    platformVersion: '10.0.0',
  },
  {
    id: 'sg-win-chrome-business',
    name: '新加坡 Windows Chrome 商务',
    region: 'SG',
    platform: 'windows',
    brand: 'Chrome',
    lang: 'en-SG',
    timezone: 'Asia/Singapore',
    resolution: '1920,1080',
    hardwareConcurrency: '8',
    platformVersion: '10.0.0',
  },
  {
    id: 'de-linux-chrome-dev',
    name: '德国 Linux Chrome 开发',
    region: 'DE',
    platform: 'linux',
    brand: 'Chrome',
    lang: 'de-DE',
    timezone: 'Europe/Berlin',
    resolution: '1920,1080',
    hardwareConcurrency: '12',
  },
]

export function capabilityModeLabel(mode: FingerprintCapabilityMode): string {
  if (mode === 'configurable') return '可配置'
  if (mode === 'seed') return '种子驱动'
  if (mode === 'platform_limited') return '平台限制'
  if (mode === 'unconfirmed') return '实测无效'
  return '已移除'
}
