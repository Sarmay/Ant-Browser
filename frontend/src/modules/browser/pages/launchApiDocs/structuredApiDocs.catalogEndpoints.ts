import type { StructuredApiEndpointDoc } from './structuredApiDocs.types'

import { API_AUTH_HEADER_FIELDS } from './structuredApiDocs.fields'

const COMMON_API_RESPONSE_CODES: StructuredApiEndpointDoc['responseCodes'] = [
  { code: '401', description: '已启用 API Key，但请求头缺失或密钥无效。' },
  { code: '405', description: '请求方法不受支持。' },
]

export const CATALOG_API_ENDPOINT_DOCS: StructuredApiEndpointDoc[] = [
  {
    id: 'api-catalog-proxies-detail',
    parentId: 'api-catalog',
    label: '代理目录',
    method: 'GET',
    path: '/api/proxies',
    purpose: '列出代理池节点，供创建或复制实例时选择 proxyId。',
    description: '只读目录。返回代理 ID、名称、配置和分组，不执行测速或健康检测。',
    fields: [...API_AUTH_HEADER_FIELDS],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl ${launchBaseUrl}/api/proxies \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "count": 1,
  "items": [
    {
      "proxyId": "proxy-us",
      "proxyName": "US Residential",
      "proxyConfig": "http://127.0.0.1:7890",
      "groupName": "海外"
    }
  ]
}`,
    },
    responseCodes: [
      { code: '200', description: '返回代理目录。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '503', description: '代理目录当前不可用。' },
    ],
    notes: [
      '只读。不要用这个接口改代理池。',
      '创建或复制实例时优先传 proxyId；自定义代理传 proxyConfig。',
    ],
  },
  {
    id: 'api-catalog-groups-detail',
    parentId: 'api-catalog',
    label: '分组目录',
    method: 'GET',
    path: '/api/groups',
    purpose: '列出实例分组，供 selector.groupId 或创建实例时使用。',
    description: '只读目录。返回分组 ID、名称、父分组和当前实例数量。',
    fields: [...API_AUTH_HEADER_FIELDS],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl ${launchBaseUrl}/api/groups \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "count": 1,
  "items": [
    {
      "groupId": "group-sales",
      "groupName": "电商",
      "parentId": "",
      "sortOrder": 0,
      "instanceCount": 3
    }
  ]
}`,
    },
    responseCodes: [
      { code: '200', description: '返回分组目录。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '500', description: '读取分组失败。' },
      { code: '503', description: '分组目录当前不可用。' },
    ],
    notes: [
      '未配置分组时 items 为空数组。',
    ],
  },
  {
    id: 'api-catalog-cores-detail',
    parentId: 'api-catalog',
    label: '内核目录',
    method: 'GET',
    path: '/api/cores',
    purpose: '列出可用浏览器内核，供创建实例时选择 coreId。',
    description: '只读目录。返回内核 ID、名称和是否默认，不返回本机路径。',
    fields: [...API_AUTH_HEADER_FIELDS],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl ${launchBaseUrl}/api/cores \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "count": 1,
  "items": [
    {
      "coreId": "core-chrome",
      "coreName": "Chrome",
      "isDefault": true
    }
  ]
}`,
    },
    responseCodes: [
      { code: '200', description: '返回内核目录。' },
      ...COMMON_API_RESPONSE_CODES,
      { code: '503', description: '内核目录当前不可用。' },
    ],
    notes: [
      '创建实例时 coreId 留空会尝试使用默认内核。',
    ],
  },
]
