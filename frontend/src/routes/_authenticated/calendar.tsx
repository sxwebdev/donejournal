import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import { useState } from "react"
import { startOfMonth } from "date-fns"
import { CalendarView } from "@/components/calendar/calendar-view"
import { ProjectSelector } from "@/components/projects/project-selector"

const calendarSearchSchema = z.object({
  projectId: z.string().optional(),
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
        <ProjectSelector
          value={search.projectId}
          onChange={(v) =>
            navigate({
              search: (prev) => ({ ...prev, projectId: v }),
              replace: true,
            })
          }
          placeholder="All projects"
          className="w-auto min-w-35"
        />
      </div>
      <CalendarView
        currentMonth={currentMonth}
        onMonthChange={setCurrentMonth}
        projectId={search.projectId}
      />
    </div>
  )
}
