import type { StructuredApiField } from './structuredApiDocs.types'

export const API_AUTH_HEADER_FIELDS: StructuredApiField[] = [
  {
    name: 'X-Ant-Api-Key',
    type: 'string',
    required: false,
    location: 'Header',
    description: '条件必填：启用且配置 API Key 时必须提供。Header 名可配置，默认 X-Ant-Api-Key。',
  },
]

export function createLaunchSelectorFieldDocs(
  prefix: string,
  matchModeType: string,
  matchModeDescription: string,
): StructuredApiField[] {
  const name = (field: string) => `${prefix}.${field}`

  return [
    { name: name('code'), type: 'string', required: false, location: 'Body', description: '按唯一 launchCode 选择实例。' },
    { name: name('key'), type: 'string', required: false, location: 'Body', description: '单值快捷检索键，优先用于实例关键词匹配。' },
    { name: name('profileId'), type: 'string', required: false, location: 'Body', description: '按实例 ID 选择。' },
    { name: name('profileName'), type: 'string', required: false, location: 'Body', description: '按实例名称选择。' },
    { name: name('keyword'), type: 'string', required: false, location: 'Body', description: '单个实例关键词；会与 keywords 合并。' },
    { name: name('keywords'), type: 'string[]', required: false, location: 'Body', description: '实例关键词列表。' },
    { name: name('tag'), type: 'string', required: false, location: 'Body', description: '单个实例标签；会与 tags 合并。' },
    { name: name('tags'), type: 'string[]', required: false, location: 'Body', description: '实例标签列表。' },
    { name: name('groupId'), type: 'string', required: false, location: 'Body', description: '按实例分组 ID 选择。' },
    { name: name('matchMode'), type: matchModeType, required: false, location: 'Body', description: matchModeDescription },
  ]
}

export function createAutomationTargetSelectorFieldDocs(prefix: string): StructuredApiField[] {
  const name = (field: string) => `${prefix}.${field}`

  return [
    { name: name('code'), type: 'string', required: false, location: 'Body', description: '按 launchCode 选择实例或模板。' },
    { name: name('profileId'), type: 'string', required: false, location: 'Body', description: '按实例 ID 选择实例或模板。' },
    { name: name('profileName'), type: 'string', required: false, location: 'Body', description: '按实例名称选择实例或模板。' },
    { name: name('groupId'), type: 'string', required: false, location: 'Body', description: '按实例分组 ID 选择实例或模板。' },
    { name: name('keywords'), type: 'string[]', required: false, location: 'Body', description: '按关键词列表选择实例或模板。' },
    { name: name('tags'), type: 'string[]', required: false, location: 'Body', description: '按标签列表选择实例或模板。' },
  ]
}
