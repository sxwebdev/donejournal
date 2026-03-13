import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import { useState } from "react"
import { startOfMonth } from "date-fns"
import { CalendarView } from "@/components/calendar/calendar-view"
import { WorkspaceSelector } from "@/components/workspaces/workspace-selector"

const calendarSearchSchema = z.object({
  workspaceId: z.string().optional(),
})

export const Route = createFileRoute("/_authenticated/calendar")({
  validateSearch: calendarSearchSchema,
  component: CalendarPage,
})

function CalendarPage() {
  const [currentMonth, setCurrentMonth] = useState(() =>
    startOfMonth(new Date())
  )
  const search = Route.useSearch()
  const navigate = Route.useNavigate()

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Calendar</h1>
          <p className="mt-1 text-sm text-muted-foreground">View todos by date</p>
        </div>
        <WorkspaceSelector
          value={search.workspaceId}
          onChange={(v) =>
            navigate({
              search: (prev) => ({ ...prev, workspaceId: v }),
              replace: true,
            })
          }
          placeholder="All workspaces"
          className="w-auto min-w-35"
        />
      </div>
      <CalendarView
        currentMonth={currentMonth}
        onMonthChange={setCurrentMonth}
        workspaceId={search.workspaceId}
      />
    </div>
  )
}
