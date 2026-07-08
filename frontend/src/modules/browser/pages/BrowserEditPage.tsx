import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { FolderOpen, Layers, ShieldCheck } from 'lucide-react'
import { Button, Card, ConfirmModal, FormItem, Input, Modal, Select, Textarea, toast } from '../../../shared/components'
import type { BrowserCore, BrowserFingerprintCheckResult, BrowserProfileInput, BrowserProxy, BrowserGroup, ProxyLocationResolveResult } from '../types'
import { browserProxyResolveLocation, checkBrowserProfileFingerprint, createBrowserProfile, fetchAllTags, fetchBrowserCores, fetchBrowserProfiles, fetchBrowserProxies, fetchBrowserSettings, fetchGroups, openBrowserFingerprintCheck, openUserDataDir, updateBrowserProfile, validateProxyConfig } from '../api'
import { FingerprintPanel } from '../components/FingerprintPanel'
import { applyLocaleToFingerprintArgs, validateFingerprintArgs } from '../utils/fingerprintSerializer'
import { TagInput } from '../components/TagInput'
import { GroupSelector } from '../components/GroupSelector'
import { ProxyPickerModal } from '../components/ProxyPickerModal'

const fallbackLowLaunchArgs = ['--disable-sync', '--no-first-run']
const directProxyID = '__direct__'
type ProxySourceMode = 'pool' | 'local'
type BrowserProfileEditForm = BrowserProfileInput & { lastLaunchArgs?: string[] }

function normalizeLaunchArgs(args: string[]): string[] {
  return (args || []).map(item => item.trim()).filter(Boolean)
}

function resolveDefaultLaunchArgs(args: string[]): string[] {
  const normalized = normalizeLaunchArgs(args)
  return normalized.length > 0 ? normalized : fallbackLowLaunchArgs
}

function joinValues(values?: string[]): string {
  return values?.length ? values.join(', ') : '-'
}

function displayValue(value: unknown): string {
  if (value === undefined || value === null || value === '') return '-'
  if (Array.isArray(value)) return joinValues(value)
  return String(value)
}

function matchStatus(expected: unknown, actual: unknown, mode: 'exact' | 'contains' = 'exact'): 'match' | 'mismatch' | 'unknown' {
  if (expected === undefined || expected === null || expected === '') return 'unknown'
  if (Array.isArray(actual)) {
    const expectedItems = String(expected).split(',').map(item => item.trim()).filter(Boolean)
    const actualItems = actual.map(item => String(item).trim()).filter(Boolean)
    return expectedItems.length > 0 && expectedItems.every((item, index) => actualItems[index] === item) ? 'match' : 'mismatch'
  }
  if (mode === 'contains') return String(actual).includes(String(expected)) ? 'match' : 'mismatch'
  return String(expected) === String(actual) ? 'match' : 'mismatch'
}

function FingerprintCheckRow({ label, expected, actual, mode }: { label: string; expected?: unknown; actual: unknown; mode?: 'exact' | 'contains' }) {
  const status = matchStatus(expected, actual, mode)
  const statusText = status === 'match' ? '一致' : status === 'mismatch' ? '不一致' : '仅展示'
  const statusClass = status === 'match'
    ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
    : status === 'mismatch'
      ? 'bg-amber-50 text-amber-700 border-amber-200'
      : 'bg-slate-50 text-slate-600 border-slate-200'
  return (
    <div className="grid grid-cols-[120px_minmax(0,1fr)_minmax(0,1fr)_70px] gap-3 items-start py-2 border-b border-[var(--color-border)] last:border-b-0 text-sm">
      <div className="text-[var(--color-text-muted)]">{label}</div>
      <div className="font-mono text-xs break-all text-[var(--color-text-secondary)]">{displayValue(expected)}</div>
      <div className="font-mono text-xs break-all text-[var(--color-text-primary)]">{displayValue(actual)}</div>
      <div><span className={`inline-flex px-2 py-0.5 rounded border text-xs ${statusClass}`}>{statusText}</span></div>
    </div>
  )
}

function resolvePoolProxySelection(
  proxyId: string,
  proxyConfig: string,
  proxies: BrowserProxy[],
): { mode: ProxySourceMode; proxyId: string; proxyConfig: string } {
  const normalizedProxyId = proxyId.trim()
  if (normalizedProxyId) {
    const matchedByID = proxies.find((proxy) => proxy.proxyId.trim() === normalizedProxyId)
    if (matchedByID?.proxyId) {
      return { mode: 'pool', proxyId: matchedByID.proxyId, proxyConfig: '' }
    }
  }

  const rawProxyConfig = proxyConfig.trim()
  const normalizedConfig = rawProxyConfig.toLowerCase()
  if (normalizedConfig) {
    const matchedByConfig = proxies.find((proxy) => (proxy.proxyConfig || '').trim().toLowerCase() === normalizedConfig)
    if (matchedByConfig?.proxyId) {
      return { mode: 'pool', proxyId: matchedByConfig.proxyId, proxyConfig: '' }
    }
    return { mode: 'local', proxyId: '', proxyConfig: rawProxyConfig }
  }

  const directProxy = proxies.find((proxy) => proxy.proxyId === directProxyID)
  return { mode: 'pool', proxyId: directProxy?.proxyId || '', proxyConfig: '' }
}

export function BrowserEditPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isCreate = id === 'new'
  const [formData, setFormData] = useState<BrowserProfileEditForm>({
    profileName: '',
    userDataDir: '',
    coreId: '',
    fingerprintArgs: [],
    proxyId: directProxyID,
    proxyConfig: '',
    launchArgs: [],
    lastLaunchArgs: [],
    tags: [],
    keywords: [],
    groupId: '',
  })
  const [cores, setCores] = useState<BrowserCore[]>([])
  const [proxies, setProxies] = useState<BrowserProxy[]>([])
  const [groups, setGroups] = useState<BrowserGroup[]>([])
  const [launchArgsText, setLaunchArgsText] = useState('')
  const [allTags, setAllTags] = useState<string[]>([])
  const [saving, setSaving] = useState(false)
  const [proxyPickerOpen, setProxyPickerOpen] = useState(false)
  const [proxyMode, setProxyMode] = useState<ProxySourceMode>('pool')
  const [isDirty, setIsDirty] = useState(false)
  const [leaveConfirm, setLeaveConfirm] = useState(false)
  const [saveError, setSaveError] = useState('')
  const [locationResolving, setLocationResolving] = useState(false)
  const [locationResult, setLocationResult] = useState<ProxyLocationResolveResult | null>(null)
  const [fingerprintChecking, setFingerprintChecking] = useState(false)
  const [fingerprintPageOpening, setFingerprintPageOpening] = useState(false)
  const [fingerprintCheckResult, setFingerprintCheckResult] = useState<BrowserFingerprintCheckResult | null>(null)
  const [fingerprintCheckOpen, setFingerprintCheckOpen] = useState(false)

  useEffect(() => {
    const loadData = async () => {
      const [coreList, proxyList, tagList, groupList, settings] = await Promise.all([
        fetchBrowserCores(),
        fetchBrowserProxies(),
        fetchAllTags(),
        fetchGroups(),
        fetchBrowserSettings(),
      ])
      const resolvedDefaultLaunchArgs = resolveDefaultLaunchArgs(settings.defaultLaunchArgs || [])
      setCores(coreList)
      setProxies(proxyList)
      setAllTags(tagList)
      setGroups(groupList)

      if (isCreate) {
        const resolved = resolvePoolProxySelection('', '', proxyList)
        setProxyMode('pool')
        setFormData((prev) => ({ ...prev, proxyId: resolved.proxyId || directProxyID, proxyConfig: '' }))
        setLaunchArgsText(resolvedDefaultLaunchArgs.join('\n'))
        return
      }
      const list = await fetchBrowserProfiles()
      const current = list.find(item => item.profileId === id)
      if (!current) return
      const currentLaunchArgs = normalizeLaunchArgs(current.launchArgs)
      const normalizedCoreId = !current.coreId || current.coreId.toLowerCase() === 'default'
        ? ''
        : current.coreId
      const resolvedProxy = resolvePoolProxySelection(current.proxyId || '', current.proxyConfig || '', proxyList)
      setProxyMode(resolvedProxy.mode)
      setFormData({
        profileName: current.profileName,
        userDataDir: current.userDataDir,
        coreId: normalizedCoreId,
        fingerprintArgs: current.fingerprintArgs,
        proxyId: resolvedProxy.proxyId,
        proxyConfig: resolvedProxy.proxyConfig,
        launchArgs: currentLaunchArgs,
        lastLaunchArgs: current.lastLaunchArgs || [],
        tags: current.tags,
        keywords: current.keywords || [],
        groupId: current.groupId || '',
      })
      setLaunchArgsText(currentLaunchArgs.join('\n'))
    }
    loadData()
  }, [id, isCreate])

  const handleChange = (field: keyof BrowserProfileInput, value: string | string[]) => {
    setIsDirty(true)
    setFormData(prev => {
      if (field === 'proxyId') {
        return { ...prev, proxyId: typeof value === 'string' ? value : '' }
      }
      return { ...prev, [field]: value }
    })
  }

  const handleProxyModeChange = (mode: ProxySourceMode) => {
    setIsDirty(true)
    setProxyMode(mode)
    if (mode === 'pool') {
      setFormData((prev) => {
        if (prev.proxyId.trim()) {
          return prev
        }
        const directProxy = proxies.find((proxy) => proxy.proxyId === directProxyID)
        return {
          ...prev,
          proxyId: directProxy?.proxyId || '',
        }
      })
    }
  }

  const handleSave = async () => {
    const resolvedProxyId = proxyMode === 'pool' ? (formData.proxyId || '').trim() : ''
    const resolvedProxyConfig = proxyMode === 'local' ? (formData.proxyConfig || '').trim() : ''
    if (proxyMode === 'local' && !resolvedProxyConfig) {
      setSaveError('请输入本地代理地址')
      return
    }

    const payload: BrowserProfileInput = {
      profileName: formData.profileName,
      userDataDir: formData.userDataDir,
      coreId: formData.coreId,
      fingerprintArgs: formData.fingerprintArgs,
      tags: formData.tags,
      keywords: formData.keywords,
      groupId: formData.groupId,
      proxyId: resolvedProxyId,
      proxyConfig: resolvedProxyConfig,
      launchArgs: normalizeLaunchArgs(launchArgsText.split('\n')),
    }
    const fingerprintValidation = validateFingerprintArgs(payload.fingerprintArgs)
    if (!fingerprintValidation.valid) {
      setSaveError(fingerprintValidation.issues.filter(issue => issue.level === 'error').map(issue => issue.message).join('\n'))
      return
    }
    if (proxyMode === 'pool' && !resolvedProxyId) {
      payload.proxyId = directProxyID
      payload.proxyConfig = ''
    }

    setSaving(true)
    try {
      const validation = await validateProxyConfig(payload.proxyConfig, payload.proxyId)
      if (!validation.supported) {
        setSaveError(validation.errorMsg || '代理配置无效')
        return
      }
      if (isCreate) {
        await createBrowserProfile(payload)
        toast.success('配置已创建')
      } else if (id) {
        await updateBrowserProfile(id, payload)
        toast.success('配置已更新')
      }
      setIsDirty(false)
      navigate('/browser/list')
    } catch (error: any) {
      setSaveError(typeof error === 'string' ? error : error?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleBack = () => {
    if (isDirty) { setLeaveConfirm(true) } else { navigate('/browser/list') }
  }

  const handleApplyProxyLocation = async () => {
    if (proxyMode !== 'pool' || !formData.proxyId || formData.proxyId === directProxyID) {
      toast.error('请选择代理池中的非直连节点')
      return
    }
    setLocationResolving(true)
    setLocationResult(null)
    try {
      const result = await browserProxyResolveLocation(formData.proxyId)
      setLocationResult(result)
      if (!result.ok || !result.lang || !result.timezone) {
        toast.error(result.error || '无法根据代理 IP 匹配定位')
        return
      }
      const nextArgs = applyLocaleToFingerprintArgs(formData.fingerprintArgs, result.lang, result.timezone)
      handleChange('fingerprintArgs', nextArgs)
      toast.success(`已设置 ${result.lang} / ${result.timezone}`)
    } catch (error: unknown) {
      toast.error((error as Error)?.message || '代理定位失败')
    } finally {
      setLocationResolving(false)
    }
  }

  const handleFingerprintCheck = async () => {
    if (isCreate || !id) {
      toast.warning('请先保存实例，再启动后自测')
      return
    }
    setFingerprintChecking(true)
    try {
      const result = await checkBrowserProfileFingerprint(id)
      setFingerprintCheckResult(result)
      setFingerprintCheckOpen(true)
    } catch (error: unknown) {
      toast.error((error as Error)?.message || '指纹自测失败')
    } finally {
      setFingerprintChecking(false)
    }
  }

  const handleOpenFingerprintPage = async () => {
    if (isCreate || !id) {
      toast.warning('请先保存实例，再打开检测页')
      return
    }
    if (isDirty) {
      toast.warning('当前有未保存修改，请先保存后再检测')
      return
    }
    setFingerprintPageOpening(true)
    try {
      const profile = await openBrowserFingerprintCheck(id)
      if (profile?.lastLaunchArgs) {
        setFormData(prev => ({ ...prev, lastLaunchArgs: profile.lastLaunchArgs || [] }))
      }
      toast.success('已在目标浏览器打开指纹检测页')
    } catch (error: unknown) {
      toast.error((error as Error)?.message || '打开指纹检测页失败')
    } finally {
      setFingerprintPageOpening(false)
    }
  }

  const defaultCore = cores.find(c => c.isDefault)
  const selectedPoolProxy = proxies.find((proxy) => proxy.proxyId === formData.proxyId)

  const handleOpenUserDataDir = async () => {
    if (!formData.userDataDir.trim()) {
      toast.error('请先输入用户数据目录')
      return
    }
    try {
      await openUserDataDir(formData.userDataDir)
    } catch (error: unknown) {
      toast.error((error as Error)?.message || '打开目录失败')
    }
  }

  const handleProxyListUpdated = (nextProxies: BrowserProxy[]) => {
    setProxies(nextProxies)
  }

  const handleProxyDeleted = (deletedProxyId: string, nextProxies: BrowserProxy[]) => {
    setProxies(nextProxies)
    if (formData.proxyId !== deletedProxyId) {
      return
    }

    const fallbackProxy = nextProxies.find((proxy) => proxy.proxyId === directProxyID)
    if (fallbackProxy) {
      handleChange('proxyId', fallbackProxy.proxyId)
      return
    }

    handleChange('proxyId', '')
  }

  return (
    <div className="space-y-5 animate-fade-in">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-[var(--color-text-primary)]">{isCreate ? '新建配置' : '编辑配置'}</h1>
          <p className="text-sm text-[var(--color-text-muted)] mt-1">完善指纹与启动参数</p>
        </div>
        <div className="flex gap-2">
          <Button variant="secondary" size="sm" onClick={handleBack}>返回列表</Button>
          <Button size="sm" onClick={handleSave} loading={saving}>保存配置</Button>
        </div>
      </div>

      <Card title="基础信息" subtitle="实例与配置名称">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormItem label="配置名称" required>
            <Input value={formData.profileName} onChange={e => handleChange('profileName', e.target.value)} placeholder="请输入配置名称" />
          </FormItem>
          <FormItem label="用户数据目录（留空自动生成）">
            <div className="flex gap-2">
              <Input
                value={formData.userDataDir}
                onChange={e => handleChange('userDataDir', e.target.value)}
                placeholder="留空自动生成"
                className="flex-1"
              />
              <Button variant="secondary" size="sm" onClick={handleOpenUserDataDir} title="在资源管理器中打开">
                <FolderOpen className="w-4 h-4" />
              </Button>
            </div>
          </FormItem>
          <FormItem label="内核">
            <Select
              value={formData.coreId}
              onChange={e => handleChange('coreId', e.target.value)}
              options={
                cores.length > 0 ? [
                  { value: '', label: defaultCore ? `使用默认 (${defaultCore.coreName})` : '使用默认内核' },
                  ...cores.map(c => ({ value: c.coreId, label: c.coreName })),
                ] : [
                  { value: '', label: '暂无内核，请添加内核' }
                ]
              }
            />
          </FormItem>
          <FormItem label="标签">
            <TagInput
              value={formData.tags}
              onChange={tags => handleChange('tags', tags)}
              suggestions={allTags}
              placeholder="输入标签后按回车，支持从已有标签选择"
            />
          </FormItem>
          <FormItem label="分组">
            <GroupSelector
              groups={groups}
              value={formData.groupId || ''}
              onChange={groupId => handleChange('groupId', groupId)}
              placeholder="未分组"
              className="w-full"
            />
          </FormItem>
        </div>
      </Card>

      <Card title="代理配置" subtitle="支持代理池节点或本地代理地址">
        <div className="grid grid-cols-1 gap-4">
          <FormItem label="代理来源">
            <Select
              value={proxyMode}
              onChange={e => handleProxyModeChange(e.target.value as ProxySourceMode)}
              options={[
                { value: 'pool', label: '代理池' },
                { value: 'local', label: '本地代理' },
              ]}
            />
          </FormItem>
          {proxyMode === 'pool' ? (
            <FormItem label="代理地址选择">
              <div className="flex gap-2">
                <Select
                  value={formData.proxyId}
                  onChange={e => { handleChange('proxyId', e.target.value); setLocationResult(null) }}
                  options={
                    proxies.length > 0
                      ? proxies.map(p => ({ value: p.proxyId, label: p.proxyName || p.proxyId }))
                      : [{ value: '', label: '暂无代理，请先到代理池创建' }]
                  }
                  className="flex-1"
                />
                <Button variant="secondary" size="sm" onClick={() => setProxyPickerOpen(true)} title="按分组选择代理">
                  <Layers className="w-4 h-4" />
                </Button>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={handleApplyProxyLocation}
                  loading={locationResolving}
                  disabled={!formData.proxyId || formData.proxyId === directProxyID}
                >
                  按代理匹配定位
                </Button>
              </div>
              {locationResult && (
                <div className="mt-2 text-xs text-[var(--color-text-muted)]">
                  {locationResult.ok
                    ? `出口 ${locationResult.ip || '-'} · ${[locationResult.country, locationResult.region, locationResult.city].filter(Boolean).join(' / ') || '-'} · ${locationResult.lang} · ${locationResult.timezone}`
                    : locationResult.error || '未匹配到定位'}
                </div>
              )}
            </FormItem>
          ) : (
            <FormItem label="本地代理地址" hint="支持 http://、https://、socks5://">
              <Input
                value={formData.proxyConfig}
                onChange={e => handleChange('proxyConfig', e.target.value)}
                placeholder="http://127.0.0.1:7890"
              />
            </FormItem>
          )}
        </div>
        <p className="text-xs text-[var(--color-text-muted)] mt-2">
          {proxyMode === 'pool'
            ? `当前使用代理池节点${selectedPoolProxy?.proxyName ? `：${selectedPoolProxy.proxyName}` : '。'}`
            : '本地代理不会进入代理池，只对当前实例保存生效。'}
        </p>
      </Card>

      <ProxyPickerModal
        open={proxyPickerOpen}
        currentProxyId={formData.proxyId}
        onSelect={proxy => { handleChange('proxyId', proxy.proxyId); setLocationResult(null) }}
        onProxyListUpdated={handleProxyListUpdated}
        onProxyDeleted={handleProxyDeleted}
        onClose={() => setProxyPickerOpen(false)}
      />

      <Card title="指纹配置" subtitle="配置浏览器指纹参数">
        <div className="flex justify-end gap-2 mb-3">
          <Button
            variant="secondary"
            size="sm"
            onClick={handleOpenFingerprintPage}
            loading={fingerprintPageOpening}
            disabled={isCreate}
          >
            浏览器内检测
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={handleFingerprintCheck}
            loading={fingerprintChecking}
            disabled={isCreate}
          >
            <ShieldCheck className="w-4 h-4" />
            自测当前实例
          </Button>
        </div>
        <FingerprintPanel
          value={formData.fingerprintArgs}
          onChange={args => handleChange('fingerprintArgs', args)}
        />
      </Card>

      <Card title="启动参数" subtitle={isCreate ? '新建时默认填入轻量参数模板，直接改这里即可' : '每行一个参数'}>
        <div className="space-y-2">
          <Textarea
            value={launchArgsText}
            onChange={e => { setLaunchArgsText(e.target.value); setIsDirty(true) }}
            rows={6}
            placeholder="--disable-sync"
          />
          {isCreate && (
            <p className="text-xs text-[var(--color-text-muted)]">这里默认就是轻量参数模板；需要更复杂的参数，直接在此基础上修改。</p>
          )}
          {!isCreate && formData.lastLaunchArgs && formData.lastLaunchArgs.length > 0 && (
            <div className="rounded-lg border border-[var(--color-border)] overflow-hidden">
              <div className="px-3 py-2 text-xs font-medium text-[var(--color-text-muted)] border-b border-[var(--color-border)]">上次实际启动参数</div>
              <Textarea
                value={formData.lastLaunchArgs.join('\n')}
                readOnly
                rows={6}
                className="border-0 rounded-none bg-[var(--color-bg-hover)] font-mono text-xs"
              />
            </div>
          )}
        </div>
      </Card>

      <ConfirmModal
        open={leaveConfirm}
        onClose={() => setLeaveConfirm(false)}
        onConfirm={() => navigate('/browser/list')}
        title="放弃未保存的更改？"
        content="当前页面有未保存的修改，离开后将丢失这些更改。"
        confirmText="放弃并离开"
        cancelText="继续编辑"
        danger
      />

      <Modal
        open={fingerprintCheckOpen}
        onClose={() => setFingerprintCheckOpen(false)}
        title="指纹自测结果"
        width="860px"
      >
        {fingerprintCheckResult ? (
          <div className="space-y-4">
            <div className="grid grid-cols-[120px_minmax(0,1fr)_minmax(0,1fr)_70px] gap-3 text-xs font-medium text-[var(--color-text-muted)] border-b border-[var(--color-border)] pb-2">
              <div>项目</div>
              <div>配置值</div>
              <div>实际值</div>
              <div>状态</div>
            </div>
            <div>
              <FingerprintCheckRow label="语言" expected={fingerprintCheckResult.expected.language} actual={fingerprintCheckResult.runtime.language} />
              <FingerprintCheckRow label="语言列表" expected={fingerprintCheckResult.expected.acceptLanguage} actual={fingerprintCheckResult.runtime.languages} />
              <FingerprintCheckRow label="时区" expected={fingerprintCheckResult.expected.timezone} actual={fingerprintCheckResult.runtime.timezone} />
              <FingerprintCheckRow label="CPU 核心" expected={fingerprintCheckResult.expected.hardwareConcurrency} actual={fingerprintCheckResult.runtime.hardwareConcurrency} />
              <FingerprintCheckRow label="窗口大小" expected={fingerprintCheckResult.expected.windowSize} actual={`${fingerprintCheckResult.runtime.innerWidth},${fingerprintCheckResult.runtime.innerHeight}`} />
              <FingerprintCheckRow label="平台" expected={fingerprintCheckResult.expected.platform} actual={fingerprintCheckResult.runtime.platform} />
              <FingerprintCheckRow label="品牌版本" expected={fingerprintCheckResult.expected.brandVersion} actual={fingerprintCheckResult.runtime.userAgent} mode="contains" />
              <FingerprintCheckRow label="系统版本" expected={fingerprintCheckResult.expected.platformVersion} actual={fingerprintCheckResult.runtime.userAgent} mode="contains" />
              <FingerprintCheckRow label="UA" expected={fingerprintCheckResult.expected.brand} actual={fingerprintCheckResult.runtime.userAgent} mode="contains" />
              <FingerprintCheckRow label="UA Data" actual={fingerprintCheckResult.runtime.userAgentData} />
              <FingerprintCheckRow label="Webdriver" expected={false} actual={fingerprintCheckResult.runtime.webdriver} />
              <FingerprintCheckRow label="内存" actual={fingerprintCheckResult.runtime.deviceMemory} />
              <FingerprintCheckRow label="屏幕" actual={`${fingerprintCheckResult.runtime.screenWidth}x${fingerprintCheckResult.runtime.screenHeight} depth ${fingerprintCheckResult.runtime.colorDepth} @ ${fingerprintCheckResult.runtime.devicePixelRatio}`} />
              <FingerprintCheckRow label="WebGL" actual={[fingerprintCheckResult.runtime.webglVendor, fingerprintCheckResult.runtime.webglRenderer].filter(Boolean).join(' / ')} />
              <FingerprintCheckRow label="Canvas Hash" actual={fingerprintCheckResult.runtime.canvasHash} />
              <FingerprintCheckRow label="Audio Hash" actual={fingerprintCheckResult.runtime.audioHash} />
              <FingerprintCheckRow label="ClientRects Hash" actual={fingerprintCheckResult.runtime.clientRectsHash} />
              <FingerprintCheckRow label="插件" actual={fingerprintCheckResult.runtime.plugins} />
              <FingerprintCheckRow label="种子" actual={fingerprintCheckResult.expected.seed || '未显式设置'} />
              <FingerprintCheckRow label="排除伪装" actual={fingerprintCheckResult.expected.disableSpoofing || '无'} />
              <FingerprintCheckRow label="WebRTC" actual={fingerprintCheckResult.expected.webrtcPolicy || '未显式设置'} />
            </div>
          </div>
        ) : (
          <div className="text-sm text-[var(--color-text-muted)]">暂无自测结果</div>
        )}
      </Modal>

      <Modal
        open={!!saveError}
        onClose={() => setSaveError('')}
        title="保存失败"
        width="420px"
        footer={<Button onClick={() => setSaveError('')}>知道了</Button>}
      >
        <div className="text-[var(--color-text-secondary)]">{saveError}</div>
      </Modal>
    </div>
  )
}
