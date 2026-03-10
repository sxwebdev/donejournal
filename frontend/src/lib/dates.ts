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

export function formatDateISO(date: Date): string {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, "0")
  const d = String(date.getDate()).padStart(2, "0")
  return `${y}-${m}-${d}`
}
