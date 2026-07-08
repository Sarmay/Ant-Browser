import { useEffect, useState } from 'react'
import { ChevronDown, ChevronUp, RefreshCw } from 'lucide-react'
import { ConfirmModal, FormItem, Input, Select, Switch, Textarea } from '../../../shared/components'
import {
  type FingerprintConfig,
  FINGERPRINT_PRESETS,
  PRESET_RESOLUTIONS,
  buildFingerprintConfigFromPersona,
  deserialize,
  getSystemTimezone,
  randomFingerprintSeed,
  serialize,
  validateFingerprintArgs,
} from '../utils/fingerprintSerializer'
import { FINGERPRINT_CAPABILITIES, FINGERPRINT_PERSONAS, capabilityModeLabel } from '../utils/fingerprintCapabilities'

interface FingerprintPanelProps {
  value: string[]
  onChange: (args: string[]) => void
}

const BRAND_OPTIONS = [
  { value: '', label: '不设置' },
  { value: 'Chrome', label: 'Chrome' },
  { value: 'Edge', label: 'Edge' },
  { value: 'Opera', label: 'Opera' },
  { value: 'Vivaldi', label: 'Vivaldi' },
]

const PLATFORM_OPTIONS = [
  { value: '', label: '不设置' },
  { value: 'windows', label: 'Windows' },
  { value: 'macos', label: 'macOS' },
  { value: 'linux', label: 'Linux' },
]

const LANG_OPTIONS = [
  { value: '', label: '不设置' },
  { value: 'zh-CN', label: '中文 (zh-CN)' },
  { value: 'zh-HK', label: '繁體中文香港 (zh-HK)' },
  { value: 'zh-TW', label: '繁體中文台灣 (zh-TW)' },
  { value: 'en-US', label: 'English US (en-US)' },
  { value: 'en-GB', label: 'English UK (en-GB)' },
  { value: 'en-CA', label: 'English Canada (en-CA)' },
  { value: 'en-AU', label: 'English Australia (en-AU)' },
  { value: 'en-SG', label: 'English Singapore (en-SG)' },
  { value: 'en-IN', label: 'English India (en-IN)' },
  { value: 'ja-JP', label: '日本語 (ja-JP)' },
  { value: 'ko-KR', label: '한국어 (ko-KR)' },
  { value: 'fr-FR', label: 'Français (fr-FR)' },
  { value: 'de-DE', label: 'Deutsch (de-DE)' },
  { value: 'nl-NL', label: 'Nederlands (nl-NL)' },
  { value: 'ru-RU', label: 'Русский (ru-RU)' },
  { value: 'pt-BR', label: 'Português Brasil (pt-BR)' },
]

const TIMEZONE_OPTIONS = [
  { value: '', label: '不设置' },
  { value: 'system', label: '跟随系统时区' },
  { value: 'Asia/Shanghai', label: 'Asia/Shanghai (UTC+8)' },
  { value: 'Asia/Tokyo', label: 'Asia/Tokyo (UTC+9)' },
  { value: 'Asia/Seoul', label: 'Asia/Seoul (UTC+9)' },
  { value: 'Asia/Singapore', label: 'Asia/Singapore (UTC+8)' },
  { value: 'Asia/Hong_Kong', label: 'Asia/Hong_Kong (UTC+8)' },
  { value: 'Asia/Taipei', label: 'Asia/Taipei (UTC+8)' },
  { value: 'Asia/Dubai', label: 'Asia/Dubai (UTC+4)' },
  { value: 'Asia/Kolkata', label: 'Asia/Kolkata (UTC+5:30)' },
  { value: 'America/New_York', label: 'America/New_York (UTC-5)' },
  { value: 'America/Los_Angeles', label: 'America/Los_Angeles (UTC-8)' },
  { value: 'America/Chicago', label: 'America/Chicago (UTC-6)' },
  { value: 'America/Denver', label: 'America/Denver (UTC-7)' },
  { value: 'America/Toronto', label: 'America/Toronto (UTC-5)' },
  { value: 'America/Vancouver', label: 'America/Vancouver (UTC-8)' },
  { value: 'America/Phoenix', label: 'America/Phoenix (UTC-7)' },
  { value: 'America/Sao_Paulo', label: 'America/Sao_Paulo (UTC-3)' },
  { value: 'Europe/London', label: 'Europe/London (UTC+0)' },
  { value: 'Europe/Paris', label: 'Europe/Paris (UTC+1)' },
  { value: 'Europe/Berlin', label: 'Europe/Berlin (UTC+1)' },
  { value: 'Europe/Moscow', label: 'Europe/Moscow (UTC+3)' },
  { value: 'Australia/Sydney', label: 'Australia/Sydney (UTC+10)' },
  { value: 'Australia/Melbourne', label: 'Australia/Melbourne (UTC+10)' },
  { value: 'Australia/Perth', label: 'Australia/Perth (UTC+8)' },
  { value: 'Pacific/Auckland', label: 'Pacific/Auckland (UTC+12)' },
]

const RESOLUTION_OPTIONS = [
  { value: '', label: '不设置' },
  ...PRESET_RESOLUTIONS.map(r => ({ value: r, label: r })),
  { value: 'custom', label: '自定义...' },
]

const HARDWARE_CONCURRENCY_OPTIONS = [
  { value: '', label: '不设置' },
  { value: '2', label: '2 核' },
  { value: '4', label: '4 核' },
  { value: '6', label: '6 核' },
  { value: '8', label: '8 核' },
  { value: '10', label: '10 核' },
  { value: '12', label: '12 核' },
  { value: '16', label: '16 核' },
]

const WEBRTC_OPTIONS = [
  { value: '', label: '不设置' },
  { value: 'disable_non_proxied_udp', label: '禁用非代理 UDP' },
  { value: 'default_public_interface_only', label: '仅公网接口' },
  { value: 'default_public_and_private_interfaces', label: '公网+私网接口' },
]

const SPOOFING_OPTIONS = [
  { value: 'font', label: '字体' },
  { value: 'audio', label: '音频' },
  { value: 'canvas', label: 'Canvas' },
  { value: 'clientrects', label: 'ClientRects' },
  { value: 'gpu', label: 'GPU' },
]

const PRESET_OPTIONS = [
  { value: '', label: '选择预设...' },
  ...FINGERPRINT_PRESETS.map(p => ({ value: p.id, label: p.name })),
]

const PERSONA_OPTIONS = [
  { value: '', label: '选择高级画像...' },
  ...FINGERPRINT_PERSONAS.map(item => ({ value: item.id, label: item.name })),
]

export function FingerprintPanel({ value, onChange }: FingerprintPanelProps) {
  const [config, setConfig] = useState<FingerprintConfig>(() => deserialize(value))
  const [advancedOpen, setAdvancedOpen] = useState(false)
  const [confirmSeedOpen, setConfirmSeedOpen] = useState(false)

  useEffect(() => {
    setConfig(deserialize(value))
  }, [value.join('\n')])

  const update = (patch: Partial<FingerprintConfig>) => {
    const next = { ...config, ...patch }
    setConfig(next)
    onChange(serialize(next))
  }

  const handlePresetChange = (presetId: string) => {
    if (!presetId) return
    const preset = FINGERPRINT_PRESETS.find(p => p.id === presetId)
    if (!preset) return
    const next: FingerprintConfig = {
      ...preset.config,
      seed: randomFingerprintSeed(),
      unknownArgs: config.unknownArgs,
    }
    setConfig(next)
    onChange(serialize(next))
  }

  const handleAdvancedChange = (text: string) => {
    const args = text.split('\n').map(s => s.trim()).filter(Boolean)
    const parsed = deserialize(args)
    setConfig(parsed)
    onChange(serialize(parsed))
  }

  const handlePersonaChange = (personaId: string) => {
    if (!personaId) return
    const persona = FINGERPRINT_PERSONAS.find(item => item.id === personaId)
    if (!persona) return
    const next: FingerprintConfig = {
      ...buildFingerprintConfigFromPersona(persona),
      unknownArgs: config.unknownArgs,
    }
    setConfig(next)
    onChange(serialize(next))
  }

  const toggleDisableSpoofing = (key: string, checked: boolean) => {
    const current = config.disableSpoofing ?? []
    const next = checked ? [...current, key] : current.filter(item => item !== key)
    update({ disableSpoofing: next.length ? next : undefined })
  }

  const advancedText = serialize(config).join('\n')
  const validation = validateFingerprintArgs(value)
  const validationTone = validation.issues.some(issue => issue.level === 'error')
    ? 'border-red-200 bg-red-50 text-red-700'
    : validation.issues.some(issue => issue.level === 'warning')
      ? 'border-amber-200 bg-amber-50 text-amber-700'
      : 'border-emerald-200 bg-emerald-50 text-emerald-700'
  const validationTitle = validation.valid
    ? validation.issues.some(issue => issue.level === 'warning') ? '配置可用，有提示' : '配置有效'
    : '配置需要修正'

  return (
    <div className="space-y-4">
      <div className={`rounded-lg border px-3 py-2 text-sm ${validationTone}`}>
        <div className="font-medium">{validationTitle}</div>
        {validation.issues.length > 0 && (
          <ul className="mt-1 space-y-1">
            {validation.issues.slice(0, 4).map((issue, index) => (
              <li key={`${issue.level}-${index}`} className="text-xs">{issue.message}</li>
            ))}
            {validation.issues.length > 4 && <li className="text-xs">还有 {validation.issues.length - 4} 项，请在高级模式中检查</li>}
          </ul>
        )}
      </div>

      <div className="p-3 rounded-lg bg-[var(--color-bg-hover)] border border-[var(--color-border)] space-y-2">
        <div className="flex items-center justify-between gap-3">
          <span className="text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wide">指纹种子</span>
          <button
            type="button"
            className="inline-flex items-center gap-1 text-xs text-[var(--color-primary)] hover:underline"
            onClick={() => setConfirmSeedOpen(true)}
          >
            <RefreshCw className="w-3 h-3" />
            重新生成
          </button>
        </div>
        <Input value={config.seed ?? ''} onChange={e => update({ seed: e.target.value || undefined })} placeholder="留空则启动时按实例 ID 自动生成" />
      </div>

      <ConfirmModal
        open={confirmSeedOpen}
        onClose={() => setConfirmSeedOpen(false)}
        onConfirm={() => update({ seed: randomFingerprintSeed() })}
        title="重新生成指纹种子"
        content="重新生成后，会影响当前内核支持的随机指纹项；具体生效范围以检测结果为准。确定继续？"
        confirmText="确定重新生成"
      />

      <FormItem label="快速预设">
        <Select value="" onChange={e => handlePresetChange(e.target.value)} options={PRESET_OPTIONS} />
      </FormItem>

      <FormItem label="高级画像">
        <Select value="" onChange={e => handlePersonaChange(e.target.value)} options={PERSONA_OPTIONS} />
      </FormItem>

      <div className="rounded-lg border border-[var(--color-border)] overflow-hidden">
        <div className="grid grid-cols-[110px_88px_minmax(0,1fr)] gap-3 px-3 py-2 text-xs font-medium text-[var(--color-text-muted)] border-b border-[var(--color-border)]">
          <div>能力</div>
          <div>模式</div>
          <div>覆盖</div>
        </div>
        {FINGERPRINT_CAPABILITIES.map(item => (
          <div key={item.id} className="grid grid-cols-[110px_88px_minmax(0,1fr)] gap-3 px-3 py-2 text-xs border-b last:border-b-0 border-[var(--color-border)]">
            <div className="font-medium text-[var(--color-text-primary)]">{item.name}</div>
            <div className="text-[var(--color-text-secondary)]">{capabilityModeLabel(item.mode)}</div>
            <div className="text-[var(--color-text-muted)]">{item.coverage}</div>
          </div>
        ))}
      </div>

      <div>
        <p className="text-xs font-medium text-[var(--color-text-muted)] mb-2 uppercase tracking-wide">身份与定位</p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormItem label="浏览器品牌">
            <Select value={config.brand ?? ''} onChange={e => update({ brand: e.target.value || undefined })} options={BRAND_OPTIONS} />
          </FormItem>
          <FormItem label="品牌版本">
            <Input value={config.brandVersion ?? ''} onChange={e => update({ brandVersion: e.target.value || undefined })} placeholder="默认跟随内核" />
          </FormItem>
          <FormItem label="平台">
            <Select value={config.platform ?? ''} onChange={e => update({ platform: e.target.value || undefined })} options={PLATFORM_OPTIONS} />
          </FormItem>
          <FormItem label="系统版本">
            <Input value={config.platformVersion ?? ''} onChange={e => update({ platformVersion: e.target.value || undefined })} placeholder="如 15.2.0" />
          </FormItem>
          <FormItem label="语言">
            <Select value={config.lang ?? ''} onChange={e => update({ lang: e.target.value || undefined, acceptLang: undefined })} options={LANG_OPTIONS} />
          </FormItem>
          <FormItem label="时区">
            <Select
              value={config.timezone ?? ''}
              onChange={e => update({ timezone: e.target.value || undefined })}
              options={TIMEZONE_OPTIONS.map(opt => opt.value === 'system' ? { ...opt, label: `跟随系统时区 (当前: ${getSystemTimezone()})` } : opt)}
            />
          </FormItem>
        </div>
      </div>

      <div>
        <p className="text-xs font-medium text-[var(--color-text-muted)] mb-2 uppercase tracking-wide">窗口与硬件</p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormItem label="窗口大小">
            <Select value={config.resolution ?? ''} onChange={e => update({ resolution: e.target.value || undefined })} options={RESOLUTION_OPTIONS} />
          </FormItem>
          {config.resolution === 'custom' && (
            <FormItem label="自定义分辨率">
              <Input value={config.customResolution ?? ''} onChange={e => update({ customResolution: e.target.value || undefined })} placeholder="1920,1080" />
            </FormItem>
          )}
          <FormItem label="CPU 核心数">
            <Select value={config.hardwareConcurrency ?? ''} onChange={e => update({ hardwareConcurrency: e.target.value || undefined })} options={HARDWARE_CONCURRENCY_OPTIONS} />
          </FormItem>
        </div>
      </div>

      <div>
        <p className="text-xs font-medium text-[var(--color-text-muted)] mb-2 uppercase tracking-wide">网络</p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormItem label="WebRTC 策略">
            <Select value={config.webrtcPolicy ?? ''} onChange={e => update({ webrtcPolicy: e.target.value || undefined })} options={WEBRTC_OPTIONS} />
          </FormItem>
        </div>
      </div>

      <div>
        <p className="text-xs font-medium text-[var(--color-text-muted)] mb-2 uppercase tracking-wide">内核伪装</p>
        <div className="rounded-lg border border-[var(--color-border)] divide-y divide-[var(--color-border)]">
          <div className="px-3 py-2 text-sm text-[var(--color-text-secondary)]">
            种子会启用音频、字体、Canvas、ClientRects、GPU 等内核级伪装；这里仅用于排除某项伪装。
          </div>
          <div className="grid grid-cols-1 md:grid-cols-5 gap-0">
            {SPOOFING_OPTIONS.map(option => (
              <label key={option.value} className="flex items-center justify-between gap-3 px-3 py-2 text-sm text-[var(--color-text-primary)] border-t md:border-t-0 md:border-r last:border-r-0 border-[var(--color-border)]">
                <span>{option.label}</span>
                <Switch
                  checked={(config.disableSpoofing ?? []).includes(option.value)}
                  onChange={checked => toggleDisableSpoofing(option.value, checked)}
                />
              </label>
            ))}
          </div>
        </div>
      </div>

      <div className="border border-[var(--color-border)] rounded-lg overflow-hidden">
        <button
          type="button"
          className="w-full flex items-center justify-between px-4 py-2.5 text-sm text-[var(--color-text-muted)] hover:bg-[var(--color-bg-hover)] transition-colors"
          onClick={() => setAdvancedOpen(v => !v)}
        >
          <span>高级模式（原始参数）</span>
          {advancedOpen ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
        </button>
        {advancedOpen && (
          <div className="px-4 pb-4 pt-2 border-t border-[var(--color-border)]">
            <p className="text-xs text-[var(--color-text-muted)] mb-2">仅保留当前内核已支持或未知的原始参数。</p>
            <Textarea
              value={advancedText}
              onChange={e => handleAdvancedChange(e.target.value)}
              rows={6}
              placeholder="--fingerprint-brand=Chrome"
            />
          </div>
        )}
      </div>
    </div>
  )
}
