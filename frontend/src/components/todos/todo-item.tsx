import { useState } from "react"
import { format } from "date-fns"
import {
  MoreHorizontal,
  Pencil,
  Trash2,
  Circle,
  CheckCircle2,
  XCircle,
  Check,
} from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { StatusBadge } from "./status-badge"
import { TodoDialog } from "./todo-dialog"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  completeTodo,
  countTodos,
  deleteTodo,
  getCalendarEntries,
  listTodos,
  updateTodo,
} from "@/api/gen/donejournal/todos/v1/todos-TodoService_connectquery"
import type { Todo } from "@/api/gen/donejournal/todos/v1/todos_pb"
import { TodoStatus } from "@/api/gen/donejournal/todos/v1/todos_pb"
import { toDate, fromDateOnly } from "@/lib/dates"
import { Calendar } from "@/components/ui/calendar"
import { cn } from "@/lib/utils"

const STATUS_OPTIONS: { value: TodoStatus; label: string }[] = [
  { value: TodoStatus.PENDING, label: "Pending" },
  { value: TodoStatus.IN_PROGRESS, label: "In Progress" },
  { value: TodoStatus.COMPLETED, label: "Completed" },
  { value: TodoStatus.CANCELLED, label: "Cancelled" },
]

type Props = {
  todo: Todo
  isOverdue?: boolean
}

export function TodoItem({ todo, isOverdue }: Props) {
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [statusOpen, setStatusOpen] = useState(false)
  const [dateOpen, setDateOpen] = useState(false)
  const qc = useQueryClient()

  const invalidateTodos = () =>
    Promise.all([
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: listTodos,
          cardinality: "finite",
        }),
      }),
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: getCalendarEntries,
          cardinality: "finite",
        }),
      }),
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: countTodos,
          cardinality: "finite",
        }),
      }),
    ])

  const completeMutation = useMutation(completeTodo, {
    onSuccess: invalidateTodos,
  })
  const uncheckMutation = useMutation(updateTodo, {
    onSuccess: invalidateTodos,
  })
  const deleteMutation = useMutation(deleteTodo, { onSuccess: invalidateTodos })

  const isDone =
    todo.status === TodoStatus.COMPLETED || todo.status === TodoStatus.CANCELLED
  const canComplete =
    todo.status === TodoStatus.PENDING || todo.status === TodoStatus.IN_PROGRESS
  const canUncheck = todo.status === TodoStatus.COMPLETED

  const plannedDate = toDate(todo.plannedDate)

  const changeStatus = (status: TodoStatus) => {
    if (status === TodoStatus.COMPLETED) {
      completeMutation.mutate({ id: todo.id })
    } else {
      uncheckMutation.mutate({ id: todo.id, status })
    }
    setStatusOpen(false)
  }

  return (
    <>
      <div className="flex items-start gap-3 rounded-lg border bg-card p-3 shadow-sm transition-colors hover:bg-accent/30">
        <button
          onClick={() => {
            if (canComplete) completeMutation.mutate({ id: todo.id })
            else if (canUncheck)
              uncheckMutation.mutate({
                id: todo.id,
                status: TodoStatus.PENDING,
              })
          }}
          disabled={
            (!canComplete && !canUncheck) ||
            completeMutation.isPending ||
            uncheckMutation.isPending
          }
          className={cn(
            "mt-0.5 shrink-0 text-muted-foreground transition-colors",
            (canComplete || canUncheck) && "cursor-pointer hover:text-primary",
            !canComplete && !canUncheck && "cursor-default"
          )}
          aria-label={canUncheck ? "Uncheck todo" : "Complete todo"}
        >
          {todo.status === TodoStatus.COMPLETED ? (
            <CheckCircle2 className="h-5 w-5 text-green-500" />
          ) : todo.status === TodoStatus.CANCELLED ? (
            <XCircle className="h-5 w-5 text-red-500" />
          ) : (
            <Circle className="h-5 w-5" />
          )}
        </button>

        <div className="min-w-0 flex-1">
          <p
            className={cn(
              "text-sm leading-snug font-medium",
              isDone && "text-muted-foreground line-through"
            )}
          >
            {todo.title}
          </p>
          {todo.description && (
            <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
              {todo.description}
            </p>
          )}
          {plannedDate && (
            <Popover open={dateOpen} onOpenChange={setDateOpen}>
              <PopoverTrigger
                className={cn(
                  "mt-1 cursor-pointer text-xs hover:underline",
                  isOverdue ? "font-medium text-red-500" : "text-muted-foreground"
                )}
              >
                {format(plannedDate, "MMM d, yyyy")}
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0" align="start">
                <Calendar
                  mode="single"
                  selected={plannedDate}
                  onSelect={(date) => {
                    if (date) {
                      uncheckMutation.mutate({
                        id: todo.id,
                        plannedDate: fromDateOnly(date),
                      })
                    }
                    setDateOpen(false)
                  }}
                />
              </PopoverContent>
            </Popover>
          )}
        </div>

        <div className="flex shrink-0 items-center gap-1.5">
          <Popover open={statusOpen} onOpenChange={setStatusOpen}>
            <PopoverTrigger className="cursor-pointer">
              <StatusBadge status={todo.status} />
            </PopoverTrigger>
            <PopoverContent className="w-44 p-2" align="end">
              <div className="flex flex-col gap-1">
                {STATUS_OPTIONS.map((opt) => (
                  <button
                    key={opt.value}
                    onClick={() => changeStatus(opt.value)}
                    className={cn(
                      "flex w-full items-center justify-between rounded-md px-1.5 py-1 transition-colors hover:bg-accent",
                      todo.status === opt.value && ""
                    )}
                  >
                    <StatusBadge status={opt.value} />
                    {todo.status === opt.value && (
                      <Check className="h-3.5 w-3.5 text-muted-foreground" />
                    )}
                  </button>
                ))}
              </div>
            </PopoverContent>
          </Popover>

          <DropdownMenu>
            <DropdownMenuTrigger className="inline-flex h-7 w-7 items-center justify-center rounded-md transition-colors hover:bg-muted">
              <MoreHorizontal className="h-4 w-4" />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => setEditOpen(true)}>
                <Pencil className="mr-2 h-3.5 w-3.5" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                variant="destructive"
                onClick={() => setDeleteOpen(true)}
              >
                <Trash2 className="mr-2 h-3.5 w-3.5" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      <TodoDialog
        mode="edit"
        todo={todo}
        open={editOpen}
        onOpenChange={setEditOpen}
      />

      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete todo?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => deleteMutation.mutate({ id: todo.id })}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
