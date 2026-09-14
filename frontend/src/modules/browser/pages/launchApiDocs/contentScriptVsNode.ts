export const DOC_SCRIPT_VS_NODE = `# 内置脚本与外部 Node

两条路底层都是 **Node.js + Playwright-Core (CDP)**。差别不在「会不会用 Playwright」，而在谁负责复制实例、启动、连 CDP、记日志。

## 先选这条

页面操作用内置脚本。邮箱 API、批次、重试、落库用外部 Node 调 Launch API。不要自己再写一套启动和 CDP 连接。

\`\`\`text
外部 Node（编排）
  申请邮箱 / 排队 / 重试 / 写结果
  POST /api/automation/scripts/run  或  /api/automation/hooks/{path}
        ↓
Ant Browser 内置脚本（执行）
  create 复制主实例 -> useBrowser 打开网页 -> 填表 / 截图
  浏览器窗口保持打开，脚本结束只断开 CDP
\`\`\`

## 能力对照

| 能力 | 内置 \`playwright-cdp\` 脚本 | 外部 Node \`connectOverCDP\` |
|------|-----------------------------|------------------------------|
| 复制主实例 | \`targetConfig.mode = "create"\` 直接 \`Copy()\` | \`POST /api/profiles/{profileId}/copy\` |
| 指纹 / 代理 | 启动链路已带上 | 必须走 \`/api/launch\` 或 \`/api/runtime/session\`，不能自己 \`chromium.launch()\` |
| 打开页面 | \`useBrowser\` / \`openPage\` | 自己处理空白页、前台、权限 |
| 站内带 Cookie 请求 | \`browserFetch\` | 自己在 page 里 \`evaluate(fetch)\` |
| 运行记录 / 产物 | 应用内 Runs + \`artifact()\` | 自己打日志、自己存文件 |
| npm 依赖 | 脚本包不能带 \`node_modules\`，只用 Node 内置模块和注入的 Playwright | 任意 npm |
| 超时 | 默认 5 分钟，上限 30 分钟（1000–1800000ms） | 自己定 |
| 入口格式 | CommonJS \`index.cjs\` 导出 \`run\` | ESM / TS 都可以 |

## 什么时候只用内置脚本

- 流程都在浏览器里：打开网页、填表、点按钮、读页面上的邮箱或验证码
- 单次任务不超过 30 分钟
- 不需要队列、数据库、第三方 SDK
- 需要应用里直接点「运行」看日志和产物

已有示例：\`proton-mail-first-message\` 会打开 Proton Inbox，按查询条件读最新邮件并抽出验证码。

## 什么时候必须上外部 Node

- 邮箱来自 HTTP API，要轮换、记余额、失败重试
- 一次跑多个号，要并发、排队、落盘账号
- 任务比 30 分钟更长，或要跨多次脚本调用拼状态机
- 需要 TypeScript 工程、测试、CI

外部 Node 仍然不要自己启动 Chromium。正确接法：

\`\`\`javascript
import { chromium } from "playwright-core";

const res = await fetch("http://127.0.0.1:19876/api/runtime/session", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    selector: { code: "BUYER_001" },
    skipDefaultStartUrls: true,
    timeoutMs: 45000
  })
});
const data = await res.json();
if (!data.debugReady && !data.ready) {
  throw new Error("instance not ready");
}
const browser = await chromium.connectOverCDP(data.cdpUrl);
\`\`\`

Playwright 版本应对齐应用内运行时：当前是 **playwright-core 1.59.0**。稳定接管用 \`POST /api/runtime/session\`，不要自己轮询 debug 端口。

## 推荐的任务拆法

以「复制主实例 -> 打开网页 -> 要邮箱 -> 做完」为例：

1. 外部 Node 向邮箱服务申请邮箱
2. 调用脚本，\`instance.type = "create"\`，从主实例复制新号
3. 脚本用 \`useBrowser\` 打开目标页并填入邮箱
4. 外部 Node 读验证码（邮箱 API，或再调 \`proton-mail-first-message\`）
5. 再调一次脚本提交验证码，或在同一脚本里轮询（注意总超时）
6. 外部 Node 把 \`profileId\`、邮箱、代理写入自己的存储

人机验证出现时，让脚本 \`waitForSelector\` 等到主内容出现；脚本结束后实例不会被关掉，可以在桌面窗口里手工过验证。

## 不要做的事

- 不要用外部 Playwright 直接 \`launch()\` 普通 Chrome，指纹和代理都会丢
- 不要在脚本包里塞 \`node_modules\`，导入会被拒绝
- 不要把源实例的 \`--fingerprint=\` 原样 POST 到新实例
- 不要只看 HTTP 200 判断脚本成功，要看响应里的 \`ok\` 和 \`status\`
`
