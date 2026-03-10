import { createFileRoute } from "@tanstack/react-router"
import { useState } from "react"
import { startOfMonth } from "date-fns"
import { CalendarView } from "@/components/calendar/calendar-view"

export const Route = createFileRoute("/_authenticated/calendar")({
  component: CalendarPage,
})

function CalendarPage() {
  const [currentMonth, setCurrentMonth] = useState(() =>
    startOfMonth(new Date())
  )

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Calendar</h1>
        <p className="mt-1 text-sm text-muted-foreground">View todos by date</p>
      </div>
      <CalendarView
        currentMonth={currentMonth}
        onMonthChange={setCurrentMonth}
      />
    </div>
  )
}
