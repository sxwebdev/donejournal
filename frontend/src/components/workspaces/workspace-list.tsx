import { useCallback, useRef } from "react"
import { useQuery } from "@connectrpc/connect-query"
import { create } from "@bufbuild/protobuf"
import { listWorkspaces } from "@/api/gen/donejournal/workspaces/v1/workspaces-WorkspaceService_connectquery"
import { SubscribeWorkspacesRequestSchema } from "@/api/gen/donejournal/workspaces/v1/workspaces_pb"
import { workspacesClient } from "@/api/client"
import { useSubscriptionRefetch } from "@/hooks/use-subscription-refetch"
import { WorkspaceCard } from "./workspace-card"
import { Skeleton } from "@/components/ui/skeleton"
import { FolderOpen } from "lucide-react"

type Props = {
  search?: string
  includeArchived?: boolean
}

export function WorkspaceList({ search, includeArchived }: Props) {
  const query = useQuery(listWorkspaces, {
    pageSize: 100,
    search: search || undefined,
    includeArchived: includeArchived ?? false,
  })

  const subRef = useRef<{ abort: () => void } | null>(null)
  const subscribe = useCallback(
    (signal: AbortSignal) =>
      workspacesClient.subscribeWorkspaces(
        create(SubscribeWorkspacesRequestSchema),
        { signal }
      ),
    []
  )
  useSubscriptionRefetch({ refetch: query.refetch, subscribe, ref: subRef })

  const { data, isLoading } = query

  if (isLoading) {
    return (
      <div className="space-y-2">
        {[...Array(4)].map((_, i) => (
          <Skeleton key={i} className="h-20 w-full rounded-lg" />
        ))}
      </div>
    )
  }

  const workspaces = data?.workspaces ?? []

  if (!workspaces.length) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <FolderOpen className="mb-3 h-10 w-10 text-muted-foreground/50" />
        <p className="font-medium text-muted-foreground">No workspaces found</p>
        <p className="mt-1 text-sm text-muted-foreground">
          Create one above to organize your tasks
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {workspaces.map((ws) => (
        <WorkspaceCard key={ws.workspace?.id} stats={ws} />
      ))}
    </div>
  )
}
