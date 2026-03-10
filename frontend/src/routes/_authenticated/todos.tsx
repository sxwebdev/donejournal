import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import { useState } from "react"
import { TodoList } from "@/components/todos/todo-list"
import { TodoFilters } from "@/components/todos/todo-filters"
import { TodoDialog } from "@/components/todos/todo-dialog"
import { Button } from "@/components/ui/button"
import { Plus } from "lucide-react"
import { TodoStatus } from "@/api/gen/donejournal/todos/v1/todos_pb"

const todosSearchSchema = z.object({
  statuses: z.array(z.nativeEnum(TodoStatus)).optional(),
  from: z.string().optional(),
  to: z.string().optional(),
})

export const Route = createFileRoute("/_authenticated/todos")({
  validateSearch: todosSearchSchema,
  component: TodosPage,
})

function TodosPage() {
  const [createOpen, setCreateOpen] = useState(false)
  const search = Route.useSearch()

  return (
    <div className="mx-auto max-w-2xl space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Todos</h1>
          <p className="text-sm text-muted-foreground">Track your tasks</p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          New Todo
        </Button>
      </div>

      <TodoFilters />

      <TodoList
        statuses={search.statuses}
        from={search.from}
        to={search.to}
      />

      <TodoDialog
        mode="create"
        open={createOpen}
        onOpenChange={setCreateOpen}
      />
    </div>
  )
}
