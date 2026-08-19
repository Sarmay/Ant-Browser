import { useLaunchContext } from '../hooks/useLaunchContext'
import { LaunchApiDocsView } from './LaunchApiDocsView'

export function LaunchApiDocsPage() {
  const { launchBaseUrl, apiAuth } = useLaunchContext()

  return <LaunchApiDocsView launchBaseUrl={launchBaseUrl} authHeader={apiAuth.header} />
}
