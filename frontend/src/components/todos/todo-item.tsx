import { useState } from "react"
import {
  MoreHorizontal,
  Pencil,
  Trash2,
  Circle,
  CheckCircle2,
  XCircle,
  Check,
  Calendar as CalendarIcon,
  AlertTriangle,
  Folder,
  Repeat,
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
import { PriorityBadge, priorityBorderClass } from "./priority-badge"
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
import { TodoStatus, TodoPriority } from "@/api/gen/donejournal/todos/v1/todos_pb"
import { toDate, fromDateOnly, formatRelativeDate } from "@/lib/dates"
import { Calendar } from "@/components/ui/calendar"
import { cn } from "@/lib/utils"
import { TodoTagBadges } from "./todo-tag-badges"
import { useWorkspaceName } from "@/hooks/use-workspace-name"
import { format } from "date-fns"

const STATUS_OPTIONS: { value: TodoStatus; label: string }[] = [
  { value: TodoStatus.PENDING, label: "Pending" },
  { value: TodoStatus.IN_PROGRESS, label: "In Progress" },
  { value: TodoStatus.COMPLETED, label: "Completed" },
  { value: TodoStatus.CANCELLED, label: "Cancelled" },
]

const PRIORITY_OPTIONS: { value: TodoPriority; label: string }[] = [
  { value: TodoPriority.NONE, label: "None" },
  { value: TodoPriority.LOW, label: "Low" },
  { value: TodoPriority.MEDIUM, label: "Medium" },
  { value: TodoPriority.HIGH, label: "High" },
  { value: TodoPriority.CRITICAL, label: "Critical" },
]

type Props = {
  todo: Todo
  isOverdue?: boolean
}

export function TodoItem({ todo, isOverdue }: Props) {
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [statusOpen, setStatusOpen] = useState(false)
  const [priorityOpen, setPriorityOpen] = useState(false)
  const [dateOpen, setDateOpen] = useState(false)
  const qc = useQueryClient()

  const workspaceName = useWorkspaceName(todo.workspaceId)

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
  const completedAt = toDate(todo.completedAt)

  const changeStatus = (status: TodoStatus) => {
    if (status === TodoStatus.COMPLETED) {
      completeMutation.mutate({ id: todo.id })
    } else {
      uncheckMutation.mutate({ id: todo.id, status })
    }
    setStatusOpen(false)
  }

  // Build date display for metadata row
  const dateInfo = (() => {
    if (isDone && completedAt) {
      const rel = formatRelativeDate(completedAt)
      return {
        label: `Выполнено ${rel.label.toLowerCase()}`,
        isOverdue: false,
        date: completedAt,
      }
    }
    if (plannedDate) {
      const rel = formatRelativeDate(plannedDate)
      return { label: rel.label, isOverdue: isOverdue || rel.isOverdue, date: plannedDate }
    }
    return null
  })()

  const recurrenceLabel: Record<string, string> = { daily: "Daily", weekly: "Weekly", monthly: "Monthly" }

  // Check if metadata row has content
  const hasMetadata = dateInfo || workspaceName || todo.tagIds.length > 0 || !!todo.recurrenceRule

  return (
    <>
      <div
        className={cn(
          "group flex items-start gap-3 rounded-lg border bg-card p-3 shadow-sm transition-colors hover:bg-accent/30",
          priorityBorderClass[todo.priority],
          isDone && "opacity-60"
        )}
      >
        {/* Status icon / checkbox */}
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

        {/* Content */}
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

          {/* Metadata row: date · workspace · tags */}
          {hasMetadata && (
            <div className="mt-1.5 flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground">
              {/* Date */}
              {dateInfo && !isDone && (
                <Popover open={dateOpen} onOpenChange={setDateOpen}>
                  <PopoverTrigger
                    className={cn(
                      "inline-flex cursor-pointer items-center gap-1 hover:underline",
                      dateInfo.isOverdue && "font-medium text-red-500"
                    )}
                  >
                    {dateInfo.isOverdue ? (
                      <AlertTriangle className="h-3 w-3" />
                    ) : (
                      <CalendarIcon className="h-3 w-3" />
                    )}
                    {dateInfo.label}
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

              {/* Completion info for done todos */}
              {dateInfo && isDone && (
                <span className="inline-flex items-center gap-1">
                  <CheckCircle2 className="h-3 w-3" />
                  {dateInfo.label}
                  {completedAt && (
                    <span className="text-muted-foreground/70">
                      {format(completedAt, "HH:mm")}
                    </span>
                  )}
                </span>
              )}

              {/* Separator */}
              {dateInfo && workspaceName && <span className="text-muted-foreground/50">·</span>}

              {/* Workspace badge */}
              {workspaceName && (
                <span className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5">
                  <Folder className="h-3 w-3" />
                  {workspaceName}
                </span>
              )}

              {/* Separator before recurrence */}
              {(dateInfo || workspaceName) && todo.recurrenceRule && (
                <span className="text-muted-foreground/50">·</span>
              )}

              {/* Recurrence badge */}
              {todo.recurrenceRule && (
                <span className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5">
                  <Repeat className="h-3 w-3" />
                  {recurrenceLabel[todo.recurrenceRule] ?? todo.recurrenceRule}
                </span>
              )}

              {/* Separator before tags */}
              {(dateInfo || workspaceName || todo.recurrenceRule) && todo.tagIds.length > 0 && (
                <span className="text-muted-foreground/50">·</span>
              )}

              {/* Tags */}
              <TodoTagBadges tagIds={todo.tagIds} inline />
            </div>
          )}
        </div>

        {/* Right side: badges + actions */}
        <div className="flex shrink-0 items-center gap-1.5">
          {/* Quick actions on hover */}
          <div className="flex items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
            {canComplete && (
              <button
                onClick={() => completeMutation.mutate({ id: todo.id })}
                disabled={completeMutation.isPending}
                className="inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-green-100 hover:text-green-600 dark:hover:bg-green-900/30"
                aria-label="Complete todo"
              >
                <Check className="h-4 w-4" />
              </button>
            )}
            <button
              onClick={() => setDeleteOpen(true)}
              className="inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-red-100 hover:text-red-600 dark:hover:bg-red-900/30"
              aria-label="Delete todo"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          </div>

          {/* Always visible: priority + status + menu */}
          <Popover open={priorityOpen} onOpenChange={setPriorityOpen}>
            <PopoverTrigger className="cursor-pointer">
              <PriorityBadge priority={todo.priority} />
            </PopoverTrigger>
            <PopoverContent className="w-36 p-2" align="end">
              <div className="flex flex-col gap-1">
                {PRIORITY_OPTIONS.map((opt) => (
                  <button
                    key={opt.value}
                    onClick={() => {
                      uncheckMutation.mutate({ id: todo.id, priority: opt.value })
                      setPriorityOpen(false)
                    }}
                    className={cn(
                      "flex w-full items-center justify-between rounded-md px-1.5 py-1 transition-colors hover:bg-accent",
                    )}
                  >
                    <PriorityBadge priority={opt.value} className="pointer-events-none" />
                    {opt.value === TodoPriority.NONE && <span className="text-xs text-muted-foreground">None</span>}
                    {todo.priority === opt.value && (
                      <Check className="h-3.5 w-3.5 text-muted-foreground" />
                    )}
                  </button>
                ))}
              </div>
            </PopoverContent>
          </Popover>
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
