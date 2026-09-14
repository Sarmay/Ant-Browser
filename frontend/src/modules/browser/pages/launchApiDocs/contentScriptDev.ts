export const DOC_SCRIPT_DEV = `# 内置自动化脚本开发

Ant Browser 支持在应用内编写、调试与分发基于 **Playwright-Core (CDP)** 的自动化脚本包。运行时是 Node.js 22 + playwright-core 1.59.0，由应用注入 \`useBrowser\` / \`launch\` / \`connect\`，脚本自己不要启动 Chromium。

---

## 1. 脚本包目录结构

\`\`\`text
my-script/
├── automation.script.json   # 必选：脚本元数据与默认参数
├── index.cjs                # 必选：脚本入口 CommonJS 文件
└── helper.cjs               # 可选：辅助模块
\`\`\`

- **演示脚本库**：\`backend/internal/automation/demo-library/\`（随仓库提交）。
- **用户脚本库**：\`data/automation/scripts/\`（本机运行与编辑）。
- 脚本包不能包含 \`node_modules\`，导入会被拒绝。只能 \`require\` 相对路径模块和 Node 内置模块。

---

## 2. 配置文件：automation.script.json

\`\`\`json
{
  "id": "my-script",
  "name": "示例抓取脚本",
  "description": "打开网页抓取数据并保存为产物文件",
  "type": "playwright-cdp",
  "entryFile": "index.cjs",
  "targetConfig": {
    "mode": "existing",
    "selector": {
      "code": "BUYER_001"
    }
  },
  "defaultParams": {
    "keyword": "OpenAI",
    "timeoutMs": 30000
  }
}
\`\`\`

\`targetConfig.mode\` 决定执行前怎么选实例：

| mode | 行为 |
|------|------|
| \`manual\` / 不填 | 调用方传入 selector，或沿用脚本里保存的 selector |
| \`existing\` | 必须唯一命中一个已有实例 |
| \`create\` | 按 \`templateSelector\` 复制主实例，再启动新实例 |
| \`rotate\` | 按条件命中一组实例，轮流挑一个 |

\`create\` 的模板选择器和名称模板：

\`\`\`json
{
  "targetConfig": {
    "mode": "create",
    "templateSelector": { "code": "MAIN_001" },
    "createNameTemplate": "\${templateName}-\${timestamp}"
  }
}
\`\`\`

默认名称模板是 \`\${templateName}-\${timestamp}\`。\`create\` 走应用内 \`Copy()\`：拷配置、换新用户数据目录、指纹改用全局默认值。只复制实例、不跑脚本时用 \`POST /api/profiles/{profileId}/copy\`。详情见「复制实例」。

---

## 3. 脚本入口与 Runner 上下文

入口必须是 CommonJS，导出 \`run\`：

\`\`\`javascript
module.exports.run = async ({
  useBrowser,        // 一站式：启动实例 + CDP + 打开 URL
  launch,            // 调 LaunchServer 启动 / 准备实例
  connect,           // 用 Playwright 连接 CDP
  openPage,          // 导航、复用标签、置顶
  grantPermissions,  // 给域名授权
  browserFetch,      // 页面上下文发请求（带 Cookie，不受 CORS 限制）
  pageAPI,           // browserFetch 别名
  selector,          // 本次目标实例
  params,            // 业务参数
  log,               // 结构化步骤日志
  artifact,          // 产物落盘路径
  artifactsDir,      // 产物目录
  chromium,          // playwright-core
}) => {
  // 业务逻辑...
};
\`\`\`

\`launch-api\` 类型脚本拿不到 Playwright 对象，只能用 \`baseUrl\` / \`apiKey\` 调 Launch API，例如内置的双实例切换脚本。

---

## 4. 核心 API

### 4.1 \`useBrowser(options)\`（推荐）

内部顺序：启动实例 -> 连 CDP -> 打开 URL -> 授权。

\`\`\`javascript
const runtime = await useBrowser({
  selector,                        // 如 { code: "BUYER_001" }
  url: "https://www.example.com",
  startUrls: ["https://www.example.com"],
  skipDefaultStartUrls: true,
  reuseCurrentPage: true,
  bringToFront: true,
  waitUntil: "domcontentloaded",   // load | domcontentloaded | networkidle | commit
  timeoutMs: 30000,
  permissions: ["clipboard-read"],
});

const { browser, context, page, session } = runtime;
\`\`\`

\`session\` 里有 \`profileId\`、\`launchCode\`、\`cdpUrl\`、\`debugPort\`、\`debugReady\`。

### 4.2 \`launch\` + \`connect\`

需要同时开多个实例时，不要用 \`useBrowser\` 一把梭：

\`\`\`javascript
const session = await launch({
  selector: { code: "BUYER_001" },
  skipDefaultStartUrls: true,
  startUrls: ["https://example.com"],
});
const connection = await connect(session, { timeoutMs: 30000 });
\`\`\`

### 4.3 \`openPage(connection, options)\`

\`\`\`javascript
const opened = await openPage(connection, {
  url: "https://mail.proton.me/",
  reuseCurrentPage: true,
  bringToFront: true,
  waitUntil: "domcontentloaded",
  timeoutMs: 45000,
  permissions: ["notifications"],
  permissionOrigin: "https://mail.proton.me",
});
const page = opened.page;
\`\`\`

当前页是 \`about:blank\` 或已经是目标 URL 时，默认会复用标签，避免多开空白页。

### 4.4 \`browserFetch(page, request)\`

在页面上下文发请求，带当前站点 Cookie / Storage，也不吃浏览器 CORS：

\`\`\`javascript
const res = await browserFetch(page, {
  url: "https://api.example.com/data/list",
  method: "POST",
  json: { query: "keyword" },
  credentials: "include",
  timeoutMs: 15000,
  parseJSON: true,
  throwOnError: false,
});
\`\`\`

### 4.5 \`artifact(name)\` 与 \`log\`

\`\`\`javascript
const fs = require("fs");
const resultPath = artifact("result.json");
fs.writeFileSync(resultPath, JSON.stringify(data, null, 2), "utf8");
await page.screenshot({ path: artifact("page.png"), fullPage: true });

log("step", "NAVIGATE");
log("keyword", params.keyword);
\`\`\`

产物会出现在应用「运行记录」里，目录默认是 \`data/automation/artifacts/<scriptId>/\`。

---

## 5. 返回值

\`\`\`javascript
return {
  ok: true,
  summary: "成功抓取 10 条数据",
  error: "",
  title: await page.title(),
  url: page.url(),
  itemCount: 10,
};
\`\`\`

\`ok: false\` 会把这次运行标成失败。不需要 \`browser.close()\`：runner 结束时只断开 CDP，不关 Ant Browser 实例。

---

## 6. 超时

| 来源 | 默认 | 范围 |
|------|------|------|
| \`POST /api/automation/scripts/run\` 的 \`timeoutMs\` | 5 分钟 | 1000–1800000（30 分钟） |
| Public Hook | 脚本 \`publicAPI.timeoutMs\`，常见 300000 | Body > Query > 脚本配置 |
| 脚本内部 \`page.goto\` / \`useBrowser\` | 你自己传的 \`timeoutMs\` | 须小于整次运行超时 |

等邮箱、等人机验证时，调用方要把 \`timeoutMs\` 提到足够大，最大也只有 30 分钟。更长的流程拆成多次脚本调用，由外部 Node 串起来。

---

## 7. 拟人化与拦截

直接 \`fill()\` 容易被部分站点识别，推荐逐字输入：

\`\`\`javascript
async function humanType(page, selector, text, delayMs = 50) {
  const locator = page.locator(selector).first();
  await locator.click();
  await page.keyboard.press(process.platform === "darwin" ? "Meta+A" : "Control+A");
  await page.keyboard.press("Backspace");
  for (const char of text) {
    await page.keyboard.type(char);
    await page.waitForTimeout(Math.max(10, delayMs + Math.floor(Math.random() * 40) - 20));
  }
}
\`\`\`

动态页面用可见元素轮询；碰到验证码就 \`log\` 提示，并 \`waitForSelector\` 等人工完成。脚本结束后窗口还在。

---

## 8. 外部怎么调

\`\`\`bash
curl -X POST http://127.0.0.1:19876/api/automation/scripts/run \\
  -H "Content-Type: application/json" \\
  -d '{
    "scriptId": "my-script",
    "selector": { "code": "BUYER_001" },
    "params": { "keyword": "OpenAI" },
    "timeoutMs": 300000
  }'
\`\`\`

脚本已绑定 \`existing\` / \`create\` 时，可以只传 \`scriptId\`。覆盖 create 模板用 \`targetInput\`。启用了 Public API 时走 \`POST /api/automation/hooks/{path}\`。

调试看应用内运行记录，或：

\`\`\`bash
curl http://127.0.0.1:19876/api/automation/scripts/runs?limit=10
\`\`\`
`
