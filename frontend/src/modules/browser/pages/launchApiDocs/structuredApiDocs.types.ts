export type StructuredApiMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'WS'

export type StructuredApiSectionId =
  | 'api-system'
  | 'api-profiles-launch'
  | 'api-catalog'
  | 'api-runtime'
  | 'api-automation'

export type StructuredApiDocId =
  | StructuredApiSectionId
  | 'api-health-detail'
  | 'api-launch-logs-detail'
  | 'api-profiles-list-detail'
  | 'api-profiles-create-detail'
  | 'api-profiles-get-detail'
  | 'api-profiles-update-detail'
  | 'api-profiles-delete-detail'
  | 'api-profiles-copy-detail'
  | 'api-profiles-copy-selector-detail'
  | 'api-profiles-status-detail'
  | 'api-profiles-stop-detail'
  | 'api-profiles-restart-detail'
  | 'api-profiles-open-url-detail'
  | 'api-profiles-cookies-get-detail'
  | 'api-profiles-cookies-put-detail'
  | 'api-catalog-proxies-detail'
  | 'api-catalog-groups-detail'
  | 'api-catalog-cores-detail'
  | 'api-launch-code-detail'
  | 'api-launch-body-detail'
  | 'api-runtime-active-detail'
  | 'api-runtime-session-detail'
  | 'api-runtime-status-detail'
  | 'api-runtime-stop-detail'
  | 'api-cdp-version-detail'
  | 'api-cdp-list-detail'
  | 'api-cdp-ws-detail'
  | 'api-automation-list-detail'
  | 'api-automation-script-detail'
  | 'api-automation-run-detail'
  | 'api-automation-runs-detail'
  | 'api-automation-run-get-detail'
  | 'api-automation-hook-detail'

export interface StructuredApiExampleContext {
  launchBaseUrl: string
  authHeader: string
}

export interface StructuredApiExample {
  language: string
  code: (ctx: StructuredApiExampleContext) => string
}

export interface StructuredApiField {
  name: string
  type: string
  required: boolean
  location: 'Path' | 'Query' | 'Body' | 'Header'
  description: string
}

export interface StructuredApiResponseCode {
  code: string
  description: string
}

export interface StructuredApiSectionDoc {
  id: StructuredApiSectionId
  title: string
  intro: string
  highlights: string[]
}

export interface StructuredApiEndpointDoc {
  id: Exclude<StructuredApiDocId, StructuredApiSectionId>
  parentId: StructuredApiSectionId
  label: string
  method: StructuredApiMethod
  path: string
  purpose: string
  description: string
  fields: StructuredApiField[]
  requestExample?: StructuredApiExample
  responseExample?: StructuredApiExample
  responseCodes: StructuredApiResponseCode[]
  notes: string[]
}
