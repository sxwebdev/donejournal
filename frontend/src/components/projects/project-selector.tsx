import { useQuery } from "@connectrpc/connect-query"
import { listProjects } from "@/api/gen/donejournal/projects/v1/projects-ProjectService_connectquery"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

type Props = {
  value?: string
  onChange: (value: string | undefined) => void
  placeholder?: string
  className?: string
}

export function ProjectSelector({
  value,
  onChange,
  placeholder = "No project",
  className,
}: Props) {
  const { data } = useQuery(listProjects, { pageSize: 100 })
  const projects = data?.projects ?? []

  return (
    <Select
      value={value ?? "__none__"}
      onValueChange={(v: string | null) => onChange(!v || v === "__none__" ? undefined : v)}
    >
      <SelectTrigger className={className}>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="__none__">{placeholder}</SelectItem>
        {projects.map((ps) => (
          <SelectItem key={ps.project?.id} value={ps.project?.id ?? ""}>
            {ps.project?.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
