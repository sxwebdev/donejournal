import { useCallback, useRef } from "react"
import { useQuery } from "@connectrpc/connect-query"
import { create } from "@bufbuild/protobuf"
import { listTodos } from "@/api/gen/donejournal/todos/v1/todos-TodoService_connectquery"
import type { Todo } from "@/api/gen/donejournal/todos/v1/todos_pb"
import {
  TodoStatus,
  SubscribeTodosRequestSchema,
} from "@/api/gen/donejournal/todos/v1/todos_pb"
import { todosClient } from "@/api/client"
import { useSubscriptionRefetch } from "@/hooks/use-subscription-refetch"
import { TodoItem } from "./todo-item"
import { Skeleton } from "@/components/ui/skeleton"
import { CheckSquare } from "lucide-react"
import { fromDateOnly, endOfDateOnly, toDate } from "@/lib/dates"
import { isToday, isTomorrow, format, startOfDay, parseISO } from "date-fns"

type Props = {
  statuses?: TodoStatus[]
  from?: string
  to?: string
}

type Group = {
  label: string
  todos: Todo[]
}

function groupTodos(todos: Todo[]): Group[] {
  const groups = new Map<string, Todo[]>()
  const noDate: Todo[] = []

  for (const todo of todos) {
    const d = toDate(todo.plannedDate)
    if (!d) {
      noDate.push(todo)
      continue
    }
    const key = format(startOfDay(d), "yyyy-MM-dd")
    if (!groups.has(key)) groups.set(key, [])
    groups.get(key)!.push(todo)
  }

  const result: Group[] = []
  const sorted = [...groups.entries()].sort(([a], [b]) => a.localeCompare(b))

  for (const [key, items] of sorted) {
    const d = parseISO(key)
    let label: string
    if (isToday(d)) label = "Today"
    else if (isTomorrow(d)) label = "Tomorrow"
    else label = format(d, "MMMM d, yyyy")
    result.push({ label, todos: items })
  }

  if (noDate.length) {
    result.push({ label: "No date", todos: noDate })
  }

  return result
}

export function TodoList({ statuses, from, to }: Props) {
  const query = useQuery(listTodos, {
    pageSize: 100,
    statuses: statuses ?? [],
    plannedDateFrom: from ? fromDateOnly(parseISO(from)) : undefined,
    plannedDateTo: to ? endOfDateOnly(parseISO(to)) : undefined,
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

  const { data, isLoading } = query

  if (isLoading) {
    return (
      <div className="space-y-2">
        {[...Array(5)].map((_, i) => (
          <Skeleton key={i} className="h-16 w-full rounded-lg" />
        ))}
      </div>
    )
  }

  const todos = data?.todos ?? []

  if (!todos.length) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <CheckSquare className="mb-3 h-10 w-10 text-muted-foreground/50" />
        <p className="font-medium text-muted-foreground">No todos found</p>
        <p className="mt-1 text-sm text-muted-foreground">
          Create one above or adjust your filters
        </p>
      </div>
    )
  }

  const groups = groupTodos(todos)

  return (
    <div className="space-y-6">
      {groups.map((group) => (
        <div key={group.label}>
          <h3 className="mb-2 text-xs font-semibold tracking-wider text-muted-foreground uppercase">
            {group.label}
          </h3>
          <div className="space-y-2">
            {group.todos.map((todo) => (
              <TodoItem key={todo.id} todo={todo} />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
