import { useQuery } from "@connectrpc/connect-query"
import { listTags } from "@/api/gen/donejournal/tags/v1/tags-TagService_connectquery"
import { TagBadge } from "@/components/tags/tag-badge"

type Props = {
  tagIds: string[]
}

export function NoteTagBadges({ tagIds }: Props) {
  const { data } = useQuery(listTags, { pageSize: 100 })

  if (!tagIds.length) return null

  const tags = data?.tags ?? []
  const matched = tags.filter((t) => tagIds.includes(t.id))

  if (!matched.length) return null

  return (
    <div className="mt-1 flex flex-wrap gap-1">
      {matched.map((tag) => (
        <TagBadge key={tag.id} name={tag.name} color={tag.color} />
      ))}
    </div>
  )
}
