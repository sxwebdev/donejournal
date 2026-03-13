import { useNavigate, useSearch } from "@tanstack/react-router"
import { format } from "date-fns"
import { CalendarIcon, X, CircleDot, Check, Folder } from "lucide-react"
import { Button, buttonVariants } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Calendar } from "@/components/ui/calendar"
import { cn } from "@/lib/utils"
import { TodoStatus } from "@/api/gen/donejournal/todos/v1/todos_pb"
import { StatusBadge } from "./status-badge"
import { useQuery } from "@connectrpc/connect-query"
import { listWorkspaces } from "@/api/gen/donejournal/workspaces/v1/workspaces-WorkspaceService_connectquery"
import { TagFilter } from "@/components/tags/tag-filter"

const STATUS_OPTIONS = [
  { value: TodoStatus.PENDING, label: "Pending" },
  { value: TodoStatus.IN_PROGRESS, label: "In Progress" },
  { value: TodoStatus.COMPLETED, label: "Completed" },
  { value: TodoStatus.CANCELLED, label: "Cancelled" },
]

export function TodoFilters() {
  const navigate = useNavigate({ from: "/todos" })
  const {
    statuses = [],
    from,
    to,
    workspaceId,
    tagIds = [],
  } = useSearch({ from: "/_authenticated/todos" })

  const fromDate = from ? new Date(from) : undefined
  const toDate = to ? new Date(to) : undefined

  const toggleStatus = (status: TodoStatus) => {
    const next = statuses.includes(status)
      ? statuses.filter((s) => s !== status)
      : [...statuses, status]
    navigate({
      search: (prev) => ({ ...prev, statuses: next.length ? next : undefined }),
    })
  }

  const setFrom = (date: Date | undefined) => {
    navigate({
      search: (prev) => ({
        ...prev,
        from: date ? format(date, "yyyy-MM-dd") : undefined,
      }),
    })
  }

  const setTo = (date: Date | undefined) => {
    navigate({
      search: (prev) => ({
        ...prev,
        to: date ? format(date, "yyyy-MM-dd") : undefined,
      }),
    })
  }

  const clearAll = () => {
    navigate({ search: {} })
  }

  const setWorkspaceId = (id: string | undefined) => {
    navigate({
      search: (prev) => ({ ...prev, workspaceId: id }),
    })
  }

  const setTagIds = (ids: string[]) => {
    navigate({
      search: (prev) => ({
        ...prev,
        tagIds: ids.length ? ids : undefined,
      }),
    })
  }

  const hasFilters =
    statuses.length > 0 || !!from || !!to || !!workspaceId || tagIds.length > 0

  const workspaces = useQuery(listWorkspaces, { pageSize: 100 }).data?.workspaces ?? []
  const selectedWorkspace = workspaceId
    ? workspaces.find((ws) => ws.workspace?.id === workspaceId)?.workspace
    : undefined

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Popover>
        <PopoverTrigger
          type="button"
          className={cn(
            buttonVariants({ variant: "outline", size: "sm" }),
            "h-7 gap-1.5 text-xs",
            statuses.length > 0 && "border-primary"
          )}
        >
          <CircleDot className="h-3 w-3" />
          {statuses.length > 0 ? (
            <span className="flex gap-1">
              {statuses.map((s) => (
                <StatusBadge key={s} status={s} />
              ))}
            </span>
          ) : (
            "Status"
          )}
        </PopoverTrigger>
        <PopoverContent className="w-48 p-2" align="start">
          <div className="flex flex-col gap-1">
            {STATUS_OPTIONS.map(({ value, label }) => (
              <button
                key={value}
                onClick={() => toggleStatus(value)}
                className="flex w-full items-center justify-between rounded-sm px-2 py-1.5 text-sm transition-colors hover:bg-accent"
              >
                <StatusBadge status={value} />
                {statuses.includes(value) && (
                  <Check className="h-3.5 w-3.5 text-muted-foreground" />
                )}
              </button>
            ))}
          </div>
        </PopoverContent>
      </Popover>

      <TagFilter value={tagIds} onChange={setTagIds} />

      <Popover>
        <PopoverTrigger
          type="button"
          className={cn(
            buttonVariants({ variant: "outline", size: "sm" }),
            "h-7 gap-1.5 text-xs",
            workspaceId && "border-primary"
          )}
        >
          <Folder className="h-3 w-3" />
          {selectedWorkspace ? selectedWorkspace.name : "Workspace"}
        </PopoverTrigger>
        <PopoverContent className="w-48 p-2" align="start">
          <div className="flex flex-col gap-1">
            {workspaceId && (
              <button
                onClick={() => setWorkspaceId(undefined)}
                className="flex w-full items-center rounded-sm px-2 py-1.5 text-xs text-muted-foreground transition-colors hover:bg-accent"
              >
                All workspaces
              </button>
            )}
            {workspaces.map((ws) => (
              <button
                key={ws.workspace?.id}
                onClick={() => setWorkspaceId(ws.workspace?.id)}
                className="flex w-full items-center justify-between rounded-sm px-2 py-1.5 text-sm transition-colors hover:bg-accent"
              >
                {ws.workspace?.name}
                {workspaceId === ws.workspace?.id && (
                  <Check className="h-3.5 w-3.5 text-muted-foreground" />
                )}
              </button>
            ))}
          </div>
        </PopoverContent>
      </Popover>

      <Popover>
        <PopoverTrigger
          type="button"
          className={cn(
            buttonVariants({ variant: "outline", size: "sm" }),
            "h-7 text-xs",
            fromDate && "border-primary"
          )}
        >
          <CalendarIcon className="mr-1.5 h-3 w-3" />
          {fromDate ? format(fromDate, "MMM d") : "From"}
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0">
          <Calendar mode="single" selected={fromDate} onSelect={setFrom} />
        </PopoverContent>
      </Popover>

      <Popover>
        <PopoverTrigger
          type="button"
          className={cn(
            buttonVariants({ variant: "outline", size: "sm" }),
            "h-7 text-xs",
            toDate && "border-primary"
          )}
        >
          <CalendarIcon className="mr-1.5 h-3 w-3" />
          {toDate ? format(toDate, "MMM d") : "To"}
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0">
          <Calendar mode="single" selected={toDate} onSelect={setTo} />
        </PopoverContent>
      </Popover>

      {hasFilters && (
        <Button
          variant="ghost"
          size="sm"
          className="h-7 text-xs"
          onClick={clearAll}
        >
          <X className="mr-1 h-3 w-3" />
          Clear
        </Button>
      )}
    </div>
  )
}
