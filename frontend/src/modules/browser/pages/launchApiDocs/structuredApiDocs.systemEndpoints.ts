import type { StructuredApiEndpointDoc } from './structuredApiDocs.types'
import { API_AUTH_HEADER_FIELDS } from './structuredApiDocs.fields'

export const SYSTEM_API_ENDPOINT_DOCS: StructuredApiEndpointDoc[] = [
  {
    id: 'api-health-detail',
    parentId: 'api-system',
    label: '健康检查',
    method: 'GET',
    path: '/api/health',
    purpose: '确认 LaunchServer 已启动并可接收请求。',
    description: '最轻量的服务存活检查。返回 200 只代表 LaunchServer 在线，不代表浏览器实例、自动化脚本或 CDP target 已就绪。',
    fields: [...API_AUTH_HEADER_FIELDS],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl ${launchBaseUrl}/api/health \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true
}`,
    },
    responseCodes: [
      { code: '200', description: 'LaunchServer 在线。' },
      { code: '401', description: '已启用 API Key，但请求头缺失或密钥无效。' },
    ],
    notes: [
      '当前实现不会校验 HTTP 方法；调用方仍应使用 GET，避免依赖这一兼容行为。',
      '检查浏览器接管状态请继续调用 GET /api/runtime/active。',
    ],
  },
  {
    id: 'api-launch-logs-detail',
    parentId: 'api-system',
    label: '调用日志',
    method: 'GET',
    path: '/api/launch/logs',
    purpose: '读取最近的启动与运行态 API 调用记录。',
    description: '返回内存中最多 500 条记录里的最近 N 条，按新到旧排列，适合定位 selector、启动和停止失败。服务重启后记录不会保留。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'limit', type: 'integer', required: false, location: 'Query', description: '返回条数，默认 50；整数会裁剪到 1–200，非整数沿用默认值。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl "${launchBaseUrl}/api/launch/logs?limit=20" \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "items": [
    {
      "timestamp": "2026-08-08T10:30:00+08:00",
      "method": "POST",
      "path": "/api/runtime/session",
      "clientIp": "127.0.0.1",
      "code": "BUYER_001",
      "selector": { "code": "BUYER_001", "matchMode": "unique" },
      "profileId": "550e8400-e29b-41d4-a716-446655440000",
      "profileName": "buyer-001",
      "params": { "skipDefaultStartUrls": true },
      "ok": true,
      "status": 200,
      "error": "",
      "durationMs": 1240
    }
  ]
}`,
    },
    responseCodes: [
      { code: '200', description: '返回最近调用记录；没有记录时 items 为空数组。' },
      { code: '401', description: '已启用 API Key，但请求头缺失或密钥无效。' },
      { code: '405', description: '请求方法不是 GET。' },
    ],
    notes: [
      '日志只覆盖会写入 LaunchCallRecord 的启动与运行态调用，不是所有 HTTP 请求的访问日志。',
      '记录可能含 selector、实例标识和临时启动参数，接入侧应按敏感运行数据处理。',
    ],
  },
]
