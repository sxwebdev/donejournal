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
  return date.toISOString().split("T")[0]
}
