import { useQuery } from "@connectrpc/connect-query"
import { listTags } from "@/api/gen/donejournal/tags/v1/tags-TagService_connectquery"
import { TagBadge } from "./tag-badge"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { buttonVariants } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { Tag, Check } from "lucide-react"

type Props = {
  value: string[]
  onChange: (value: string[]) => void
  className?: string
}

export function TagFilter({ value, onChange, className }: Props) {
  const { data } = useQuery(listTags, { pageSize: 100 })
  const tags = data?.tags ?? []

  if (!tags.length) return null

  const selectedTags = tags.filter((t) => value.includes(t.id))

  const toggle = (id: string) => {
    onChange(
      value.includes(id) ? value.filter((v) => v !== id) : [...value, id]
    )
  }

  return (
    <Popover>
      <PopoverTrigger
        type="button"
        className={cn(
          buttonVariants({ variant: "outline", size: "sm" }),
          "h-7 gap-1.5 text-xs",
          value.length > 0 && "border-primary",
          className
        )}
      >
        <Tag className="h-3 w-3" />
        {selectedTags.length > 0 ? (
          <span className="flex gap-1">
            {selectedTags.map((t) => (
              <TagBadge key={t.id} name={t.name} color={t.color} />
            ))}
          </span>
        ) : (
          "Tags"
        )}
      </PopoverTrigger>
      <PopoverContent className="w-48 p-2" align="start">
        <div className="flex flex-col gap-1">
          {tags.map((tag) => (
            <button
              key={tag.id}
              onClick={() => toggle(tag.id)}
              className="flex w-full items-center justify-between rounded-sm px-2 py-1.5 text-sm transition-colors hover:bg-accent"
            >
              <span className="flex items-center gap-2">
                <span
                  className="h-3 w-3 shrink-0 rounded-full"
                  style={{ backgroundColor: tag.color }}
                />
                {tag.name}
              </span>
              {value.includes(tag.id) && (
                <Check className="h-3.5 w-3.5 text-muted-foreground" />
              )}
            </button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}
