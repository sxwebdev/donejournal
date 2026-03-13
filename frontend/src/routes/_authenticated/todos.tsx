import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { z } from "zod"
import { useState, useEffect, useRef } from "react"
import { format, startOfDay } from "date-fns"
import { TodoList } from "@/components/todos/todo-list"
import { TodoFilters } from "@/components/todos/todo-filters"
import { TodoDialog } from "@/components/todos/todo-dialog"
import { OverdueBanner } from "@/components/todos/overdue-banner"
import { Button } from "@/components/ui/button"
import { Plus } from "lucide-react"
import { TodoStatus } from "@/api/gen/donejournal/todos/v1/todos_pb"

const todosSearchSchema = z.object({
  statuses: z.array(z.nativeEnum(TodoStatus)).optional(),
  from: z.string().optional(),
  to: z.string().optional(),
  workspaceId: z.string().optional(),
})

export const Route = createFileRoute("/_authenticated/todos")({
  validateSearch: todosSearchSchema,
  component: TodosPage,
})

function TodosPage() {
  const [createOpen, setCreateOpen] = useState(false)
  const [showOverdue, setShowOverdue] = useState(false)
  const search = Route.useSearch()
  const navigate = useNavigate({ from: "/todos" })
  const initialFromRef = useRef(search.from)

  useEffect(() => {
    if (!initialFromRef.current) {
      navigate({
        search: (prev) => ({
          ...prev,
          from: format(startOfDay(new Date()), "yyyy-MM-dd"),
        }),
        replace: true,
      })
    }
  }, [navigate])

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Todos</h1>
          <p className="mt-1 text-sm text-muted-foreground">Track your tasks</p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          New Todo
        </Button>
      </div>
      <TodoFilters />
      <OverdueBanner
        showOverdue={showOverdue}
        onToggle={() => setShowOverdue((v) => !v)}
        workspaceId={search.workspaceId}
      />
      <TodoList
        statuses={search.statuses}
        from={search.from}
        to={search.to}
        workspaceId={search.workspaceId}
        showOverdue={showOverdue}
      />
      <TodoDialog
        mode="create"
        open={createOpen}
        onOpenChange={setCreateOpen}
      />
    </div>
  )
}
