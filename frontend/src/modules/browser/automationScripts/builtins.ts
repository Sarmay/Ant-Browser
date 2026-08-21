import {
  AUTOMATION_SCRIPT_MANIFEST_VERSION,
  AUTOMATION_SCRIPT_PACKAGE_FORMAT,
  DUAL_INSTANCE_RUNTIME_SCRIPT_ID,
  type AutomationScriptRecord,
} from "./definitions";
import { createAutomationScriptPublicAPIConfig } from "./publicApi";
import {
  normalizeAutomationScriptTargetConfig,
} from "./targets";

const BACKEND_BUILTIN_SCRIPT_PLACEHOLDER = `module.exports.run = async () => {
  throw new Error('内置脚本源码由后端 demo-library 提供，请在桌面应用后端环境中加载或从脚本包导入。')
}`;


function nowIso(): string {
  return new Date().toISOString();
}

export {
  buildParamsTemplate,
  buildScriptTemplate,
  buildSelectorTemplate,
  buildNotesTemplate,
  normalizeDualInstanceRuntimeParamsText,
} from "./builtinsTemplates";
import {
  buildDualInstanceRuntimeParamsText,
  buildDualInstanceRuntimeScriptText,
} from "./builtinsTemplates";

export function createNewsTxtScriptDraft(): AutomationScriptRecord {
  const createdAt = nowIso();

  return {
    packageFormat: AUTOMATION_SCRIPT_PACKAGE_FORMAT,
    manifestVersion: AUTOMATION_SCRIPT_MANIFEST_VERSION,
    id: "news-query-txt",
    name: "查询新闻并写 TXT",
    description: "通过 Bing 搜索新闻关键词，提取结果并写入本地 txt 文件。",
    type: "playwright-cdp",
    status: "ready",
    entryFile: "index.cjs",
    tags: ["Playwright", "新闻", "TXT"],
    selectorText: "",
    paramsText: `{
  "keyword": "OpenAI",
  "limit": 10,
  "timeRange": "week",
  "outputFileName": "openai-news.txt",
  "timeoutMs": 30000,
  "waitAfterLoadMs": 1500,
  "captureScreenshot": false
}`,
    scriptText: BACKEND_BUILTIN_SCRIPT_PLACEHOLDER,
    notes:
      "脚本会优先使用 Bing 搜索真实新闻结果，并自动追加时间过滤、排除问答/聚合站点、回退查询词和质量校验；只有达到新闻质量门槛时才会判定成功，并把结果写入本地 txt。执行成功后可在结果里的 outputPath 找到文件。",
    targetConfig: normalizeAutomationScriptTargetConfig(null),
    publicAPI: createAutomationScriptPublicAPIConfig(),
    source: {
      type: "builtin",
      uri: "repo://backend/internal/automation/demo-library/news-query-txt",
      ref: "HEAD",
      path: "news-query-txt",
      importedAt: "",
    },
    createdAt,
    updatedAt: createdAt,
  };
}

export function createDualInstanceRuntimeScriptDraft(): AutomationScriptRecord {
  const createdAt = nowIso();

  return {
    packageFormat: AUTOMATION_SCRIPT_PACKAGE_FORMAT,
    manifestVersion: AUTOMATION_SCRIPT_MANIFEST_VERSION,
    id: DUAL_INSTANCE_RUNTIME_SCRIPT_ID,
    name: "双实例启动与 Runtime 切换",
    description:
      "通过 Launch API 分别启动两个实例，切换 Runtime 会话后交给 OpenClaw 执行。",
    type: "launch-api",
    status: "ready",
    entryFile: "index.cjs",
    tags: ["Launch API", "OpenClaw", "双实例"],
    selectorText: "",
    paramsText: buildDualInstanceRuntimeParamsText(),
    scriptText: buildDualInstanceRuntimeScriptText(),
    notes:
      "先通过接口启动两个实例并切换 Runtime 会话；随后把实例信息交给 OpenClaw 执行自动化动作。",
    targetConfig: normalizeAutomationScriptTargetConfig(null),
    publicAPI: createAutomationScriptPublicAPIConfig(),
    source: {
      type: "builtin",
      uri: "repo://backend/internal/automation/demo-library/dual-instance-runtime-switch",
      ref: "HEAD",
      path: "dual-instance-runtime-switch",
      importedAt: "",
    },
    createdAt,
    updatedAt: createdAt,
  };
}

export function createWebImageGenerateDownloadScriptDraft(): AutomationScriptRecord {
  const createdAt = nowIso();

  return {
    packageFormat: AUTOMATION_SCRIPT_PACKAGE_FORMAT,
    manifestVersion: AUTOMATION_SCRIPT_MANIFEST_VERSION,
    id: "web-image-generate-download",
    name: "网页图片生成并下载",
    description:
      "打开 ChatGPT，发送图片生成消息，等待图片生成后下载图片。",
    type: "playwright-cdp",
    status: "draft",
    entryFile: "index.cjs",
    tags: ["Playwright", "图片生成", "下载"],
    selectorText: "",
    paramsText: `{
  "prompt": "A cinematic chrome ant browser mascot, premium product lighting"
}`,
    scriptText: BACKEND_BUILTIN_SCRIPT_PLACEHOLDER,
    notes:
      "脚本默认打开 ChatGPT，输入图片生成提示词并发送；页面选择器、下载文件名等由脚本内部默认值处理，公开接口只需要传实例、提示词和超时时间。",
    targetConfig: normalizeAutomationScriptTargetConfig(null),
    publicAPI: {
      ...createAutomationScriptPublicAPIConfig(),
      enabled: true,
      path: "image/chatgpt-generate-download",
      timeoutMs: 300000,
      requestBodyText: `{
  "instance": {
    "type": "existing",
    "selector": {
      "code": "BUYER_001"
    }
  },
  "params": {
    "prompt": "{{prompt}}"
  },
  "timeoutMs": 300000
}`,
      responseBodyText: `{
  "ok": true,
  "status": "completed",
  "summary": "图片已生成并下载。",
  "outputPath": "\${artifactsDir}/generated-image.png",
  "downloadAddress": "\${artifactsDir}/generated-image.png"
}`,
      variables: [
        {
          name: "prompt",
          defaultValue:
            "A cinematic chrome ant browser mascot, premium product lighting",
          description: "发送到 ChatGPT 的图片生成提示词。",
          required: true,
        },
      ],
    },
    source: {
      type: "builtin",
      uri: "repo://backend/internal/automation/demo-library/web-image-generate-download",
      ref: "HEAD",
      path: "web-image-generate-download",
      importedAt: "",
    },
    createdAt,
    updatedAt: createdAt,
  };
}

export function createProtonMailFirstMessageScriptDraft(): AutomationScriptRecord {
  const createdAt = nowIso();

  return {
    packageFormat: AUTOMATION_SCRIPT_PACKAGE_FORMAT,
    manifestVersion: AUTOMATION_SCRIPT_MANIFEST_VERSION,
    id: "proton-mail-first-message",
    name: "Proton 邮件搜索并读取最新邮件",
    description:
      "打开 Proton Inbox，用查询语句搜索邮件，再用发件邮箱和收件人约束结果，返回命中邮件的正文内容和验证码。",
    type: "playwright-cdp",
    status: "ready",
    entryFile: "index.cjs",
    tags: ["Playwright", "邮箱", "Proton", "邮件处理"],
    selectorText: "",
    paramsText: `{
  "recipient": "",
  "searchQuery": "OpenAI, ChatGPT",
  "senderEmail": ""
}`,
    scriptText: BACKEND_BUILTIN_SCRIPT_PLACEHOLDER,
    notes:
      "searchQuery 只写入 Proton 搜索框；senderEmail 作为高级搜索约束，recipient 留空也可直接跑。若同一邮箱里 OpenAI / ChatGPT 邮件较多，再补 recipient 缩小范围。旧字段 recipientQuery 仍兼容。",
    targetConfig: normalizeAutomationScriptTargetConfig(null),
    publicAPI: {
      ...createAutomationScriptPublicAPIConfig(),
      enabled: true,
      path: "mail/proton-first-message",
      requestMode: "params-only",
      responseMode: "envelope",
      timeoutMs: 120000,
    },
    source: {
      type: "builtin",
      uri: "repo://backend/internal/automation/demo-library/proton-mail-first-message",
      ref: "HEAD",
      path: "proton-mail-first-message",
      importedAt: "",
    },
    createdAt,
    updatedAt: createdAt,
  };
}

export function createLianjiaWHHomeStep1ScriptDraft(): AutomationScriptRecord {
  const createdAt = nowIso();

  return {
    packageFormat: AUTOMATION_SCRIPT_PACKAGE_FORMAT,
    manifestVersion: AUTOMATION_SCRIPT_MANIFEST_VERSION,
    id: "lianjia-wh-home-step1",
    name: "链家武汉首页 S1 打开页面",
    description:
      "S1：启动目标浏览器实例并进入 https://wh.lianjia.com/，检测是否出现验证页，并导出当前链家 Cookie 供后续爬取链路使用。",
    type: "playwright-cdp",
    status: "ready",
    entryFile: "index.cjs",
    tags: ["Playwright", "链家", "武汉", "S1"],
    selectorText: "",
    paramsText: `{
  "targetUrl": "https://wh.lianjia.com/",
  "timeoutMs": 60000,
  "waitAfterLoadMs": 2500,
  "captureScreenshot": true,
  "keepOpen": true
}`,
    scriptText: BACKEND_BUILTIN_SCRIPT_PLACEHOLDER,
    notes:
      "脚本只做 S1 进入页面和状态采集，不绕过验证码，不抓取成交数据。若检测到 CAPTCHA，请在打开的浏览器实例里人工完成验证，再执行后续脚本或复制导出的 Cookie。",
    targetConfig: normalizeAutomationScriptTargetConfig(null),
    publicAPI: createAutomationScriptPublicAPIConfig(),
    source: {
      type: "builtin",
      uri: "repo://backend/internal/automation/demo-library/lianjia-wh-home-step1",
      ref: "HEAD",
      path: "lianjia-wh-home-step1",
      importedAt: "",
    },
    createdAt,
    updatedAt: createdAt,
  };
}

export function createLianjiaWHCookiePrepareScriptDraft(): AutomationScriptRecord {
  const createdAt = nowIso();

  return {
    packageFormat: AUTOMATION_SCRIPT_PACKAGE_FORMAT,
    manifestVersion: AUTOMATION_SCRIPT_MANIFEST_VERSION,
    id: "lianjia-wh-cookie-prepare",
    name: "链家武汉 Cookie 准备",
    description:
      "打开链家武汉首页，检测验证状态并导出当前 Cookie 供后续链家成交脚本使用。",
    type: "playwright-cdp",
    status: "ready",
    entryFile: "index.cjs",
    tags: ["Playwright", "链家", "武汉", "Cookie"],
    selectorText: "",
    paramsText: `{
  "targetUrl": "https://wh.lianjia.com/",
  "timeoutMs": 60000,
  "waitAfterLoadMs": 2500,
  "captureScreenshot": true,
  "keepOpen": true
}`,
    scriptText: BACKEND_BUILTIN_SCRIPT_PLACEHOLDER,
    notes:
      "脚本只做 Cookie 准备，不绕过验证码，不抓取成交数据。若检测到 CAPTCHA，请在打开的浏览器实例里人工完成验证，再执行后续脚本或复制导出的 Cookie。",
    targetConfig: normalizeAutomationScriptTargetConfig(null),
    publicAPI: createAutomationScriptPublicAPIConfig(),
    source: {
      type: "builtin",
      uri: "repo://backend/internal/automation/demo-library/lianjia-wh-cookie-prepare",
      ref: "HEAD",
      path: "lianjia-wh-cookie-prepare",
      importedAt: "",
    },
    createdAt,
    updatedAt: createdAt,
  };
}

export function createBeikeHousePriceExtractScriptDraft(): AutomationScriptRecord {
  const createdAt = nowIso();

  return {
    packageFormat: AUTOMATION_SCRIPT_PACKAGE_FORMAT,
    manifestVersion: AUTOMATION_SCRIPT_MANIFEST_VERSION,
    id: "beike-house-price-extract",
    name: "贝壳二手房价格提取",
    description:
      "打开贝壳二手房详情页，提取页面真实 DOM 中的总价和单价，并导出结构化记录。",
    type: "playwright-cdp",
    status: "ready",
    entryFile: "index.cjs",
    tags: ["Playwright", "贝壳", "二手房", "价格提取"],
    selectorText: "",
    paramsText: `{
  "targetUrl": "https://wx.ke.com/ershoufang/103147107668.html",
  "timeoutMs": 60000,
  "waitAfterLoadMs": 2500,
  "captureScreenshot": true,
  "keepOpen": true
}`,
    scriptText: BACKEND_BUILTIN_SCRIPT_PLACEHOLDER,
    notes:
      "脚本只读取页面真实内容，不填充 mock 或假设值。若未提取到总价或单价，会在 missingFields 中明确记录缺失字段。",
    targetConfig: normalizeAutomationScriptTargetConfig(null),
    publicAPI: createAutomationScriptPublicAPIConfig(),
    source: {
      type: "builtin",
      uri: "repo://backend/internal/automation/demo-library/beike-house-price-extract",
      ref: "HEAD",
      path: "beike-house-price-extract",
      importedAt: "",
    },
    createdAt,
    updatedAt: createdAt,
  };
}

export function buildDefaultAutomationScripts(): AutomationScriptRecord[] {
  return [
    createDualInstanceRuntimeScriptDraft(),
    createNewsTxtScriptDraft(),
    createProtonMailFirstMessageScriptDraft(),
    createWebImageGenerateDownloadScriptDraft(),
    createLianjiaWHHomeStep1ScriptDraft(),
    createLianjiaWHCookiePrepareScriptDraft(),
    createBeikeHousePriceExtractScriptDraft(),
  ];
}
