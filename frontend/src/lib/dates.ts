import { timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt"
import type { Timestamp } from "@bufbuild/protobuf/wkt"

export { timestampDate, timestampFromDate }
export type { Timestamp }

export function toDate(ts: Timestamp | undefined): Date | undefined {
  if (!ts) return undefined
  return timestampDate(ts)
}

export function fromDate(date: Date): Timestamp {
  return timestampFromDate(date)
}

// Use for date-only fields (plannedDate): sends UTC midnight of the local date,
// so the backend (which groups by UTC date) places it on the correct calendar day.
export function fromDateOnly(date: Date): Timestamp {
  const utc = new Date(
    Date.UTC(date.getFullYear(), date.getMonth(), date.getDate())
  )
  return timestampFromDate(utc)
}

// Use for range end filter: sends 23:59:59.999Z of the local date in UTC.
export function endOfDateOnly(date: Date): Timestamp {
  const utc = new Date(
    Date.UTC(
      date.getFullYear(),
      date.getMonth(),
      date.getDate(),
      23,
      59,
      59,
      999
    )
  )
  return timestampFromDate(utc)
}

export function formatDateISO(date: Date): string {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, "0")
  const d = String(date.getDate()).padStart(2, "0")
  return `${y}-${m}-${d}`
}
