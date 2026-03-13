import { Badge } from "@/components/ui/badge"
import { TodoPriority } from "@/api/gen/donejournal/todos/v1/todos_pb"
import { cn } from "@/lib/utils"

type Props = {
  priority: TodoPriority
  className?: string
}

const labels: Record<TodoPriority, string> = {
  [TodoPriority.UNSPECIFIED]: "",
  [TodoPriority.NONE]: "",
  [TodoPriority.LOW]: "Low",
  [TodoPriority.MEDIUM]: "Medium",
  [TodoPriority.HIGH]: "High",
  [TodoPriority.CRITICAL]: "Critical",
}

const classes: Record<TodoPriority, string> = {
  [TodoPriority.UNSPECIFIED]: "",
  [TodoPriority.NONE]: "",
  [TodoPriority.LOW]:
    "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300",
  [TodoPriority.MEDIUM]:
    "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
  [TodoPriority.HIGH]:
    "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300",
  [TodoPriority.CRITICAL]:
    "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
}

export function PriorityBadge({ priority, className }: Props) {
  if (
    priority === TodoPriority.NONE ||
    priority === TodoPriority.UNSPECIFIED
  ) {
    return null
  }

  return (
    <Badge
      variant="outline"
      className={cn(
        "border-0 text-xs font-medium",
        classes[priority],
        className
      )}
    >
      {labels[priority]}
    </Badge>
  )
}

/** Map priority to left border color class for todo cards. */
export const priorityBorderClass: Record<TodoPriority, string> = {
  [TodoPriority.UNSPECIFIED]: "",
  [TodoPriority.NONE]: "",
  [TodoPriority.LOW]: "border-l-4 border-l-blue-400",
  [TodoPriority.MEDIUM]: "border-l-4 border-l-yellow-400",
  [TodoPriority.HIGH]: "border-l-4 border-l-orange-400",
  [TodoPriority.CRITICAL]: "border-l-4 border-l-red-500",
}
