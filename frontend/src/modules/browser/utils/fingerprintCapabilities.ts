export type FingerprintCapabilityMode = 'configurable' | 'seed' | 'platform_limited' | 'removed'

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
    coverage: '启用源内核大部分指纹伪装，并保持同实例稳定',
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
    name: 'CPU 与内存',
    mode: 'configurable',
    coverage: 'CPU 核心数可配置；内存由源内核 seed 模型生成',
    args: ['--fingerprint-hardware-concurrency', '--fingerprint'],
  },
  {
    id: 'canvas-audio',
    name: 'Canvas / Audio',
    mode: 'seed',
    coverage: '由 --fingerprint 种子驱动，适合同实例稳定、不同实例差异化',
    args: ['--fingerprint', '--disable-spoofing=canvas,audio'],
  },
  {
    id: 'fonts-clientrects',
    name: '字体 / ClientRects',
    mode: 'seed',
    coverage: '由 --fingerprint 种子驱动，可用 disable-spoofing 排除单项伪装',
    args: ['--fingerprint', '--disable-spoofing=font,clientrects'],
  },
  {
    id: 'webgl',
    name: 'WebGL 图像与 GPU',
    mode: 'platform_limited',
    coverage: 'WebGL 图像由 seed 驱动；GPU 元数据在 Linux 支持更完整',
    args: ['--fingerprint', '--disable-spoofing=gpu'],
  },
  {
    id: 'webrtc',
    name: 'WebRTC 防泄漏',
    mode: 'configurable',
    coverage: '通过禁用非代理 UDP 降低本机 IP 泄漏风险',
    args: ['--disable-non-proxied-udp'],
  },
  {
    id: 'gpu-legacy',
    name: '旧 GPU 指定参数',
    mode: 'removed',
    coverage: 'Chrome 144 已移除独立 GPU vendor / renderer 参数',
    args: ['--fingerprint-gpu-vendor', '--fingerprint-gpu-renderer', '--disable-gpu-fingerprint'],
  },
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
    name: '日本 macOS Chrome 轻办公',
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
  return '已移除'
}
