import { useState, useRef, useEffect } from "react"
import { useQuery } from "@connectrpc/connect-query"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  listTags,
  createTag,
} from "@/api/gen/donejournal/tags/v1/tags-TagService_connectquery"
import { TagBadge } from "./tag-badge"
import { cn } from "@/lib/utils"
import { Plus } from "lucide-react"

type Props = {
  value: string[]
  onChange: (value: string[]) => void
  className?: string
}

export function TagSelector({ value, onChange, className }: Props) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const inputRef = useRef<HTMLInputElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const qc = useQueryClient()

  const { data } = useQuery(listTags, { pageSize: 100 })
  const tags = data?.tags ?? []

  const createMutation = useMutation(createTag, {
    onSuccess: (resp) => {
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: listTags,
          cardinality: "finite",
        }),
      })
      if (resp.tag) {
        onChange([...value, resp.tag.id])
      }
      setSearch("")
    },
  })

  const selectedTags = tags.filter((t) => value.includes(t.id))
  const availableTags = tags.filter(
    (t) =>
      !value.includes(t.id) &&
      t.name.toLowerCase().includes(search.toLowerCase())
  )
  const canCreate =
    search.trim() &&
    !tags.some((t) => t.name.toLowerCase() === search.trim().toLowerCase())

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      ) {
        setOpen(false)
      }
    }
    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  return (
    <div ref={containerRef} className={cn("relative", className)}>
      <div
        className="flex min-h-9 flex-wrap items-center gap-1 rounded-md border border-input bg-transparent px-2 py-1.5 text-sm shadow-xs cursor-text"
        onClick={() => {
          setOpen(true)
          inputRef.current?.focus()
        }}
      >
        {selectedTags.map((tag) => (
          <TagBadge
            key={tag.id}
            name={tag.name}
            color={tag.color}
            onRemove={() => onChange(value.filter((id) => id !== tag.id))}
          />
        ))}
        <input
          ref={inputRef}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          onFocus={() => setOpen(true)}
          placeholder={selectedTags.length ? "" : "Add tags..."}
          className="min-w-16 flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
        />
      </div>

      {open && (availableTags.length > 0 || canCreate) && (
        <div className="absolute z-50 mt-1 w-full rounded-md border bg-popover p-1 shadow-md">
          {availableTags.map((tag) => (
            <button
              key={tag.id}
              type="button"
              onClick={() => {
                onChange([...value, tag.id])
                setSearch("")
              }}
              className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm transition-colors hover:bg-accent"
            >
              <span
                className="h-3 w-3 shrink-0 rounded-full"
                style={{ backgroundColor: tag.color }}
              />
              {tag.name}
            </button>
          ))}
          {canCreate && (
            <button
              type="button"
              onClick={() => createMutation.mutate({ name: search.trim() })}
              disabled={createMutation.isPending}
              className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm text-primary transition-colors hover:bg-accent"
            >
              <Plus className="h-3 w-3" />
              Create "{search.trim()}"
            </button>
          )}
        </div>
      )}
    </div>
  )
}
