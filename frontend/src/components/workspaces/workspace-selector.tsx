import { useQuery } from "@connectrpc/connect-query"
import { listWorkspaces } from "@/api/gen/donejournal/workspaces/v1/workspaces-WorkspaceService_connectquery"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { cn } from "@/lib/utils"

type Props = {
  value?: string
  onChange: (value: string | undefined) => void
  placeholder?: string
  className?: string
}

export function WorkspaceSelector({
  value,
  onChange,
  placeholder = "All workspaces",
  className,
}: Props) {
  const { data } = useQuery(listWorkspaces, { pageSize: 100 })
  const workspaces = data?.workspaces ?? []

  const selectedName = value
    ? workspaces.find((ws) => ws.workspace?.id === value)?.workspace?.name
    : undefined

  return (
    <Select
      value={value ?? "__none__"}
      onValueChange={(v: string | null) =>
        onChange(!v || v === "__none__" ? undefined : v)
      }
    >
      <SelectTrigger className={cn("w-full", className)}>
        <SelectValue placeholder={placeholder}>
          {selectedName ?? placeholder}
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="__none__">{placeholder}</SelectItem>
        {workspaces.map((ws) => (
          <SelectItem key={ws.workspace?.id} value={ws.workspace?.id ?? ""}>
            {ws.workspace?.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
