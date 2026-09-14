import type { StructuredApiEndpointDoc } from './structuredApiDocs.types'

import {
  API_AUTH_HEADER_FIELDS,
  createLaunchSelectorFieldDocs,
} from './structuredApiDocs.fields'

const PROFILE_INPUT_FIELD_DOCS: StructuredApiEndpointDoc['fields'] = [
  { name: 'profile.profileName', type: 'string', required: false, location: 'Body', description: '实例显示名称。' },
  { name: 'profile.userDataDir', type: 'string', required: false, location: 'Body', description: '用户数据目录；创建时留空会使用新 profileId。' },
  { name: 'profile.coreId', type: 'string', required: false, location: 'Body', description: '内核 ID；创建时留空会尝试使用默认内核。' },
  { name: 'profile.restoreLastSession', type: 'string', required: false, location: 'Body', description: '会话恢复模式：on / true / enabled / enable 归一为 enabled；off / false / disabled / disable 归一为 disabled；空值、follow / default / inherit 和未知值均表示跟随全局设置。' },
  { name: 'profile.fingerprintArgs', type: 'string[]', required: false, location: 'Body', description: '持久化的指纹参数。' },
  { name: 'profile.proxyId', type: 'string', required: false, location: 'Body', description: '代理池节点 ID。' },
  { name: 'profile.proxyConfig', type: 'string', required: false, location: 'Body', description: '自定义代理配置；proxyId 有效时优先使用代理池节点。' },
  { name: 'profile.memoryLimitMb', type: 'integer', required: false, location: 'Body', description: '实例最大内存（MB）；小于等于 0 会归一为 0，表示不单独限制；当前没有上限。' },
  { name: 'profile.launchArgs', type: 'string[]', required: false, location: 'Body', description: '每次启动都使用的持久化附加参数。' },
  { name: 'profile.tags', type: 'string[]', required: false, location: 'Body', description: '实例标签。' },
  { name: 'profile.keywords', type: 'string[]', required: false, location: 'Body', description: '用于 selector 匹配的关键词。' },
  { name: 'profile.groupId', type: 'string', required: false, location: 'Body', description: '所属实例分组 ID。' },
]

const PROFILE_START_FIELD_DOCS: StructuredApiEndpointDoc['fields'] = [
  { name: 'start.launchArgs', type: 'string[]', required: false, location: 'Body', description: '仅本次自动启动使用的附加参数。' },
  { name: 'start.startUrls', type: 'string[]', required: false, location: 'Body', description: '仅本次自动启动额外打开的网址。' },
  { name: 'start.skipDefaultStartUrls', type: 'boolean', required: false, location: 'Body', description: '仅本次自动启动是否跳过默认网址。' },
  { name: 'start.proxyId', type: 'string', required: false, location: 'Body', description: '仅本次自动启动覆盖的代理池节点。' },
  { name: 'start.proxyConfig', type: 'string', required: false, location: 'Body', description: '仅本次自动启动覆盖的自定义代理。' },
]

const COMMON_API_RESPONSE_CODES: StructuredApiEndpointDoc['responseCodes'] = [
  { code: '401', description: '已启用 API Key，但请求头缺失或密钥无效。' },
  { code: '405', description: '请求方法不受支持。' },
]

export const PROFILE_API_ENDPOINT_DOCS: StructuredApiEndpointDoc[] = [
  {
    id: 'api-profiles-list-detail',
    parentId: 'api-profiles-launch',
    label: '实例列表',
    method: 'GET',
    path: '/api/profiles',
    purpose: '列出当前全部实例。',
    description: '读取当前实例目录中的全部实例配置，适合做实例选择器或管理后台列表。',
    fields: [...API_AUTH_HEADER_FIELDS],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl ${launchBaseUrl}/api/profiles \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "count": 1,
  "items": [
    {
      "profileId": "550e8400-e29b-41d4-a716-446655440000",
      "profileName": "buyer-001",
      "launchCode": "BUYER_001",
      "keywords": ["buyer-001"],
      "tags": ["电商"],
      "proxyId": "proxy-us",
      "running": false,
      "debugReady": false
    }
  ]
}`,
    },
    responseCodes: [
      { code: '200', description: '返回实例列表。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '503', description: '实例目录当前不可用。' },
    ],
    notes: [
      'items 中每项都是完整实例快照，包含配置字段、运行态字段、时间字段和 launchCode。',
    ],
  },
  {
    id: 'api-profiles-create-detail',
    parentId: 'api-profiles-launch',
    label: '创建实例',
    method: 'POST',
    path: '/api/profiles',
    purpose: '创建一个新实例，可选创建后立即启动。',
    description: '写入实例配置，必要时同时申请 launchCode，并支持通过 autoLaunch 在创建后直接启动浏览器。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'profile', type: 'object', required: true, location: 'Body', description: '实例配置主体。' },
      ...PROFILE_INPUT_FIELD_DOCS,
      { name: 'launchCode', type: 'string', required: false, location: 'Body', description: '指定实例 launchCode；自动 trim 并转大写，须为 4–32 位且仅含 A-Z、0-9、_、-；留空或省略时自动生成。' },
      { name: 'autoLaunch', type: 'boolean', required: false, location: 'Body', description: '是否在创建后立即启动。' },
      { name: 'start', type: 'object', required: false, location: 'Body', description: '仅本次自动启动时附加的启动参数。' },
      ...PROFILE_START_FIELD_DOCS,
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/profiles \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "profile": {
      "profileName": "buyer-001",
      "proxyId": "proxy-us",
      "keywords": ["buyer-001"],
      "tags": ["电商"]
    },
    "launchCode": "BUYER_001"
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "created": true,
  "updated": false,
  "launched": false,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "profileName": "buyer-001",
  "launchCode": "BUYER_001",
  "profile": {
    "profileId": "550e8400-e29b-41d4-a716-446655440000",
    "profileName": "buyer-001",
    "keywords": ["buyer-001"],
    "proxyId": "proxy-us"
  }
}`,
    },
    responseCodes: [
      { code: '201', description: '实例创建成功。' },
      { code: '400', description: '请求体非法或 profile 缺失。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '409', description: 'launchCode 冲突或实例数超限。' },
      { code: '500', description: '实例写入失败，或实例已创建但自动启动失败。' },
      { code: '503', description: '实例目录或 launchCode 服务当前不可用。' },
    ],
    notes: [
      '请求体最大 1 MiB；未知 JSON 字段会被拒绝。',
      'autoLaunch=true 时，响应会附带启动结果字段。',
      'start 仅在 autoLaunch=true 时生效；成功响应的 profile 是完整实例快照。',
      'profile.proxyId 与 profile.proxyConfig 同时传时，优先使用 proxyId 对应的代理池节点。',
      '若 proxyId 无效：提供 proxyConfig 则按自定义代理保存；未提供 proxyConfig 则返回 400。',
      '从主实例开新号请用 POST /api/profiles/{profileId}/copy。本接口创建空白实例时，省略 fingerprintArgs 会使用全局默认指纹；不要把源实例的 --fingerprint= 原样写回来。',
    ],
  },
  {
    id: 'api-profiles-get-detail',
    parentId: 'api-profiles-launch',
    label: '单个实例',
    method: 'GET',
    path: '/api/profiles/{profileId}',
    purpose: '查询单个实例配置。',
    description: '读取指定实例的完整配置快照，适合进入实例详情页或编辑页前预加载数据。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'profileId', type: 'string', required: true, location: 'Path', description: '实例 ID。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl ${launchBaseUrl}/api/profiles/550e8400-e29b-41d4-a716-446655440000 \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "profileName": "buyer-001",
  "launchCode": "BUYER_001",
  "profile": {
    "profileId": "550e8400-e29b-41d4-a716-446655440000",
    "profileName": "buyer-001",
    "keywords": ["buyer-001"],
    "tags": ["电商"],
    "proxyId": "proxy-us"
  }
}`,
    },
    responseCodes: [
      { code: '200', description: '返回实例详情。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '实例不存在。' },
      { code: '503', description: '实例目录当前不可用。' },
    ],
    notes: [],
  },
  {
    id: 'api-profiles-update-detail',
    parentId: 'api-profiles-launch',
    label: '更新实例',
    method: 'PUT',
    path: '/api/profiles/{profileId}',
    purpose: '更新指定实例配置。',
    description: '用整份 profile 配置覆盖更新实例，可选顺带更新 launchCode，并支持更新后立即启动。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'profileId', type: 'string', required: true, location: 'Path', description: '实例 ID。' },
      { name: 'profile', type: 'object', required: true, location: 'Body', description: '更新后的实例配置。' },
      ...PROFILE_INPUT_FIELD_DOCS,
      { name: 'launchCode', type: 'string', required: false, location: 'Body', description: '需要覆盖时传新值；格式为 4–32 位 A-Z、0-9、_、-，留空或省略会保留当前 launchCode。' },
      { name: 'autoLaunch', type: 'boolean', required: false, location: 'Body', description: '更新后是否直接启动。' },
      { name: 'start', type: 'object', required: false, location: 'Body', description: '仅本次自动启动时附加的启动参数。' },
      ...PROFILE_START_FIELD_DOCS,
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X PUT ${launchBaseUrl}/api/profiles/550e8400-e29b-41d4-a716-446655440000 \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "profile": {
      "profileName": "buyer-001-updated",
      "proxyId": "proxy-us",
      "keywords": ["buyer-001", "checkout"]
    }
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "created": false,
  "updated": true,
  "launched": false,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "profileName": "buyer-001-updated",
  "launchCode": "BUYER_001",
  "profile": {
    "profileId": "550e8400-e29b-41d4-a716-446655440000",
    "profileName": "buyer-001-updated",
    "keywords": ["buyer-001", "checkout"],
    "proxyId": "proxy-us"
  }
}`,
    },
    responseCodes: [
      { code: '200', description: '更新成功。' },
      { code: '400', description: '请求体非法。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '实例不存在。' },
      { code: '409', description: '新的 launchCode 与其他实例冲突。' },
      { code: '500', description: '实例更新失败，或实例已更新但自动启动失败。' },
      { code: '503', description: '实例目录或 launchCode 服务当前不可用。' },
    ],
    notes: [
      '请求体最大 1 MiB；未知 JSON 字段会被拒绝。',
      'profile 是整份覆盖更新，不是 PATCH；未提供的 profile 子字段会按零值保存。',
      'start 仅在 autoLaunch=true 时生效；成功响应的 profile 是完整实例快照。',
      'profile.proxyId 与 profile.proxyConfig 同时传时，优先使用 proxyId 对应的代理池节点。',
      '若 proxyId 无效：提供 proxyConfig 则按自定义代理保存；未提供 proxyConfig 则返回 400。',
    ],
  },
  {
    id: 'api-profiles-delete-detail',
    parentId: 'api-profiles-launch',
    label: '删除实例',
    method: 'DELETE',
    path: '/api/profiles/{profileId}',
    purpose: '删除一个未运行中的实例。',
    description: '删除实例配置并移除关联 launchCode；运行中的实例会被直接拒绝删除。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'profileId', type: 'string', required: true, location: 'Path', description: '实例 ID。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X DELETE ${launchBaseUrl}/api/profiles/550e8400-e29b-41d4-a716-446655440000 \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "deleted": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "profileName": "buyer-001",
  "launchCode": "BUYER_001"
}`,
    },
    responseCodes: [
      { code: '200', description: '删除成功。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '实例不存在。' },
      { code: '409', description: '实例仍在运行，不能直接删除。' },
      { code: '500', description: '删除实例配置失败。' },
      { code: '503', description: '实例目录当前不可用。' },
    ],
    notes: [
      '运行中的实例先 stop，再 delete。',
    ],
  },
  {
    id: 'api-profiles-copy-detail',
    parentId: 'api-profiles-launch',
    label: '复制实例',
    method: 'POST',
    path: '/api/profiles/{profileId}/copy',
    purpose: '从已有实例复制配置，生成新实例。',
    description: '走应用内 CopyWithOptions：拷内核、代理、分组、标签、关键字、启动参数和会话恢复设置；新实例使用新的 profileId / userDataDir / launchCode。默认 auto_fingerprint 只重生成指纹种子。不复制 Cookie、登录态和用户数据目录。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'profileId', type: 'string', required: true, location: 'Path', description: '源实例 ID。' },
      { name: 'profileName', type: 'string', required: false, location: 'Body', description: '新实例名称；留空或省略时自动生成「源名称（副本）时间戳」。' },
      { name: 'mode', type: 'auto_fingerprint | regular', required: false, location: 'Body', description: '指纹复制模式。默认 auto_fingerprint：保留源实例指纹模板，只重生成勾选的项。regular：指纹参数原样保留。' },
      { name: 'automationTargets', type: 'string[]', required: false, location: 'Body', description: 'auto_fingerprint 时要重生成的指纹项。默认 ["seed"]。可选 seed / identity / locale / screen / hardware / network；旧别名 render / fonts / devices 会映射到 seed。mode=regular 时忽略。' },
      { name: 'launchCode', type: 'string', required: false, location: 'Body', description: '指定新实例 launchCode；自动 trim 并转大写，须为 4–32 位且仅含 A-Z、0-9、_、-；留空或省略时自动生成。' },
      { name: 'proxyId', type: 'string', required: false, location: 'Body', description: '覆盖新实例代理池节点。省略时跟随源实例。' },
      { name: 'proxyConfig', type: 'string', required: false, location: 'Body', description: '覆盖新实例自定义代理。proxyId 有效时优先使用代理池节点。' },
      { name: 'autoLaunch', type: 'boolean', required: false, location: 'Body', description: '是否在复制后立即启动新实例。' },
      { name: 'start', type: 'object', required: false, location: 'Body', description: '仅本次自动启动时附加的启动参数。' },
      ...PROFILE_START_FIELD_DOCS,
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/profiles/550e8400-e29b-41d4-a716-446655440000/copy \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "profileName": "task-001",
    "mode": "auto_fingerprint",
    "automationTargets": ["seed"],
    "proxyId": "proxy-us",
    "autoLaunch": true,
    "start": {
      "skipDefaultStartUrls": true
    }
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "created": true,
  "updated": false,
  "copied": true,
  "launched": false,
  "profileId": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "profileName": "task-001",
  "launchCode": "TASK_001",
  "sourceProfileId": "550e8400-e29b-41d4-a716-446655440000",
  "profile": {
    "profileId": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "profileName": "task-001",
    "userDataDir": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "keywords": ["main"],
    "proxyId": "proxy-us"
  }
}`,
    },
    responseCodes: [
      { code: '201', description: '实例复制成功。' },
      { code: '400', description: '请求体非法、mode 不受支持，或 automationTargets 非法。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '源实例不存在。' },
      { code: '409', description: 'launchCode 冲突或实例数超限。' },
      { code: '500', description: '实例复制失败，或实例已复制但自动启动失败。' },
      { code: '503', description: '实例目录或 launchCode 服务当前不可用。' },
    ],
    notes: [
      '请求体可省略；空 body 或 {} 使用默认名称和 auto_fingerprint。请求体最大 1 MiB，未知 JSON 字段会被拒绝。',
      '默认只重生成指纹种子。多号隔离不要用 mode=regular，否则会和源实例共用同一套指纹。',
      '复制只拷配置，不拷 Cookie 和用户数据目录。要带走登录态用导出 ZIP。',
      '脚本 targetConfig.mode=create 仍走 Copy()：指纹参数整份替换成全局默认，不是本接口的 auto_fingerprint。',
      'autoLaunch=true 时，响应会附带启动结果字段。start 仅在 autoLaunch=true 时生效。',
      'proxyId / proxyConfig 都省略时跟随源实例代理。自定义代理只传 proxyConfig。无效 proxyId 且未提供 proxyConfig 会 400。',
      '只有 launchCode 时用 POST /api/profiles/copy，在 selector.code 里指定源实例。',
    ],
  },
  {
    id: 'api-profiles-copy-selector-detail',
    parentId: 'api-profiles-launch',
    label: '按 selector 复制',
    method: 'POST',
    path: '/api/profiles/copy',
    purpose: '按 selector 或 launchCode 复制实例，不必先查 profileId。',
    description: '与 POST /api/profiles/{profileId}/copy 同一套 CopyWithOptions。selector 只支持 unique / first，不支持 matchMode=all。profileName 是新实例名称，源实例名称放在 selector.profileName。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'selector', type: 'object', required: false, location: 'Body', description: '源实例选择条件；与 code / profileId 至少提供一组。' },
      ...createLaunchSelectorFieldDocs(
        'selector',
        'unique | first',
        '匹配策略；省略时按 runtime 规则默认为 unique。不支持 all。',
      ),
      { name: 'code', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.code。' },
      { name: 'profileId', type: 'string', required: false, location: 'Body', description: '兼容写法：源实例 ID，等价于 selector.profileId。' },
      { name: 'profileName', type: 'string', required: false, location: 'Body', description: '新实例名称，不是源实例名称。' },
      { name: 'mode', type: 'auto_fingerprint | regular', required: false, location: 'Body', description: '指纹复制模式，默认 auto_fingerprint。' },
      { name: 'automationTargets', type: 'string[]', required: false, location: 'Body', description: 'auto_fingerprint 时要重生成的指纹项，默认 ["seed"]。' },
      { name: 'proxyId', type: 'string', required: false, location: 'Body', description: '覆盖新实例代理池节点。' },
      { name: 'proxyConfig', type: 'string', required: false, location: 'Body', description: '覆盖新实例自定义代理。' },
      { name: 'launchCode', type: 'string', required: false, location: 'Body', description: '指定新实例 launchCode。' },
      { name: 'autoLaunch', type: 'boolean', required: false, location: 'Body', description: '是否在复制后立即启动。' },
      { name: 'start', type: 'object', required: false, location: 'Body', description: '仅本次自动启动时附加的启动参数。' },
      ...PROFILE_START_FIELD_DOCS,
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/profiles/copy \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "selector": { "code": "MAIN_001" },
    "profileName": "task-001",
    "proxyConfig": "http://127.0.0.1:7890"
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "copied": true,
  "created": true,
  "profileId": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "profileName": "task-001",
  "sourceProfileId": "550e8400-e29b-41d4-a716-446655440000",
  "launchCode": "TASK_001"
}`,
    },
    responseCodes: [
      { code: '201', description: '实例复制成功。' },
      { code: '400', description: 'selector 缺失、matchMode=all，或代理参数非法。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: 'selector 未命中源实例。' },
      { code: '409', description: 'selector 命中多个实例，或 launchCode 冲突。' },
      { code: '500', description: '复制失败，或已复制但自动启动失败。' },
      { code: '503', description: '实例目录当前不可用。' },
    ],
    notes: [
      'selector 与顶层 code / profileId 至少要能唯一命中一个源实例。',
      '顶层 profileName 永远是新实例名称；源实例名称只用 selector.profileName。',
    ],
  },
  {
    id: 'api-profiles-status-detail',
    parentId: 'api-profiles-launch',
    label: '实例状态',
    method: 'GET',
    path: '/api/profiles/{profileId}/status',
    purpose: '查询单个实例的实时运行态。',
    description: '返回运行中、debugReady、cdpUrl 等运行态字段，适合精确观察单个实例当前状态。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'profileId', type: 'string', required: true, location: 'Path', description: '实例 ID。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl ${launchBaseUrl}/api/profiles/550e8400-e29b-41d4-a716-446655440000/status \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "profileName": "buyer-001",
  "launchCode": "BUYER_001",
  "running": true,
  "pid": 10240,
  "debugPort": 9333,
  "debugReady": true,
  "runtimeWarning": "",
  "lastError": "",
  "lastStartAt": "2026-08-08T10:00:00+08:00",
  "lastStopAt": "",
  "active": true,
  "cdpPort": 19876,
  "cdpUrl": "http://127.0.0.1:19876",
  "directDebugUrl": "http://127.0.0.1:9333",
  "profile": { "profileId": "550e8400-e29b-41d4-a716-446655440000" }
}`,
    },
    responseCodes: [
      { code: '200', description: '返回实例运行态。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '实例不存在。' },
      { code: '503', description: '运行态查询或实例目录当前不可用。' },
    ],
    notes: [
      '响应同时包含完整 profile 快照，以及 pid、debugPort、runtimeWarning、最近错误和启停时间等诊断字段。',
    ],
  },
  {
    id: 'api-profiles-stop-detail',
    parentId: 'api-profiles-launch',
    label: '停止实例',
    method: 'POST',
    path: '/api/profiles/{profileId}/stop',
    purpose: '精确停止一个指定实例。',
    description: '按 profileId 停止实例，适合任务完成后的精确回收。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'profileId', type: 'string', required: true, location: 'Path', description: '实例 ID。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/profiles/550e8400-e29b-41d4-a716-446655440000/stop \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "stopped": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "running": false,
  "pid": 0,
  "debugPort": 0,
  "debugReady": false,
  "runtimeWarning": "",
  "lastError": "",
  "lastStartAt": "2026-08-08T10:00:00+08:00",
  "lastStopAt": "2026-08-08T10:15:00+08:00",
  "active": false,
  "cdpPort": 0,
  "cdpUrl": "",
  "directDebugUrl": "",
  "profile": { "profileId": "550e8400-e29b-41d4-a716-446655440000" }
}`,
    },
    responseCodes: [
      { code: '200', description: '停止成功。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '实例不存在。' },
      { code: '500', description: '停止实例失败。' },
      { code: '503', description: '当前环境不支持运行态控制。' },
    ],
    notes: [
      '响应使用统一运行态结构，并附带停止后的完整 profile 快照。',
    ],
  },
  {
    id: 'api-profiles-restart-detail',
    parentId: 'api-profiles-launch',
    label: '重启实例',
    method: 'POST',
    path: '/api/profiles/{profileId}/restart',
    purpose: '停止后立即重新启动指定实例。',
    description: '等价于 stop 再 start，但作为一次调用，避免中间状态被外部编排拆开。成功后如果 debugReady，会切到统一 CDP 入口。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'profileId', type: 'string', required: true, location: 'Path', description: '实例 ID。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/profiles/550e8400-e29b-41d4-a716-446655440000/restart \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "restarted": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "running": true,
  "debugReady": true,
  "cdpUrl": "http://127.0.0.1:19876"
}`,
    },
    responseCodes: [
      { code: '200', description: '重启成功。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '实例不存在。' },
      { code: '500', description: '停止或启动失败。' },
      { code: '503', description: '当前环境不支持运行态控制。' },
    ],
    notes: [
      '实例本来没运行时，效果接近直接启动。',
    ],
  },
  {
    id: 'api-profiles-open-url-detail',
    parentId: 'api-profiles-launch',
    label: '打开页面',
    method: 'POST',
    path: '/api/profiles/{profileId}/open-url',
    purpose: '在已经运行的实例里打开一个网址。',
    description: '不启动新实例。优先走 CDP 打开新标签，失败时回退到浏览器进程唤起。实例必须已经 running。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'profileId', type: 'string', required: true, location: 'Path', description: '实例 ID。' },
      { name: 'url', type: 'string', required: true, location: 'Body', description: '要打开的网址。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/profiles/550e8400-e29b-41d4-a716-446655440000/open-url \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{ "url": "https://example.com" }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "opened": true,
  "url": "https://example.com",
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "running": true,
  "debugReady": true
}`,
    },
    responseCodes: [
      { code: '200', description: '已打开页面。' },
      { code: '400', description: 'url 缺失或请求体非法。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '实例不存在。' },
      { code: '409', description: '实例未运行或调试接口未就绪。' },
      { code: '500', description: '打开页面失败。' },
      { code: '503', description: '当前环境不支持运行态控制。' },
    ],
    notes: [
      '启动时打开页面请用 launch / runtime/session 的 startUrls。这个接口只服务已经在跑的实例。',
    ],
  },
  {
    id: 'api-profiles-cookies-get-detail',
    parentId: 'api-profiles-launch',
    label: '导出 Cookie',
    method: 'GET',
    path: '/api/profiles/{profileId}/cookies',
    purpose: '导出运行中实例的 Cookie。',
    description: '通过 CDP 读取当前浏览器 Cookie。默认返回 JSON，同时带 Netscape 文本。format=netscape 时直接返回文本。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'profileId', type: 'string', required: true, location: 'Path', description: '实例 ID。' },
      { name: 'format', type: 'json | netscape', required: false, location: 'Query', description: 'netscape 时返回 text/plain。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl ${launchBaseUrl}/api/profiles/550e8400-e29b-41d4-a716-446655440000/cookies \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "count": 1,
  "cookies": [
    {
      "name": "session",
      "value": "abc123",
      "domain": ".example.com",
      "path": "/",
      "secure": true
    }
  ],
  "netscape": "# Netscape HTTP Cookie File\\n..."
}`,
    },
    responseCodes: [
      { code: '200', description: '返回 Cookie。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '实例不存在。' },
      { code: '409', description: '实例未运行或调试接口未就绪。' },
      { code: '503', description: 'Cookie 能力当前不可用。' },
    ],
    notes: [
      '实例必须 running 且 debugReady。',
      'Cookie 含登录态，只建议本机调用。',
    ],
  },
  {
    id: 'api-profiles-cookies-put-detail',
    parentId: 'api-profiles-launch',
    label: '导入 Cookie',
    method: 'PUT',
    path: '/api/profiles/{profileId}/cookies',
    purpose: '向运行中实例写入 Cookie。',
    description: '通过 CDP Network.setCookies 导入。可传 JSON cookies 数组，或 Netscape 文本。JSON 优先。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'profileId', type: 'string', required: true, location: 'Path', description: '实例 ID。' },
      { name: 'cookies', type: 'object[]', required: false, location: 'Body', description: 'Cookie 对象数组；与 netscape 至少提供一个。' },
      { name: 'netscape', type: 'string', required: false, location: 'Body', description: 'Netscape HTTP Cookie File 文本。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X PUT ${launchBaseUrl}/api/profiles/550e8400-e29b-41d4-a716-446655440000/cookies \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "cookies": [
      { "name": "session", "value": "abc123", "domain": ".example.com", "path": "/" }
    ]
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "imported": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "count": 1
}`,
    },
    responseCodes: [
      { code: '200', description: '导入成功。' },
      { code: '400', description: 'cookies / netscape 缺失或 Netscape 文本非法。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '实例不存在。' },
      { code: '409', description: '实例未运行或调试接口未就绪。' },
      { code: '500', description: 'CDP 写入失败。' },
      { code: '503', description: 'Cookie 能力当前不可用。' },
    ],
    notes: [
      '也接受 POST，语义与 PUT 相同。',
      '导入不会复制用户数据目录，只写入当前运行中浏览器的 Cookie。',
    ],
  },
  {
    id: 'api-launch-code-detail',
    parentId: 'api-profiles-launch',
    label: '按 Code 启动',
    method: 'GET',
    path: '/api/launch/{code}',
    purpose: '按唯一 launchCode 启动实例。',
    description: '最短路径的启动接口，适合外部系统已经拿到唯一 launchCode 的场景。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'code', type: 'string', required: true, location: 'Path', description: '实例 launchCode。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl ${launchBaseUrl}/api/launch/BUYER_001 \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "profileName": "buyer-001",
  "launchCode": "BUYER_001",
  "pid": 10240,
  "debugPort": 9333,
  "debugReady": true,
  "runtimeWarning": "",
  "cdpPort": 19876,
  "cdpUrl": "http://127.0.0.1:19876"
}`,
    },
    responseCodes: [
      { code: '200', description: '启动成功。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: 'launchCode 不存在。' },
      { code: '500', description: '浏览器启动失败。' },
      { code: '503', description: 'launchCode 服务当前不可用。' },
    ],
    notes: [
      'cdpUrl 是 LaunchServer 的统一 CDP 入口；directDebugUrl 如需直连可从运行态接口读取。',
    ],
  },
  {
    id: 'api-launch-body-detail',
    parentId: 'api-profiles-launch',
    label: '按 selector 启动',
    method: 'POST',
    path: '/api/launch',
    purpose: '按 selector 或兼容顶层字段启动实例。',
    description: '更灵活的启动入口，支持 selector、兼容顶层选择字段、launchArgs、startUrls 和 skipDefaultStartUrls 等临时参数。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'selector', type: 'object', required: false, location: 'Body', description: '目标实例选择条件；新接入推荐使用。' },
      ...createLaunchSelectorFieldDocs(
        'selector',
        'unique | first | all',
        '匹配策略；省略时，code / key / keywords 使用 first，其他选择条件使用 unique。',
      ),
      { name: 'code', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.code。' },
      { name: 'key', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.key，优先匹配实例关键词。' },
      { name: 'profileId', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.profileId。' },
      { name: 'profileName', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.profileName。' },
      { name: 'keyword', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.keyword。' },
      { name: 'keywords', type: 'string[]', required: false, location: 'Body', description: '兼容写法：等价于 selector.keywords。' },
      { name: 'tag', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.tag。' },
      { name: 'tags', type: 'string[]', required: false, location: 'Body', description: '兼容写法：等价于 selector.tags。' },
      { name: 'groupId', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.groupId。' },
      { name: 'matchMode', type: 'unique | first | all', required: false, location: 'Body', description: '兼容写法：等价于 selector.matchMode。' },
      { name: 'launchArgs', type: 'string[]', required: false, location: 'Body', description: '本次启动的临时附加参数。' },
      { name: 'startUrls', type: 'string[]', required: false, location: 'Body', description: '本次启动后额外打开的网址。' },
      { name: 'skipDefaultStartUrls', type: 'boolean', required: false, location: 'Body', description: '是否跳过实例默认启动 URL。' },
      { name: 'proxyId', type: 'string', required: false, location: 'Body', description: '仅本次启动覆盖的代理池节点，不修改实例原配置。' },
      { name: 'proxyConfig', type: 'string', required: false, location: 'Body', description: '仅本次启动覆盖的自定义代理，不修改实例原配置。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/launch \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "selector": {
      "keyword": "checkout",
      "tags": ["电商", "北美"],
      "matchMode": "unique"
    },
    "skipDefaultStartUrls": true
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "profileName": "buyer-001",
  "launchCode": "BUYER_001",
  "debugReady": true,
  "cdpUrl": "http://127.0.0.1:19876"
}`,
    },
    responseCodes: [
      { code: '200', description: '启动成功。' },
      { code: '400', description: 'selector 缺失或请求体非法。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: 'selector 未命中实例。' },
      { code: '409', description: 'selector 命中多个实例。' },
      { code: '500', description: '实例匹配或浏览器启动失败。' },
      { code: '503', description: '实例目录或启动能力当前不可用。' },
    ],
    notes: [
      '请求体最大 1 MiB；未知 JSON 字段会被拒绝。',
      'selector 为空且没有任何兼容顶层选择字段时返回 400。',
      'matchMode=all 只在这个接口可用。',
      '默认 matchMode：按 code / key / keywords 选择时为 first，其他选择条件为 unique。',
      'matchMode=all 时响应包含 matchMode、count、items 和 activeProfileId；不是单实例响应结构。',
      'proxyId / proxyConfig 只影响本次启动，不覆盖实例原代理。',
    ],
  },
]
