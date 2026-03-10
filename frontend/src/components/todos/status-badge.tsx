import { Badge } from "@/components/ui/badge"
import { TodoStatus } from "@/api/gen/donejournal/todos/v1/todos_pb"
import { cn } from "@/lib/utils"

type Props = {
  status: TodoStatus
  className?: string
}

const labels: Record<TodoStatus, string> = {
  [TodoStatus.UNSPECIFIED]: "Unknown",
  [TodoStatus.PENDING]: "Pending",
  [TodoStatus.IN_PROGRESS]: "In Progress",
  [TodoStatus.COMPLETED]: "Completed",
  [TodoStatus.CANCELLED]: "Cancelled",
}

const classes: Record<TodoStatus, string> = {
  [TodoStatus.UNSPECIFIED]: "bg-muted text-muted-foreground",
  [TodoStatus.PENDING]: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300",
  [TodoStatus.IN_PROGRESS]: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
  [TodoStatus.COMPLETED]: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
  [TodoStatus.CANCELLED]: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
}

export function StatusBadge({ status, className }: Props) {
  return (
    <Badge variant="outline" className={cn("border-0 text-xs font-medium", classes[status], className)}>
      {labels[status]}
    </Badge>
  )
}
