import { useCallback, useRef, useState } from "react"
import { useQuery, useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import { create } from "@bufbuild/protobuf"
import {
  getCalendarEntries,
  updateTodo,
  listTodos,
} from "@/api/gen/donejournal/todos/v1/todos-TodoService_connectquery"
import type { CalendarDay, Todo } from "@/api/gen/donejournal/todos/v1/todos_pb"
import {
  TodoStatus,
  SubscribeTodosRequestSchema,
} from "@/api/gen/donejournal/todos/v1/todos_pb"
import { todosClient } from "@/api/client"
import { useSubscriptionRefetch } from "@/hooks/use-subscription-refetch"
import { TodoDialog } from "@/components/todos/todo-dialog"
import { fromDate, fromDateOnly, toDate, formatDateISO } from "@/lib/dates"
import { ChevronLeft, ChevronRight, Repeat } from "lucide-react"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Link } from "@tanstack/react-router"
import {
  DndContext,
  DragOverlay,
  useDraggable,
  useDroppable,
  PointerSensor,
  useSensor,
  useSensors,
  type DragStartEvent,
  type DragEndEvent,
} from "@dnd-kit/core"
import {
  eachDayOfInterval,
  startOfWeek,
  endOfWeek,
  startOfMonth,
  endOfMonth,
  isSameMonth,
  isToday,
  format,
  addWeeks,
} from "date-fns"
import { cn } from "@/lib/utils"
import { ConnectError } from "@connectrpc/connect"
import { toast } from "sonner"

type Props = {
  currentMonth: Date
  onMonthChange: (month: Date) => void
  workspaceId?: string
}

const WEEK_START = { weekStartsOn: 1 as const }
const WEEKDAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]

const statusStyle: Partial<Record<TodoStatus, string>> = {
  [TodoStatus.PENDING]:
    "bg-blue-500/10 text-blue-700 dark:text-blue-300 border-l-2 border-blue-500",
  [TodoStatus.IN_PROGRESS]:
    "bg-yellow-500/10 text-yellow-700 dark:text-yellow-300 border-l-2 border-yellow-500",
  [TodoStatus.COMPLETED]:
    "bg-green-500/10 text-green-700 dark:text-green-300 border-l-2 border-green-500 line-through opacity-60",
  [TodoStatus.CANCELLED]:
    "bg-muted text-muted-foreground border-l-2 border-muted-foreground line-through opacity-40",
}

function chunk<T>(arr: T[], size: number): T[][] {
  const result: T[][] = []
  for (let i = 0; i < arr.length; i += size) result.push(arr.slice(i, i + size))
  return result
}

function TodoRow({
  todo,
  onClick,
  dateStr,
}: {
  todo: Todo
  onClick: () => void
  dateStr: string
}) {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: `todo-${todo.id}`,
    data: { todo, sourceDate: dateStr },
    disabled: todo.isVirtual,
  })

  if (todo.isVirtual) {
    return (
      <div
        className="w-full rounded px-1 py-0.5 text-left text-xs flex items-center gap-0.5 border-l-2 border-dashed border-muted-foreground/40 bg-muted/30 text-muted-foreground/60 cursor-default select-none"
      >
        <Repeat className="h-2.5 w-2.5 shrink-0" />
        <span className="truncate">{todo.title}</span>
      </div>
    )
  }

  return (
    <button
      ref={setNodeRef}
      {...listeners}
      {...attributes}
      onClick={(e) => {
        e.stopPropagation()
        onClick()
      }}
      className={cn(
        "w-full rounded px-1 py-0.5 text-left text-xs touch-none flex items-center gap-0.5",
        statusStyle[todo.status] ?? "bg-muted text-muted-foreground",
        isDragging && "opacity-30"
      )}
    >
      {todo.recurrenceRule && <Repeat className="h-2.5 w-2.5 shrink-0" />}
      <span className="truncate">{todo.title}</span>
    </button>
  )
}

function DayCell({
  date,
  calDay,
  isCurrentMonth,
  dateStr,
}: {
  date: Date
  calDay: CalendarDay | undefined
  isCurrentMonth: boolean
  dateStr: string
}) {
  const { setNodeRef, isOver } = useDroppable({
    id: `day-${dateStr}`,
    data: { date, dateStr },
  })
  const [editTodo, setEditTodo] = useState<Todo | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  // Use local date for URL (so user sees correct date), UTC-based key is handled in CalendarView
  const linkDateStr = format(date, "yyyy-MM-dd")
  const todos = calDay?.todos ?? []
  const visibleTodos = todos.slice(0, 2)
  const remaining = todos.length - visibleTodos.length

  return (
    <>
      <div
        ref={setNodeRef}
        className={cn(
          "flex min-h-25 flex-col border-r border-b p-1 transition-colors",
          !isCurrentMonth && "opacity-40",
          isOver && "bg-primary/10 ring-2 ring-inset ring-primary/30"
        )}
        onClick={() => setCreateOpen(true)}
      >
        <div className="mb-1 flex items-center justify-center">
          <span
            className={cn(
              "flex h-6 w-6 items-center justify-center rounded-full text-xs font-medium",
              isToday(date) && "bg-primary text-primary-foreground"
            )}
          >
            {date.getDate()}
          </span>
        </div>
        <div className="flex flex-col gap-0.5 overflow-hidden">
          {visibleTodos.map((todo) => (
            <TodoRow
              key={todo.id}
              todo={todo}
              dateStr={dateStr}
              onClick={() => setEditTodo(todo)}
            />
          ))}
          {remaining > 0 && (
            <Link
              to="/todos"
              search={{ from: linkDateStr, to: linkDateStr }}
              className="px-1 text-xs text-muted-foreground hover:text-primary"
              onClick={(e) => e.stopPropagation()}
            >
              +{remaining} more
            </Link>
          )}
        </div>
      </div>

      {editTodo && (
        <TodoDialog
          mode="edit"
          todo={editTodo}
          open={true}
          onOpenChange={(open) => {
            if (!open) setEditTodo(null)
          }}
        />
      )}
      {createOpen && (
        <TodoDialog
          mode="create"
          open={true}
          onOpenChange={(open) => {
            if (!open) setCreateOpen(false)
          }}
          initialDate={date}
        />
      )}
    </>
  )
}

function WeekDayCell({
  date,
  calDay,
  dateStr,
}: {
  date: Date
  calDay: CalendarDay | undefined
  dateStr: string
}) {
  const { setNodeRef, isOver } = useDroppable({
    id: `day-${dateStr}`,
    data: { date, dateStr },
  })
  const [editTodo, setEditTodo] = useState<Todo | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const todos = calDay?.todos ?? []

  return (
    <>
      <div
        ref={setNodeRef}
        className={cn(
          "flex min-h-40 flex-col border-r p-2 transition-colors",
          isOver && "bg-primary/10 ring-2 ring-inset ring-primary/30"
        )}
        onClick={() => setCreateOpen(true)}
      >
        <div className="mb-2 flex flex-col items-center">
          <span className="text-xs font-medium text-muted-foreground">
            {format(date, "EEE")}
          </span>
          <span
            className={cn(
              "flex h-7 w-7 items-center justify-center rounded-full text-sm font-semibold",
              isToday(date) && "bg-primary text-primary-foreground"
            )}
          >
            {date.getDate()}
          </span>
        </div>
        <div className="flex flex-col gap-1 overflow-y-auto">
          {todos.map((todo) => (
            <TodoRow
              key={todo.id}
              todo={todo}
              dateStr={dateStr}
              onClick={() => setEditTodo(todo)}
            />
          ))}
          {todos.length === 0 && (
            <span className="text-center text-xs text-muted-foreground/50">—</span>
          )}
        </div>
      </div>

      {editTodo && (
        <TodoDialog
          mode="edit"
          todo={editTodo}
          open={true}
          onOpenChange={(open) => {
            if (!open) setEditTodo(null)
          }}
        />
      )}
      {createOpen && (
        <TodoDialog
          mode="create"
          open={true}
          onOpenChange={(open) => {
            if (!open) setCreateOpen(false)
          }}
          initialDate={date}
        />
      )}
    </>
  )
}

export function CalendarView({ currentMonth, onMonthChange, workspaceId }: Props) {
  const qc = useQueryClient()
  const [activeTodo, setActiveTodo] = useState<Todo | null>(null)
  const [viewMode, setViewMode] = useState<"month" | "week">("month")
  const [currentWeek, setCurrentWeek] = useState(() =>
    startOfWeek(new Date(), WEEK_START)
  )

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } })
  )

  const invalidate = () =>
    Promise.all([
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({ schema: listTodos, cardinality: "finite" }),
      }),
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({ schema: getCalendarEntries, cardinality: "finite" }),
      }),
    ])

  const moveMutation = useMutation(updateTodo, {
    onSuccess: () => { invalidate() },
    onError: (err) => {
      toast.error("Failed to move todo", {
        description: err instanceof ConnectError ? err.rawMessage : "Unknown error",
      })
      invalidate()
    },
  })

  function handleDragStart(event: DragStartEvent) {
    setActiveTodo(event.active.data.current?.todo ?? null)
  }

  function handleDragEnd(event: DragEndEvent) {
    setActiveTodo(null)
    const { active, over } = event
    if (!over) return

    const todo = active.data.current?.todo as Todo | undefined
    const sourceDate = active.data.current?.sourceDate as string | undefined
    const targetDate = over.data.current?.dateStr as string | undefined

    if (!todo || !sourceDate || !targetDate || sourceDate === targetDate) return

    const [y, m, d] = targetDate.split("-").map(Number)
    const newDate = new Date(y, m - 1, d)

    moveMutation.mutate({ id: todo.id, plannedDate: fromDateOnly(newDate) })
  }

  const start =
    viewMode === "month"
      ? startOfWeek(startOfMonth(currentMonth), WEEK_START)
      : startOfWeek(currentWeek, WEEK_START)
  const end =
    viewMode === "month"
      ? endOfWeek(endOfMonth(currentMonth), WEEK_START)
      : endOfWeek(currentWeek, WEEK_START)

  const query = useQuery(getCalendarEntries, {
    from: fromDate(start),
    to: fromDate(end),
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

  const { data } = query

  const dayMap = new Map<string, CalendarDay>()
  for (const day of data?.days ?? []) {
    const d = toDate(day.date)
    if (d) dayMap.set(formatDateISO(d), day)
  }

  const days = eachDayOfInterval({ start, end })
  const weeks = chunk(days, 7)

  const prevMonth = new Date(currentMonth.getFullYear(), currentMonth.getMonth() - 1, 1)
  const nextMonth = new Date(currentMonth.getFullYear(), currentMonth.getMonth() + 1, 1)

  const weekStart = startOfWeek(currentWeek, WEEK_START)
  const weekEnd = endOfWeek(currentWeek, WEEK_START)
  const weekLabel =
    format(weekStart, "MMM d") +
    " – " +
    (weekStart.getMonth() === weekEnd.getMonth()
      ? format(weekEnd, "d, yyyy")
      : format(weekEnd, "MMM d, yyyy"))

  return (
    <DndContext
      sensors={sensors}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className="flex flex-col overflow-hidden rounded-xl border bg-card shadow-sm">
        {/* Header */}
        <div className="flex items-center justify-between border-b px-4 py-3">
          <button
            onClick={() =>
              viewMode === "month"
                ? onMonthChange(prevMonth)
                : setCurrentWeek(addWeeks(currentWeek, -1))
            }
            className="flex h-8 w-8 items-center justify-center rounded-md transition-colors hover:bg-muted"
            aria-label="Previous"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>

          <div className="flex items-center gap-3">
            <h2 className="text-base font-semibold">
              {viewMode === "month" ? format(currentMonth, "MMMM yyyy") : weekLabel}
            </h2>
            <Tabs value={viewMode} onValueChange={(v) => setViewMode(v as "month" | "week")}>
              <TabsList className="h-7 text-xs">
                <TabsTrigger value="month" className="px-2.5 py-1 text-xs">Month</TabsTrigger>
                <TabsTrigger value="week" className="px-2.5 py-1 text-xs">Week</TabsTrigger>
              </TabsList>
            </Tabs>
          </div>

          <button
            onClick={() =>
              viewMode === "month"
                ? onMonthChange(nextMonth)
                : setCurrentWeek(addWeeks(currentWeek, 1))
            }
            className="flex h-8 w-8 items-center justify-center rounded-md transition-colors hover:bg-muted"
            aria-label="Next"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
        </div>

        {viewMode === "month" && (
          <>
            {/* Weekday labels */}
            <div className="grid grid-cols-7 border-b">
              {WEEKDAYS.map((day) => (
                <div
                  key={day}
                  className="border-r py-2 text-center text-xs font-medium text-muted-foreground last:border-r-0"
                >
                  {day}
                </div>
              ))}
            </div>

            {/* Month grid */}
            <div className="flex-1">
              {weeks.map((week, wi) => (
                <div key={wi} className="grid grid-cols-7">
                  {week.map((date, di) => {
                    const dateStr = formatDateISO(date)
                    return (
                      <div key={di} className={cn(di === 6 && "border-r-0")}>
                        <DayCell
                          date={date}
                          calDay={dayMap.get(dateStr)}
                          isCurrentMonth={isSameMonth(date, currentMonth)}
                          dateStr={dateStr}
                        />
                      </div>
                    )
                  })}
                </div>
              ))}
            </div>
          </>
        )}

        {viewMode === "week" && (
          <>
            {/* Week grid — single row of 7 days */}
            <div className="grid grid-cols-7 border-b">
              {days.map((date, di) => {
                const dateStr = formatDateISO(date)
                return (
                  <div key={di} className={cn(di === 6 && "border-r-0")}>
                    <WeekDayCell
                      date={date}
                      calDay={dayMap.get(dateStr)}
                      dateStr={dateStr}
                    />
                  </div>
                )
              })}
            </div>
          </>
        )}

        {/* Legend */}
        <div className="flex flex-wrap gap-4 border-t px-4 py-3 text-xs text-muted-foreground">
          <span className="flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-full bg-blue-500" /> Pending
          </span>
          <span className="flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-full bg-yellow-500" /> In Progress
          </span>
          <span className="flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-full bg-green-500" /> Completed
          </span>
          <span className="flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-full bg-muted-foreground" />{" "}
            Cancelled
          </span>
        </div>
      </div>

      <DragOverlay dropAnimation={null}>
        {activeTodo ? (
          <div
            className={cn(
              "w-32 rounded px-1 py-0.5 text-xs shadow-lg flex items-center gap-0.5",
              statusStyle[activeTodo.status] ?? "bg-muted text-muted-foreground"
            )}
          >
            {activeTodo.recurrenceRule && <Repeat className="h-2.5 w-2.5 shrink-0" />}
            <span className="truncate">{activeTodo.title}</span>
          </div>
        ) : null}
      </DragOverlay>
    </DndContext>
  )
}
