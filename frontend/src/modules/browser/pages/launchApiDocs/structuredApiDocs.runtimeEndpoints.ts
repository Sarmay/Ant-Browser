import type { StructuredApiEndpointDoc } from './structuredApiDocs.types'

import {
  API_AUTH_HEADER_FIELDS,
  createLaunchSelectorFieldDocs,
} from './structuredApiDocs.fields'

const COMMON_API_RESPONSE_CODES: StructuredApiEndpointDoc['responseCodes'] = [
  { code: '401', description: '已启用 API Key，但请求头缺失或密钥无效。' },
  { code: '405', description: '请求方法不受支持。' },
]

export const RUNTIME_API_ENDPOINT_DOCS: StructuredApiEndpointDoc[] = [
  {
    id: 'api-runtime-active-detail',
    parentId: 'api-runtime',
    label: '当前活动实例',
    method: 'GET',
    path: '/api/runtime/active',
    purpose: '查看当前统一 CDP 入口挂着哪个实例。',
    description: '当外部系统只知道 LaunchServer 端口、不知道当前 active target 时，先查这个接口最直接。',
    fields: [...API_AUTH_HEADER_FIELDS],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl ${launchBaseUrl}/api/runtime/active \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "active": true,
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
  "cdpPort": 19876,
  "cdpUrl": "http://127.0.0.1:19876",
  "directDebugUrl": "http://127.0.0.1:9333",
  "profile": { "profileId": "550e8400-e29b-41d4-a716-446655440000" }
}`,
    },
    responseCodes: [
      { code: '200', description: '返回当前 active target 状态。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '503', description: '活动实例存在，但运行态查询暂不可用。' },
    ],
    notes: [
      'active=false 表示当前没有活动实例。',
      'active=true 时响应包含完整 profile 快照、进程/调试端口、最近错误和启停时间。',
    ],
  },
  {
    id: 'api-runtime-session-detail',
    parentId: 'api-runtime',
    label: '准备可接管会话',
    method: 'POST',
    path: '/api/runtime/session',
    purpose: '准备一个可 attach 的运行时会话。',
    description: '按 selector 命中实例，必要时自动启动，并在给定超时时间内等待 debugReady=true。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'selector', type: 'object', required: false, location: 'Body', description: '目标实例选择条件；新接入推荐使用。' },
      ...createLaunchSelectorFieldDocs(
        'selector',
        'unique | first',
        '匹配策略；默认 unique，运行态接口不支持 all。',
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
      { name: 'matchMode', type: 'unique | first', required: false, location: 'Body', description: '兼容写法：等价于 selector.matchMode；默认 unique，不支持 all。' },
      { name: 'timeoutMs', type: 'integer', required: false, location: 'Body', description: '等待 debugReady 的超时；默认 45000，实际值会裁剪到 1000–120000。' },
      { name: 'startUrls', type: 'string[]', required: false, location: 'Body', description: '本次启动时额外打开的网址。' },
      { name: 'skipDefaultStartUrls', type: 'boolean', required: false, location: 'Body', description: '是否跳过实例默认启动 URL。' },
      { name: 'launchArgs', type: 'string[]', required: false, location: 'Body', description: '本次启动时临时附加的启动参数。' },
      { name: 'proxyId', type: 'string', required: false, location: 'Body', description: '仅本次启动覆盖的代理池节点，不修改实例原配置。' },
      { name: 'proxyConfig', type: 'string', required: false, location: 'Body', description: '仅本次启动覆盖的自定义代理，不修改实例原配置。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/runtime/session \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "selector": { "code": "BUYER_001" },
    "timeoutMs": 45000,
    "skipDefaultStartUrls": true
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "ready": true,
  "waitTimedOut": false,
  "retryable": false,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "launchCode": "BUYER_001",
  "running": true,
  "pid": 10240,
  "debugPort": 9333,
  "debugReady": true,
  "runtimeWarning": "",
  "lastError": "",
  "active": true,
  "cdpPort": 19876,
  "cdpUrl": "http://127.0.0.1:19876",
  "directDebugUrl": "http://127.0.0.1:9333",
  "profile": { "profileId": "550e8400-e29b-41d4-a716-446655440000" },
  "timeoutMs": 45000
}`,
    },
    responseCodes: [
      { code: '200', description: '实例已 ready，可直接 attach。' },
      { code: '202', description: '实例已处理但暂未 ready，可稍后重试。' },
      { code: '400', description: 'selector 缺失或 matchMode 非法。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '目标实例不存在。' },
      { code: '409', description: 'matchMode=unique 时 selector 命中多个实例。' },
      { code: '500', description: '实例启动或等待 debugReady 失败。' },
      { code: '503', description: '实例目录、启动能力或运行态查询当前不可用。' },
    ],
    notes: [
      '请求体最大 1 MiB；未知 JSON 字段会被拒绝。',
      'selector 为空且没有任何兼容顶层选择字段时返回 400。',
      '200 表示 ready，可直接接管。',
      '202 表示未 ready，需要重试。',
      '响应回显裁剪后的 timeoutMs，并包含完整统一运行态字段与 profile 快照。',
      'proxyId / proxyConfig 只影响本次启动，不覆盖实例原代理。',
    ],
  },
  {
    id: 'api-runtime-status-detail',
    parentId: 'api-runtime',
    label: '按 selector 查状态',
    method: 'POST',
    path: '/api/runtime/status',
    purpose: '按 selector 查询实例当前运行态。',
    description: '不启动新实例，不等待 ready，只看当前 selector 命中的实例状态。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'selector', type: 'object', required: false, location: 'Body', description: '目标实例选择条件；新接入推荐使用。' },
      ...createLaunchSelectorFieldDocs(
        'selector',
        'unique | first',
        '匹配策略；默认 unique，运行态接口不支持 all。',
      ),
      { name: 'code', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.code。' },
      { name: 'key', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.key。' },
      { name: 'profileId', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.profileId。' },
      { name: 'profileName', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.profileName。' },
      { name: 'keyword', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.keyword。' },
      { name: 'keywords', type: 'string[]', required: false, location: 'Body', description: '兼容写法：等价于 selector.keywords。' },
      { name: 'tag', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.tag。' },
      { name: 'tags', type: 'string[]', required: false, location: 'Body', description: '兼容写法：等价于 selector.tags。' },
      { name: 'groupId', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.groupId。' },
      { name: 'matchMode', type: 'unique | first', required: false, location: 'Body', description: '兼容写法：等价于 selector.matchMode；默认 unique，不支持 all。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/runtime/status \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "selector": { "keyword": "checkout", "matchMode": "first" }
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "launchCode": "BUYER_001",
  "running": true,
  "pid": 10240,
  "debugPort": 9333,
  "debugReady": false,
  "runtimeWarning": "浏览器仍在初始化",
  "lastError": "",
  "active": false,
  "cdpPort": 0,
  "cdpUrl": "",
  "directDebugUrl": "http://127.0.0.1:9333",
  "profile": { "profileId": "550e8400-e29b-41d4-a716-446655440000" }
}`,
    },
    responseCodes: [
      { code: '200', description: '返回运行态。' },
      { code: '400', description: 'selector 缺失或 matchMode 非法。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '目标实例不存在。' },
      { code: '409', description: 'matchMode=unique 时 selector 命中多个实例。' },
      { code: '500', description: '实例匹配或运行态查询失败。' },
      { code: '503', description: '实例目录或运行态查询当前不可用。' },
    ],
    notes: [
      '请求体最大 1 MiB；未知 JSON 字段会被拒绝。',
      '不会启动实例。',
      'selector 为空且没有任何兼容顶层选择字段时返回 400。',
      '响应使用统一运行态结构，并附带完整 profile 快照。',
    ],
  },
  {
    id: 'api-runtime-stop-detail',
    parentId: 'api-runtime',
    label: '按 selector 停止',
    method: 'POST',
    path: '/api/runtime/stop',
    purpose: '按 selector 停止实例。',
    description: '和 runtime/status 一样使用 selector，但动作改为停止实例，适合编排侧做统一回收。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'selector', type: 'object', required: false, location: 'Body', description: '目标实例选择条件；新接入推荐使用。' },
      ...createLaunchSelectorFieldDocs(
        'selector',
        'unique | first',
        '匹配策略；默认 unique，运行态接口不支持 all。',
      ),
      { name: 'code', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.code。' },
      { name: 'key', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.key。' },
      { name: 'profileId', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.profileId。' },
      { name: 'profileName', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.profileName。' },
      { name: 'keyword', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.keyword。' },
      { name: 'keywords', type: 'string[]', required: false, location: 'Body', description: '兼容写法：等价于 selector.keywords。' },
      { name: 'tag', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.tag。' },
      { name: 'tags', type: 'string[]', required: false, location: 'Body', description: '兼容写法：等价于 selector.tags。' },
      { name: 'groupId', type: 'string', required: false, location: 'Body', description: '兼容写法：等价于 selector.groupId。' },
      { name: 'matchMode', type: 'unique | first', required: false, location: 'Body', description: '兼容写法：等价于 selector.matchMode；默认 unique，不支持 all。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/runtime/stop \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "selector": { "code": "BUYER_001" }
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "stopped": true,
  "profileId": "550e8400-e29b-41d4-a716-446655440000",
  "launchCode": "BUYER_001",
  "running": false,
  "pid": 0,
  "debugPort": 0,
  "debugReady": false,
  "runtimeWarning": "",
  "lastError": "",
  "active": false,
  "cdpPort": 0,
  "cdpUrl": "",
  "directDebugUrl": "",
  "profile": { "profileId": "550e8400-e29b-41d4-a716-446655440000" }
}`,
    },
    responseCodes: [
      { code: '200', description: '停止成功。' },
      { code: '400', description: 'selector 缺失或 matchMode 非法。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '404', description: '目标实例不存在。' },
      { code: '409', description: 'matchMode=unique 时 selector 命中多个实例。' },
      { code: '500', description: '实例匹配或停止失败。' },
      { code: '503', description: '实例目录或停止能力当前不可用。' },
    ],
    notes: [
      '请求体最大 1 MiB；未知 JSON 字段会被拒绝。',
      'selector 为空且没有任何兼容顶层选择字段时返回 400。',
      '不支持 matchMode=all。',
      '响应使用统一运行态结构，并附带停止后的完整 profile 快照。',
    ],
  },
  {
    id: 'api-cdp-version-detail',
    parentId: 'api-runtime',
    label: 'CDP 版本信息',
    method: 'GET',
    path: '/json/version',
    purpose: '读取统一 CDP 入口的版本信息。',
    description: '这个接口透传当前 active target 的 CDP 版本信息，适合 attach 前探测调试入口是否可用。',
    fields: [],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl }) => `curl ${launchBaseUrl}/json/version`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "Browser": "Chrome/<version>",
  "Protocol-Version": "1.3",
  "User-Agent": "Mozilla/5.0",
  "webSocketDebuggerUrl": "ws://127.0.0.1:9333/devtools/browser/<id>"
}`,
    },
    responseCodes: [
      { code: '200', description: '返回当前 active target 的版本信息。' },
      { code: '502', description: '已记录 active target，但 CDP 上游无法连接。' },
      { code: '503', description: '当前没有可透传的 active target。' },
    ],
    notes: [
      '无 active target 时返回 503。',
      '响应由活动浏览器原样透传，webSocketDebuggerUrl 通常仍指向浏览器 debugPort，不会自动改写为 LaunchServer 地址。',
      '/json/* 不属于 /api/*，不会校验 API Key，但仍只允许 localhost 访问。',
    ],
  },
  {
    id: 'api-cdp-list-detail',
    parentId: 'api-runtime',
    label: 'CDP Target 列表',
    method: 'GET',
    path: '/json/list',
    purpose: '读取统一 CDP 入口当前暴露的 target 列表。',
    description: '给 Playwright、Puppeteer 或诊断工具查看当前活动 target 时使用。',
    fields: [],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl }) => `curl ${launchBaseUrl}/json/list`,
    },
    responseExample: {
      language: 'json',
      code: () => `[
  {
    "id": "page-1",
    "type": "page",
    "title": "Checkout",
    "url": "https://example.com/checkout",
    "webSocketDebuggerUrl": "ws://127.0.0.1:9333/devtools/page/page-1"
  }
]`,
    },
    responseCodes: [
      { code: '200', description: '返回当前活动实例的 target 列表。' },
      { code: '502', description: '已记录 active target，但 CDP 上游无法连接。' },
      { code: '503', description: '当前没有可透传的 active target。' },
    ],
    notes: [
      '无 active target 时返回 503。',
      '响应由活动浏览器原样透传，webSocketDebuggerUrl 通常仍指向浏览器 debugPort。',
      '/json/* 不属于 /api/*，不会校验 API Key，但仍只允许 localhost 访问。',
    ],
  },
  {
    id: 'api-cdp-ws-detail',
    parentId: 'api-runtime',
    label: 'CDP WebSocket',
    method: 'WS',
    path: '/devtools/...',
    purpose: '通过统一 WebSocket 入口接管当前活动实例。',
    description: '实际 attach 时使用的就是这个 WebSocket 入口。外部工具通常先拿 /json/version 或 /json/list，再连对应 websocketDebuggerUrl。',
    fields: [],
    requestExample: {
      language: 'javascript',
      code: ({ launchBaseUrl }) => `const gateway = new URL("${launchBaseUrl}");
const version = await fetch(new URL("/json/version", gateway)).then((res) => res.json());
const endpoint = new URL(version.webSocketDebuggerUrl);

// 保留 /devtools/... 路径，只把 origin 换成 LaunchServer。
endpoint.protocol = gateway.protocol === "https:" ? "wss:" : "ws:";
endpoint.host = gateway.host;

const browser = await chromium.connectOverCDP(endpoint.toString());`,
    },
    responseExample: {
      language: 'text',
      code: () => `WebSocket 握手成功后进入标准 Chrome DevTools Protocol 消息流。`,
    },
    responseCodes: [
      { code: '101', description: 'WebSocket 升级成功。' },
      { code: '502', description: '已记录 active target，但 CDP 上游无法连接。' },
      { code: '503', description: '当前没有可透传的 active target。' },
    ],
    notes: [
      '先调 runtime/session，再连 WS。',
      '不要直接连接 LaunchServer 的 WebSocket 根路径；必须保留 /json/version 或 /json/list 返回的 /devtools/... 路径。',
      '/devtools/* 不属于 /api/*，不会校验 API Key，但仍只允许 localhost 访问。',
    ],
  },
]
