import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { z } from "zod"
import { useState, useEffect, useCallback } from "react"
import { NoteList } from "@/components/notes/note-list"
import { NoteDialog } from "@/components/notes/note-dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Plus, Search } from "lucide-react"

const notesSearchSchema = z.object({
  search: z.string().optional(),
})

export const Route = createFileRoute("/_authenticated/notes")({
  validateSearch: notesSearchSchema,
  component: NotesPage,
})

function NotesPage() {
  const [createOpen, setCreateOpen] = useState(false)
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
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          New Note
        </Button>
      </div>

      <div className="relative">
        <Search className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search notes..."
          value={searchInput}
          onChange={handleSearchChange}
          className="pl-9"
        />
      </div>

      <NoteList search={search.search} />

      <NoteDialog
        mode="create"
        open={createOpen}
        onOpenChange={setCreateOpen}
      />
    </div>
  )
}
