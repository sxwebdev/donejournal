import { useNavigate, useSearch } from "@tanstack/react-router"
import { format } from "date-fns"
import { CalendarIcon, X } from "lucide-react"
import { Button, buttonVariants } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Calendar } from "@/components/ui/calendar"
import { cn } from "@/lib/utils"
import { TodoStatus } from "@/api/gen/donejournal/todos/v1/todos_pb"
import { ProjectSelector } from "@/components/projects/project-selector"

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
    projectId,
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

  const setProjectId = (id: string | undefined) => {
    navigate({
      search: (prev) => ({ ...prev, projectId: id }),
    })
  }

  const hasFilters = statuses.length > 0 || !!from || !!to || !!projectId

  return (
    <div className="flex flex-wrap items-center gap-2">
      {STATUS_OPTIONS.map(({ value, label }) => (
        <Button
          key={value}
          variant={statuses.includes(value) ? "default" : "outline"}
          size="sm"
          onClick={() => toggleStatus(value)}
          className="h-7 text-xs"
        >
          {label}
        </Button>
      ))}

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

      <ProjectSelector
        value={projectId}
        onChange={setProjectId}
        placeholder="All projects"
        className="h-7 w-auto min-w-35 text-xs"
      />

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
