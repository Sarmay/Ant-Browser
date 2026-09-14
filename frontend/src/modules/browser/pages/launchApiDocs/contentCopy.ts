export const DOC_COPY_INSTANCE = `# 复制实例

从主实例复制出一个新实例，适合「模板号 -> 新号」：代理、内核、启动参数跟着走，指纹种子换新，浏览器用户数据不跟着走。

## 复制和导出不是一回事

| 动作 | 复制了什么 | 没复制什么 |
|------|------------|------------|
| 复制实例 | 内核、代理、分组、标签、关键字、启动参数、会话恢复设置 | Cookie、登录态、插件单独配置、用户数据目录 |
| 导出 / 导入 ZIP | 实例配置 + 完整用户数据目录 | 代理只按本地同名适配 |

新实例的 \`userDataDir\` 是新的 profileId，启动后是干净环境。需要保留登录态时用导出导入，不要用复制。

## 应用内复制

\`\`\`text
实例列表 -> 选中主实例 -> 复制 / 配置复制
\`\`\`

两种指纹模式：

| 模式 | 行为 |
|------|------|
| \`regular\` | 指纹参数原样保留，和源实例同一套指纹 |
| \`auto_fingerprint\`（默认） | 保留源实例指纹模板，只重生成勾选的项；默认只重生成指纹种子 |

默认勾选「指纹种子」。多号隔离时不要用 \`regular\`，否则指纹会撞车。

## 脚本里按模板新建

内置自动化脚本的 \`targetConfig.mode = "create"\` 会在执行前调用应用内的 \`Copy()\`。

\`\`\`json
{
  "targetConfig": {
    "mode": "create",
    "templateSelector": {
      "code": "MAIN_001"
    },
    "createNameTemplate": "\${templateName}-\${timestamp}"
  }
}
\`\`\`

- \`templateSelector\` 必须唯一命中主实例，推荐 \`code\` 或 \`profileId\`
- 默认名称模板是 \`\${templateName}-\${timestamp}\`
- 脚本 create 走的是 \`Copy()\`：指纹参数整份替换成全局默认指纹，不是 UI 里可勾选项的 \`auto_fingerprint\`
- 复制完成后，脚本会启动这个新实例再跑 Playwright

用 Public Hook 临时按模板新建：

\`\`\`bash
curl -X POST http://127.0.0.1:19876/api/automation/hooks/your-hook-path \\
  -H "Content-Type: application/json" \\
  -d '{
    "instance": {
      "type": "create",
      "templateSelector": { "code": "MAIN_001" },
      "createNameTemplate": "task-\${timestamp}"
    },
    "params": {
      "url": "https://example.com"
    },
    "timeoutMs": 300000
  }'
\`\`\`

\`POST /api/automation/scripts/run\` 也可以传 \`targetInput\`，覆盖脚本里保存的模板选择器或名称模板。

## HTTP 复制实例

\`\`\`text
POST /api/profiles/{profileId}/copy
\`\`\`

走应用内 \`CopyWithOptions()\`，和 UI「配置复制」同一套逻辑：拷内核、代理、分组、标签、关键字、启动参数、会话恢复设置；新实例使用新的 \`profileId\` / \`userDataDir\` / \`launchCode\`。不拷 Cookie、登录态和用户数据目录。

空请求体合法。默认 \`mode=auto_fingerprint\`，默认只重生成指纹种子。

\`\`\`bash
curl -X POST http://127.0.0.1:19876/api/profiles/SOURCE_PROFILE_ID/copy \\
  -H "Content-Type: application/json" \\
  -d '{
    "profileName": "task-001",
    "mode": "auto_fingerprint",
    "automationTargets": ["seed"],
    "autoLaunch": true,
    "start": {
      "skipDefaultStartUrls": true
    }
  }'
\`\`\`

\`\`\`json
{
  "ok": true,
  "created": true,
  "copied": true,
  "launched": true,
  "profileId": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "profileName": "task-001",
  "launchCode": "K2M9QX",
  "sourceProfileId": "SOURCE_PROFILE_ID",
  "debugReady": true,
  "cdpUrl": "http://127.0.0.1:19876"
}
\`\`\`

| 字段 | 默认 | 说明 |
|------|------|------|
| \`profileName\` | 源名称（副本）时间戳 | 新实例显示名 |
| \`mode\` | \`auto_fingerprint\` | \`regular\` 会原样保留源指纹，多号隔离不要用 |
| \`automationTargets\` | \`["seed"]\` | 仅 \`auto_fingerprint\` 生效：\`seed / identity / locale / screen / hardware / network\` |
| \`launchCode\` | 自动生成 | 4–32 位 \`A-Z 0-9 _ -\` |
| \`autoLaunch\` | \`false\` | 复制后立即启动 |
| \`start\` | 无 | 仅 \`autoLaunch=true\` 时生效 |
| \`proxyId\` | 跟随源实例 | 新实例使用代理池节点 |
| \`proxyConfig\` | 跟随源实例 | 新实例使用自定义代理；\`proxyId\` 有效时优先代理池 |

只有 launchCode 时：

\`\`\`bash
curl -X POST http://127.0.0.1:19876/api/profiles/copy \\
  -H "Content-Type: application/json" \\
  -d '{
    "selector": { "code": "MAIN_001" },
    "profileName": "task-001",
    "proxyConfig": "http://127.0.0.1:7890"
  }'
\`\`\`

HTTP copy 和脚本 \`create\` 的指纹处理不一样：

| 入口 | 指纹结果 |
|------|----------|
| HTTP copy / UI 复制 | 默认只换指纹种子，其它指纹参数跟着源实例 |
| 脚本 \`Copy()\` | 指纹参数整份替换成全局默认 |

不要用 \`GET\` 源实例再 \`POST /api/profiles\` 自己拼复制。那样容易把 \`--fingerprint=\` 原样写回去，多个号会撞指纹。

## 怎么选

| 场景 | 用法 |
|------|------|
| 应用里手动开一个新号 | 实例列表 -> 复制，默认 \`auto_fingerprint\` |
| 外部自己管实例生命周期 | \`POST /api/profiles/{profileId}/copy\` 或 \`POST /api/profiles/copy\` |
| 脚本自动「复制主实例再打开网页」 | \`targetConfig.mode = "create"\` |
| 外部编排只负责调脚本 | Hook \`instance.type = "create"\` |
| 要把登录态一起带走 | 导出 ZIP，不要复制 |
`
