import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import {
  DOC_GROUPS,
  findDocById,
  getDefaultDoc,
  getAdjacentDocs,
  renderDocWithLaunchContext,
} from './launchApiDocs/catalog'
import { LaunchDocsFlowPage } from './launchApiDocs/LaunchDocsFlowPage'
import { LaunchDocsLayout } from './launchApiDocs/LaunchDocsLayout'
import { LaunchDocsMarkdownContent } from './launchApiDocs/LaunchDocsMarkdownContent'
import { LaunchDocsPager } from './launchApiDocs/LaunchDocsPager'
import { LaunchDocsSidebar } from './launchApiDocs/LaunchDocsSidebar'
import { StructuredApiDocsPage } from './launchApiDocs/StructuredApiDocsPage'
import {
  getStructuredApiParentDocId,
  isStructuredApiDocId,
  isStructuredApiEndpointDocId,
  type StructuredApiDocId,
} from './launchApiDocs/structuredApiDocs'

interface LaunchApiDocsViewProps {
  launchBaseUrl: string
  authHeader: string
  standalone?: boolean
}

export function LaunchApiDocsView({
  launchBaseUrl,
  authHeader,
  standalone = false,
}: LaunchApiDocsViewProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const firstDoc = getDefaultDoc()
  const [activeId, setActiveId] = useState(firstDoc.id)

  const activeDoc = findDocById(activeId) || firstDoc
  const { previous, next } = isStructuredApiEndpointDocId(activeDoc.id) ? { previous: null, next: null } : getAdjacentDocs(activeDoc.id)
  const sidebarActiveId = isStructuredApiDocId(activeDoc.id) ? getStructuredApiParentDocId(activeDoc.id) : activeDoc.id

  const selectDoc = (id: string, syncURL: boolean) => {
    const doc = findDocById(id)
    if (!doc) {
      return false
    }

    setActiveId(doc.id)
    if (syncURL) {
      setSearchParams({ doc: doc.id })
    }
    return true
  }

  useEffect(() => {
    const requestedDoc = searchParams.get('doc')?.trim() || ''
    if (!requestedDoc || requestedDoc === activeId) {
      return
    }

    if (!selectDoc(requestedDoc, false)) {
      setSearchParams({ doc: firstDoc.id })
    }
  }, [activeId, firstDoc.id, searchParams, setSearchParams])

  const renderedContent = renderDocWithLaunchContext(activeDoc.content, launchBaseUrl, authHeader)

  return (
    <LaunchDocsLayout
      standalone={standalone}
      sidebar={(
        <LaunchDocsSidebar
          groups={DOC_GROUPS}
          activeId={sidebarActiveId}
          onSelect={(id) => {
            void selectDoc(id, true)
          }}
        />
      )}
      header={null}
      content={(
        <div className="space-y-5">
          {activeDoc.id === 'tutorial-flow'
            ? <LaunchDocsFlowPage baseUrl={launchBaseUrl} />
            : isStructuredApiDocId(activeDoc.id)
              ? (
                <StructuredApiDocsPage
                  docId={activeDoc.id as StructuredApiDocId}
                  launchBaseUrl={launchBaseUrl}
                  authHeader={authHeader}
                  onOpenDoc={(id) => {
                    void selectDoc(id, true)
                  }}
                />
              )
              : (
                <LaunchDocsMarkdownContent
                  content={renderedContent}
                  docId={activeDoc.id}
                  standalone={standalone}
                />
              )}
          <LaunchDocsPager
            previous={previous}
            next={next}
            onSelect={(id) => {
              void selectDoc(id, true)
            }}
          />
        </div>
      )}
    />
  )
}
