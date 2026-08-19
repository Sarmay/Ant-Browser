import type { StructuredApiEndpointDoc } from './structuredApiDocs.types'

import {
  API_AUTH_HEADER_FIELDS,
  createAutomationTargetSelectorFieldDocs,
} from './structuredApiDocs.fields'

export const AUTOMATION_API_ENDPOINT_DOCS: StructuredApiEndpointDoc[] = [
  {
    id: 'api-automation-list-detail',
    parentId: 'api-automation',
    label: '脚本列表',
    method: 'GET',
    path: '/api/automation/scripts',
    purpose: '查询可执行脚本清单。',
    description: '返回脚本元数据，用于拿 scriptId、默认 selector / params 和脚本类型。',
    fields: [...API_AUTH_HEADER_FIELDS],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl ${launchBaseUrl}/api/automation/scripts \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "status": "success",
  "data": {
    "count": 1,
    "items": [
      {
        "id": "news-query-txt",
        "name": "查询新闻并写 TXT",
        "description": "查询新闻并写入文本文件",
        "type": "playwright-cdp",
        "status": "ready",
        "entryFile": "index.cjs",
        "tags": ["news", "export"],
        "selector": { "code": "BUYER_001" },
        "params": { "keyword": "OpenAI", "limit": 10 },
        "notes": "",
        "targetConfig": {
          "mode": "existing",
          "selector": { "code": "BUYER_001" }
        },
        "publicAPI": {
          "enabled": true,
          "method": "POST",
          "path": "news/query",
          "requestMode": "standard",
          "responseMode": "envelope",
          "timeoutMs": 300000,
          "requestBodyText": "",
          "responseBodyText": "",
          "variables": []
        },
        "createdAt": "2026-08-08T10:00:00+08:00",
        "updatedAt": "2026-08-08T10:00:00+08:00"
      }
    ]
  }
}`,
    },
    responseCodes: [
      { code: '200', description: '返回脚本列表。' },
      { code: '401', description: '已启用 API Key，但请求头缺失或密钥无效。' },
      { code: '405', description: '请求方法不是 GET。' },
      { code: '500', description: '读取脚本目录失败。' },
      { code: '503', description: '自动化脚本能力未启用。' },
    ],
    notes: [
      '不返回 scriptText。',
      'publicAPI.enabled=true 时，可用 publicAPI.path 拼出 /api/automation/hooks/{hookPath}。',
    ],
  },
  {
    id: 'api-automation-script-detail',
    parentId: 'api-automation',
    label: '脚本详情',
    method: 'GET',
    path: '/api/automation/scripts/{scriptId}',
    purpose: '按 scriptId 查询单个脚本详情。',
    description: '标准单资源读取接口，用于从脚本列表进入某个脚本时补充其来源和包格式等元数据。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'scriptId', type: 'string', required: true, location: 'Path', description: '脚本唯一 ID；须为单一路径段，仅允许字母、数字、点、下划线和连字符。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl ${launchBaseUrl}/api/automation/scripts/news-query-txt \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "status": "success",
  "data": {
    "item": {
      "id": "news-query-txt",
      "name": "查询新闻并写 TXT",
      "description": "查询新闻并写入文本文件",
      "type": "playwright-cdp",
      "status": "ready",
      "entryFile": "index.cjs",
      "tags": ["news", "export"],
      "selector": { "code": "BUYER_001" },
      "params": { "keyword": "OpenAI", "limit": 10 },
      "notes": "",
      "targetConfig": {
        "mode": "existing",
        "selector": { "code": "BUYER_001" }
      },
      "publicAPI": {
        "enabled": true,
        "method": "POST",
        "path": "news/query",
        "requestMode": "standard",
        "responseMode": "envelope",
        "timeoutMs": 300000,
        "requestBodyText": "",
        "responseBodyText": "",
        "variables": []
      },
      "packageFormat": "ant-automation-script",
      "manifestVersion": 1,
      "source": {
        "type": "git",
        "uri": "https://example.com/repo.git",
        "ref": "main"
      }
    }
  }
}`,
    },
    responseCodes: [
      { code: '200', description: '返回脚本详情。' },
      { code: '400', description: 'scriptId 格式非法。' },
      { code: '401', description: '已启用 API Key，但请求头缺失或密钥无效。' },
      { code: '404', description: '脚本不存在。' },
      { code: '405', description: '请求方法不是 GET。' },
      { code: '500', description: '读取脚本详情失败。' },
      { code: '503', description: '自动化脚本能力未启用。' },
    ],
    notes: [
      '不返回 scriptText。',
    ],
  },
  {
    id: 'api-automation-run-detail',
    parentId: 'api-automation',
    label: '执行脚本',
    method: 'POST',
    path: '/api/automation/scripts/run',
    purpose: '按 scriptId 执行脚本。',
    description: '外部调用方传入脚本 ID，并可用 selector、targetInput 或 params 覆盖脚本默认配置。推荐优先传 selector.code；如果脚本已在 UI 中绑定目标实例，也可以只传 scriptId 直接执行。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'scriptId', type: 'string', required: true, location: 'Body', description: '要执行的脚本 ID。' },
      { name: 'selector', type: 'object', required: false, location: 'Body', description: '覆盖脚本默认 selector。' },
      { name: 'targetInput', type: 'object', required: false, location: 'Body', description: '覆盖脚本目标策略的对象输入；提供后不会沿用脚本默认 selector。' },
      { name: 'params', type: 'object', required: false, location: 'Body', description: '覆盖脚本默认 params。' },
      { name: 'useScriptSelector', type: 'boolean', required: false, location: 'Body', description: '显式指定是否沿用脚本默认 selector。' },
      { name: 'useScriptParams', type: 'boolean', required: false, location: 'Body', description: '显式指定是否沿用脚本默认 params。' },
      { name: 'timeoutMs', type: 'integer', required: false, location: 'Body', description: '本次脚本执行超时时间，范围 1000 到 1800000。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/automation/scripts/run \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "scriptId": "news-query-txt",
    "selector": { "code": "BUYER_001" },
    "params": { "keyword": "OpenAI", "limit": 10 }
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "status": "success",
  "message": "已抓取 10 条新闻并写入 TXT",
  "data": {
    "run": {
      "id": "run-1",
      "scriptId": "news-query-txt",
      "scriptName": "查询新闻并写 TXT",
      "scriptType": "playwright-cdp",
      "status": "success",
      "summary": "已抓取 10 条新闻并写入 TXT",
      "error": "",
      "resultText": "{\\\"result\\\":{\\\"count\\\":10}}",
      "logText": "",
      "startedAt": "2026-08-08T10:00:00+08:00",
      "finishedAt": "2026-08-08T10:00:12+08:00",
      "durationMs": 12034
    },
    "summary": "已抓取 10 条新闻并写入 TXT",
    "result": { "count": 10 }
  }
}`,
    },
    responseCodes: [
      { code: '200', description: '执行成功。' },
      { code: '400', description: 'scriptId 缺失、selector / targetInput / params 不是对象、字段冲突，或 timeoutMs 超出范围。' },
      { code: '401', description: '已启用 API Key，但请求头缺失或密钥无效。' },
      { code: '405', description: '请求方法不是 POST。' },
      { code: '500', description: '脚本执行失败。' },
      { code: '503', description: '自动化脚本能力未启用。' },
    ],
    notes: [
      'selector / targetInput / params 必须是 JSON object；请求体最大 1 MiB，未知字段会被拒绝。',
      '不传时沿用脚本默认配置。',
      'useScriptSelector=true 不能同时传 selector；useScriptSelector=false 时必须传 selector。useScriptParams 同理。',
      'timeoutMs=0 表示使用默认值；非零时必须在 1000–1800000 之间。',
    ],
  },
  {
    id: 'api-automation-runs-detail',
    parentId: 'api-automation',
    label: '运行记录',
    method: 'GET',
    path: '/api/automation/scripts/runs',
    purpose: '查询最近脚本运行记录。',
    description: '返回最近 N 次脚本执行记录，适合调试、审计和任务结果回看。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'limit', type: 'integer', required: false, location: 'Query', description: '返回记录条数，默认 20，最小 1，最大 200。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl "${launchBaseUrl}/api/automation/scripts/runs?limit=20" \\
  -H "${authHeader}: <your-api-key>"`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "status": "success",
  "data": {
    "count": 1,
    "limit": 20,
    "items": [
      {
        "id": "run-1",
        "scriptId": "news-query-txt",
        "scriptName": "查询新闻并写 TXT",
        "scriptType": "playwright-cdp",
        "status": "success",
        "summary": "已抓取 10 条新闻并写入 TXT",
        "error": "",
        "resultText": "{\\\"result\\\":{\\\"count\\\":10}}",
        "logText": "",
        "startedAt": "2026-08-08T10:00:00+08:00",
        "finishedAt": "2026-08-08T10:00:12+08:00",
        "durationMs": 12034
      }
    ]
  }
}`,
    },
    responseCodes: [
      { code: '200', description: '返回运行记录。' },
      { code: '401', description: '已启用 API Key，但请求头缺失或密钥无效。' },
      { code: '405', description: '请求方法不是 GET。' },
      { code: '500', description: '读取运行记录失败。' },
      { code: '503', description: '自动化脚本能力未启用。' },
    ],
    notes: [
      'limit 非整数时沿用默认值 20；整数会裁剪到 1–200。',
      'resultText 和 logText 可能较大，接入方应按需保存。',
    ],
  },
  {
    id: 'api-automation-hook-detail',
    parentId: 'api-automation',
    label: '公开脚本 Hook',
    method: 'POST',
    path: '/api/automation/hooks/{hookPath}',
    purpose: '通过脚本配置的稳定路径直接触发自动化任务。',
    description: '脚本启用 Public API 后，会暴露一个无需 scriptId 的 POST 入口。hookPath 来自脚本详情的 publicAPI.path；路径会规范化为小写，并在全部脚本中保持唯一。',
    fields: [
      ...API_AUTH_HEADER_FIELDS,
      { name: 'hookPath', type: 'string', required: true, location: 'Path', description: '脚本 publicAPI.path，例如 image/chatgpt-generate-download。' },
      { name: 'code', type: 'string', required: false, location: 'Body', description: '兼容写法：按 launchCode 使用已有实例；不能和 instance 同时传。' },
      { name: 'instance', type: 'object', required: false, location: 'Body', description: '本次实例策略；不传时沿用脚本默认目标。' },
      { name: 'instance.type', type: 'script-default | existing | rotate | create', required: false, location: 'Body', description: '目标模式；传 instance 时必填。' },
      { name: 'instance.selector', type: 'object', required: false, location: 'Body', description: 'existing / rotate 模式必填；以下子字段至少提供一项。' },
      ...createAutomationTargetSelectorFieldDocs('instance.selector'),
      { name: 'instance.templateSelector', type: 'object', required: false, location: 'Body', description: 'create 模式必填；以下子字段至少提供一项。' },
      ...createAutomationTargetSelectorFieldDocs('instance.templateSelector'),
      { name: 'instance.createNameTemplate', type: 'string', required: false, location: 'Body', description: 'create 模式的新实例名称模板。' },
      { name: 'instance.profileName', type: 'string', required: false, location: 'Body', description: 'createNameTemplate 的兼容字段。' },
      { name: 'params', type: 'object', required: false, location: 'Body', description: '脚本参数，会与脚本默认 params 递归合并。' },
      { name: 'timeoutMs', type: 'integer', required: false, location: 'Body', description: '执行超时；不传或为 0 时继续读取 Query / 脚本配置，非零范围 1000–1800000。' },
      { name: 'timeoutMs', type: 'integer', required: false, location: 'Query', description: 'Body 未设置时使用；仅可解析的正整数会覆盖脚本配置，否则静默回退。当前 Query 值没有 1800000 上限校验，执行层仍会裁剪。' },
    ],
    requestExample: {
      language: 'bash',
      code: ({ launchBaseUrl, authHeader }) => `curl -X POST ${launchBaseUrl}/api/automation/hooks/image/chatgpt-generate-download \\
  -H "Content-Type: application/json" \\
  -H "${authHeader}: <your-api-key>" \\
  -d '{
    "instance": {
      "type": "existing",
      "selector": { "code": "BUYER_001" }
    },
    "params": {
      "prompt": "A cinematic chrome ant browser mascot"
    },
    "timeoutMs": 300000
  }'`,
    },
    responseExample: {
      language: 'json',
      code: () => `{
  "ok": true,
  "status": "success",
  "summary": "图片已生成并下载",
  "message": "图片已生成并下载",
  "data": {
    "downloadPath": "/path/to/image.png",
    "contentType": "image/png"
  },
  "result": {
    "downloadPath": "/path/to/image.png",
    "contentType": "image/png"
  }
}`,
    },
    responseCodes: [
      { code: '200', description: '脚本执行结束；脚本自身失败时也返回 200，但 ok=false。' },
      { code: '400', description: '请求体非法、实例策略不完整、字段冲突、必填变量缺失或 timeoutMs 超出范围。' },
      { code: '401', description: '已启用 API Key，但请求头缺失或密钥无效。' },
      { code: '404', description: 'hookPath 不存在，或对应脚本未启用 Public API。' },
      { code: '405', description: '请求方法不是 POST。' },
      { code: '500', description: 'Hook 查找或脚本启动发生内部错误。' },
      { code: '503', description: '自动化脚本能力未启用。' },
    ],
    notes: [
      '请求体可以为空或 null，此时使用脚本默认目标与默认参数；请求体最大 1 MiB，未知字段会被拒绝。',
      '超时优先级：Body timeoutMs > Query timeoutMs > 脚本 publicAPI.timeoutMs。',
      '脚本可用 {{name}} 或 ${name} 声明 Public API 变量；params 中的同名值优先于变量默认值。',
      '不要只依据 HTTP 状态判断脚本是否成功；始终检查响应的 ok 与 status。',
    ],
  },
]
