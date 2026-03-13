import { useQuery } from "@connectrpc/connect-query"
import { listTags } from "@/api/gen/donejournal/tags/v1/tags-TagService_connectquery"
import { TagBadge } from "@/components/tags/tag-badge"

type Props = {
  tagIds: string[]
  /** When true, renders badges without wrapper div (for inline use in metadata row) */
  inline?: boolean
}

const MAX_VISIBLE = 3

export function TodoTagBadges({ tagIds, inline }: Props) {
  const { data } = useQuery(listTags, { pageSize: 100 })

  if (!tagIds.length) return null

  const tags = data?.tags ?? []
  const matched = tags.filter((t) => tagIds.includes(t.id))

  if (!matched.length) return null

  const visible = matched.slice(0, MAX_VISIBLE)
  const remaining = matched.length - MAX_VISIBLE

  const badges = (
    <>
      {visible.map((tag) => (
        <TagBadge key={tag.id} name={tag.name} color={tag.color} />
      ))}
      {remaining > 0 && (
        <span className="inline-flex items-center rounded-full bg-muted px-1.5 py-0.5 text-[10px] font-medium leading-none text-muted-foreground">
          +{remaining}
        </span>
      )}
    </>
  )

  if (inline) return badges

  return (
    <div className="mt-1 flex flex-wrap gap-1">
      {badges}
    </div>
  )
}
