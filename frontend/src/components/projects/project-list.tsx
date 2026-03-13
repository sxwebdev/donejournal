import { useCallback, useRef } from "react"
import { useQuery } from "@connectrpc/connect-query"
import { create } from "@bufbuild/protobuf"
import { listProjects } from "@/api/gen/donejournal/projects/v1/projects-ProjectService_connectquery"
import { SubscribeProjectsRequestSchema } from "@/api/gen/donejournal/projects/v1/projects_pb"
import { projectsClient } from "@/api/client"
import { useSubscriptionRefetch } from "@/hooks/use-subscription-refetch"
import { ProjectCard } from "./project-card"
import { Skeleton } from "@/components/ui/skeleton"
import { FolderOpen } from "lucide-react"

type Props = {
  search?: string
  includeArchived?: boolean
}

export function ProjectList({ search, includeArchived }: Props) {
  const query = useQuery(listProjects, {
    pageSize: 100,
    search: search || undefined,
    includeArchived: includeArchived ?? false,
  })

  const subRef = useRef<{ abort: () => void } | null>(null)
  const subscribe = useCallback(
    (signal: AbortSignal) =>
      projectsClient.subscribeProjects(
        create(SubscribeProjectsRequestSchema),
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

  const projects = data?.projects ?? []

  if (!projects.length) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <FolderOpen className="mb-3 h-10 w-10 text-muted-foreground/50" />
        <p className="font-medium text-muted-foreground">No projects found</p>
        <p className="mt-1 text-sm text-muted-foreground">
          Create one above to organize your tasks
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {projects.map((ps) => (
        <ProjectCard key={ps.project?.id} stats={ps} />
      ))}
    </div>
  )
}
