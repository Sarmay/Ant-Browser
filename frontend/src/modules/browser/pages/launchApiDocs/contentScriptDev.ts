export const DOC_SCRIPT_DEV = `# 内置自动化脚本开发

Ant Browser 支持在应用内编写、调试与分发基于 **Playwright-Core (CDP)** 的自动化脚本包。

---

## 1. 脚本包目录结构

自动化脚本采用独立目录包结构（Folder-as-Package）：

\`\`\`text
my-script/
├── automation.script.json   # 必选：脚本元数据与默认参数
├── index.cjs                # 必选：脚本入口 CommonJS 文件
└── helper.cjs               # 可选：辅助模块
\`\`\`

- **演示脚本库**：\`backend/internal/automation/demo-library/\`（提交至 Git 仓库）。
- **用户脚本库**：\`data/automation/scripts/\`（本地持久化与编辑）。

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
    "type": "existing",
    "selector": {
      "code": "BUYER_001"
    }
  },
  "defaultSelector": {
    "code": "BUYER_001"
  },
  "defaultParams": {
    "keyword": "OpenAI",
    "timeoutMs": 30000
  }
}
\`\`\`

---

## 3. 脚本入口与 Runner 上下文

脚本入口必须以 CommonJS 格式导出 \`run\` 异步方法：

\`\`\`javascript
module.exports.run = async ({
  useBrowser,        // 高层一站式初始化方法
  launch,            // 底层调用 LaunchServer 启动/准备实例
  connect,           // 底层 CDP 连接 Playwright
  openPage,          // 页面导航、重用与前台激活
  grantPermissions,  // 为域名授予浏览器权限
  browserFetch,      // 页面上下文发请求（自带 Cookie 与跨域豁免）
  pageAPI,           // browserFetch 别名
  selector,          // 目标实例选择器对象
  params,            // 业务参数对象
  log,               // 结构化步骤日志输出
  artifact,          // 产物文件落盘与追踪
  chromium,          // Playwright Core 模块
}) => {
  // 业务逻辑...
};
\`\`\`

---

## 4. 核心 API 参考

### 4.1 \`useBrowser(options)\`（推荐）

一站式完成实例启动、CDP 连接、打开目标 URL 及置顶激活：

\`\`\`javascript
const runtime = await useBrowser({
  selector,                        // 实例选择器（如 { code: "BUYER_001" }）
  url: "https://www.baidu.com",     // 目标 URL（可选）
  startUrls: ["https://www.baidu.com"],
  skipDefaultStartUrls: true,      // 跳过默认首页（推荐 true）
  reuseCurrentPage: true,          // 是否重用已有活跃标签页
  bringToFront: true,              // 将标签页切至前台
  timeoutMs: 30000,
});

const { browser, context, page, session } = runtime;
\`\`\`

### 4.2 \`browserFetch(page, request)\`（站内接口请求）

在网页内部上下文发起请求，**天然携带当前会话的 Cookie / Storage 并免受 CORS 限制**：

\`\`\`javascript
const res = await browserFetch(page, {
  url: "https://api.example.com/data/list",
  method: "POST",
  json: { query: "keyword" },
  timeoutMs: 15000,
  parseJSON: true,
});

if (res.ok) {
  console.log("接口返回:", res.json);
}
\`\`\`

### 4.3 \`artifact(name)\`（产物生成）

用于将脚本运行生成的结果（TXT, JSON, PNG, CSV 等）安全落盘，并自动在前端“运行记录”中呈现：

\`\`\`javascript
const fs = require('fs');

// 获取专属落盘路径
const resultPath = artifact('result.json');
fs.writeFileSync(resultPath, JSON.stringify(data, null, 2), 'utf8');

// 截图并关联产物
const shotPath = artifact('page.png');
await page.screenshot({ path: shotPath, fullPage: true });
\`\`\`

### 4.4 \`log(key, value)\`（步骤日志）

\`\`\`javascript
log('step', 'NAVIGATE');
log('keyword', params.keyword);
log('itemCount', items.length);
\`\`\`

---

## 5. 返回值规范

\`\`\`javascript
return {
  ok: true,                     // 执行是否成功
  summary: "成功抓取 10 条数据",  // 摘要文案（显示在 UI 卡片上）
  error: "",                    // 失败时的错误信息
  title: await page.title(),
  url: page.url(),
};
\`\`\`
`
