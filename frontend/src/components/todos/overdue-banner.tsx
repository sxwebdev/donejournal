import { useCallback, useMemo, useRef } from "react"
import { useQuery } from "@connectrpc/connect-query"
import { create } from "@bufbuild/protobuf"
import { countTodos } from "@/api/gen/donejournal/todos/v1/todos-TodoService_connectquery"
import {
  TodoStatus,
  SubscribeTodosRequestSchema,
} from "@/api/gen/donejournal/todos/v1/todos_pb"
import { todosClient } from "@/api/client"
import { useSubscriptionRefetch } from "@/hooks/use-subscription-refetch"
import { endOfDateOnly } from "@/lib/dates"
import { subDays } from "date-fns"
import { AlertTriangle } from "lucide-react"
import { Button } from "@/components/ui/button"

type Props = {
  showOverdue: boolean
  onToggle: () => void
  workspaceId?: string
}

export function OverdueBanner({ showOverdue, onToggle, workspaceId }: Props) {
  const yesterday = useMemo(() => subDays(new Date(), 1), [])

  const query = useQuery(countTodos, {
    statuses: [TodoStatus.PENDING, TodoStatus.IN_PROGRESS],
    plannedDateTo: endOfDateOnly(yesterday),
    workspaceId,
  })

  const subRef = useRef<{ abort: () => void } | null>(null)
  const subscribe = useCallback(
    (signal: AbortSignal) =>
      todosClient.subscribeTodos(create(SubscribeTodosRequestSchema), {
        signal,
      }),
    []
  )
  useSubscriptionRefetch({ refetch: query.refetch, subscribe, ref: subRef })

  const count = query.data?.count ?? 0
  if (count === 0) return null

  return (
    <div className="flex items-center gap-3 rounded-lg border border-orange-200 bg-orange-50 px-4 py-2.5 dark:border-orange-900/50 dark:bg-orange-950/30">
      <AlertTriangle className="h-4 w-4 shrink-0 text-orange-500" />
      <span className="text-sm font-medium text-orange-700 dark:text-orange-400">
        {count} overdue {count === 1 ? "task" : "tasks"}
      </span>
      <Button variant="outline" size="sm" className="ml-auto" onClick={onToggle}>
        {showOverdue ? "Hide" : "Show"}
      </Button>
    </div>
  )
}
