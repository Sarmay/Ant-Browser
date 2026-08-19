import { useEffect, useState } from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { LaunchApiDocsView } from './modules/browser/pages/LaunchApiDocsView'
import { ThemeProvider } from './shared/theme'
import './index.css'

interface DocsContext {
  baseUrl: string
  authHeader: string
}

const defaultDocsContext: DocsContext = {
  baseUrl: window.location.origin,
  authHeader: 'X-Ant-Api-Key',
}

function StandaloneDocsApp() {
  const [docsContext, setDocsContext] = useState(defaultDocsContext)

  useEffect(() => {
    const controller = new AbortController()
    void fetch('./context.json', {
      cache: 'no-store',
      signal: controller.signal,
    })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error(`docs context request failed: ${response.status}`)
        }
        return response.json() as Promise<Partial<DocsContext>>
      })
      .then((payload) => {
        setDocsContext({
          baseUrl: String(payload.baseUrl || defaultDocsContext.baseUrl),
          authHeader: String(payload.authHeader || defaultDocsContext.authHeader),
        })
      })
      .catch((error: unknown) => {
        if (error instanceof DOMException && error.name === 'AbortError') {
          return
        }
        console.warn('Unable to load docs context; using same-origin defaults.', error)
      })

    return () => controller.abort()
  }, [])

  return (
    <ThemeProvider>
      <LaunchApiDocsView
        launchBaseUrl={docsContext.baseUrl}
        authHeader={docsContext.authHeader}
        standalone
      />
    </ThemeProvider>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <BrowserRouter>
    <StandaloneDocsApp />
  </BrowserRouter>,
)
