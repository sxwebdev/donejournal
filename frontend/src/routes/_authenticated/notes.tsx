import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import { z } from "zod"
import { useEffect, useCallback, useState } from "react"
import { NoteList } from "@/components/notes/note-list"
import { buttonVariants } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Plus, Search } from "lucide-react"
import { WorkspaceSelector } from "@/components/workspaces/workspace-selector"

const notesSearchSchema = z.object({
  search: z.string().optional(),
  workspaceId: z.string().optional(),
})

export const Route = createFileRoute("/_authenticated/notes")({
  validateSearch: notesSearchSchema,
  component: NotesPage,
})

function NotesPage() {
  const search = Route.useSearch()
  const navigate = useNavigate({ from: "/notes" })
  const [searchInput, setSearchInput] = useState(search.search ?? "")

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      navigate({
        search: (prev) => ({
          ...prev,
          search: searchInput || undefined,
        }),
        replace: true,
      })
    }, 300)
    return () => clearTimeout(timer)
  }, [searchInput, navigate])

  const handleSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchInput(e.target.value)
    },
    []
  )

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Notes</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Your markdown notes
          </p>
        </div>
        <Link to="/notes/add" className={buttonVariants()}>
          <Plus className="mr-2 h-4 w-4" />
          New Note
        </Link>
      </div>

      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search notes..."
            value={searchInput}
            onChange={handleSearchChange}
            className="pl-9"
          />
        </div>
        <WorkspaceSelector
          value={search.workspaceId}
          onChange={(v) =>
            navigate({
              search: (prev) => ({ ...prev, workspaceId: v }),
              replace: true,
            })
          }
          placeholder="All workspaces"
          className="w-auto min-w-35"
        />
      </div>

      <NoteList search={search.search} workspaceId={search.workspaceId} />
    </div>
  )
}
