import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const repositoryRoot = resolve(scriptDir, '..', '..')
const docsDir = resolve(
  repositoryRoot,
  'frontend/src/modules/browser/pages/launchApiDocs',
)

const endpointSourceFiles = [
  'structuredApiDocs.systemEndpoints.ts',
  'structuredApiDocs.profileEndpoints.ts',
  'structuredApiDocs.runtimeEndpoints.ts',
  'structuredApiDocs.automationEndpoints.ts',
]

const docsSource = endpointSourceFiles
  .map((fileName) => readFileSync(resolve(docsDir, fileName), 'utf8'))
  .join('\n')

const documentedOperations = new Set(
  [...docsSource.matchAll(/\bmethod:\s*'(GET|POST|PUT|DELETE|WS)',\s*\n\s*path:\s*'([^']+)'/g)]
    .map((match) => `${match[1]} ${match[2]}`),
)

const routeRequirements = new Map([
  ['/api/health', ['GET /api/health']],
  ['/api/automation/scripts', ['GET /api/automation/scripts']],
  ['/api/automation/scripts/', ['GET /api/automation/scripts/{scriptId}']],
  ['/api/automation/scripts/run', ['POST /api/automation/scripts/run']],
  ['/api/automation/scripts/runs', ['GET /api/automation/scripts/runs']],
  ['/api/automation/hooks/', ['POST /api/automation/hooks/{hookPath}']],
  ['/api/profiles', ['GET /api/profiles', 'POST /api/profiles']],
  ['/api/profiles/', [
    'GET /api/profiles/{profileId}',
    'PUT /api/profiles/{profileId}',
    'DELETE /api/profiles/{profileId}',
    'GET /api/profiles/{profileId}/status',
    'POST /api/profiles/{profileId}/stop',
  ]],
  ['/api/runtime/active', ['GET /api/runtime/active']],
  ['/api/runtime/session', ['POST /api/runtime/session']],
  ['/api/runtime/status', ['POST /api/runtime/status']],
  ['/api/runtime/stop', ['POST /api/runtime/stop']],
  ['/api/launch', ['POST /api/launch']],
  ['/api/launch/logs', ['GET /api/launch/logs']],
  ['/api/launch/', ['GET /api/launch/{code}']],
  ['/', ['GET /json/version', 'GET /json/list', 'WS /devtools/...']],
])

const serverHTTPSource = readFileSync(
  resolve(repositoryRoot, 'backend/internal/launchcode/server_http.go'),
  'utf8',
)
const registeredRoutes = [
  ...serverHTTPSource.matchAll(/mux\.HandleFunc\("([^"]+)"/g),
]
  .map((match) => match[1])
  .filter((route) => route === '/' || route.startsWith('/api/'))

const undocumentedRegistrations = registeredRoutes.filter(
  (route) => !routeRequirements.has(route),
)
if (undocumentedRegistrations.length > 0) {
  throw new Error(
    `LaunchServer 新增了尚未纳入文档覆盖表的路由：${undocumentedRegistrations.join(', ')}`,
  )
}

const staleRequirements = [...routeRequirements.keys()].filter(
  (route) => !registeredRoutes.includes(route),
)
if (staleRequirements.length > 0) {
  throw new Error(
    `接口文档覆盖表包含后端未注册的路由：${staleRequirements.join(', ')}`,
  )
}

const missingOperations = [...routeRequirements.values()]
  .flat()
  .filter((operation) => !documentedOperations.has(operation))
if (missingOperations.length > 0) {
  throw new Error(
    `结构化接口文档缺少以下操作：${missingOperations.join(', ')}`,
  )
}

const documentedIds = [
  ...docsSource.matchAll(/\bid:\s*'([^']+)',\s*\n\s*parentId:/g),
].map((match) => match[1])
const duplicateIds = documentedIds.filter(
  (id, index) => documentedIds.indexOf(id) !== index,
)
if (duplicateIds.length > 0) {
  throw new Error(`结构化接口文档存在重复 ID：${[...new Set(duplicateIds)].join(', ')}`)
}

console.log(`Launch API 文档覆盖检查通过：${documentedOperations.size} 个操作`)
