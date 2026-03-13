import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import { useState, useEffect, useCallback } from "react"
import { ProjectList } from "@/components/projects/project-list"
import { ProjectDialog } from "@/components/projects/project-dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Plus, Search } from "lucide-react"

const projectsSearchSchema = z.object({
  search: z.string().optional(),
  includeArchived: z.boolean().optional(),
})

export const Route = createFileRoute("/_authenticated/projects")({
  validateSearch: projectsSearchSchema,
  component: ProjectsPage,
})

function ProjectsPage() {
  const [createOpen, setCreateOpen] = useState(false)
  const search = Route.useSearch()
  const navigate = Route.useNavigate()
  const [searchInput, setSearchInput] = useState(search.search ?? "")

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

  const toggleArchived = useCallback(() => {
    navigate({
      search: (prev) => ({
        ...prev,
        includeArchived: prev.includeArchived ? undefined : true,
      }),
      replace: true,
    })
  }, [navigate])

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Projects</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Organize your tasks and notes
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          New Project
        </Button>
      </div>

      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search projects..."
            value={searchInput}
            onChange={handleSearchChange}
            className="pl-9"
          />
        </div>
        <div className="flex items-center gap-2">
          <Switch
            id="archived"
            checked={search.includeArchived ?? false}
            onCheckedChange={toggleArchived}
          />
          <Label htmlFor="archived" className="text-sm whitespace-nowrap">
            Show archived
          </Label>
        </div>
      </div>

      <ProjectList
        search={search.search}
        includeArchived={search.includeArchived}
      />

      <ProjectDialog
        mode="create"
        open={createOpen}
        onOpenChange={setCreateOpen}
      />
    </div>
  )
}
