import { useQuery } from "@connectrpc/connect-query"
import { listWorkspaces } from "@/api/gen/donejournal/workspaces/v1/workspaces-WorkspaceService_connectquery"

export function useWorkspaceName(workspaceId: string | undefined): string | undefined {
  const { data } = useQuery(listWorkspaces, { pageSize: 100 })
  if (!workspaceId) return undefined
  return data?.workspaces.find((ws) => ws.workspace?.id === workspaceId)?.workspace?.name
}
