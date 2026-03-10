import { format } from "date-fns"
import { Link } from "@tanstack/react-router"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { StatusBadge } from "@/components/todos/status-badge"
import type { CalendarDay } from "@/api/gen/donejournal/todos/v1/todos_pb"

type Props = {
  day: CalendarDay
  date: Date
  children: React.ReactNode
}

export function CalendarDayPopover({ day, date, children }: Props) {
  const dateStr = format(date, "yyyy-MM-dd")

  return (
    <Popover>
      <PopoverTrigger className="w-full h-full bg-transparent border-0 p-0 cursor-pointer">{children}</PopoverTrigger>
      <PopoverContent className="w-72 p-3" align="center">
        <p className="mb-2 text-sm font-semibold">{format(date, "MMMM d, yyyy")}</p>
        <div className="space-y-1.5">
          {day.todos.slice(0, 5).map((todo) => (
            <div key={todo.id} className="flex items-center justify-between gap-2">
              <p className="min-w-0 truncate text-sm">{todo.title}</p>
              <StatusBadge status={todo.status} className="flex-shrink-0 text-[10px]" />
            </div>
          ))}
        </div>
        {day.totalCount > 5 && (
          <Link
            to="/todos"
            search={{ from: dateStr, to: dateStr }}
            className="mt-2 block text-xs text-primary hover:underline"
          >
            Show all {day.totalCount} todos →
          </Link>
        )}
        {day.totalCount <= 5 && day.totalCount > 0 && (
          <Link
            to="/todos"
            search={{ from: dateStr, to: dateStr }}
            className="mt-2 block text-xs text-muted-foreground hover:text-primary hover:underline"
          >
            View in todos →
          </Link>
        )}
      </PopoverContent>
    </Popover>
  )
}
