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
