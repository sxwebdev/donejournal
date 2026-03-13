import { useState } from "react"
import { formatDistanceToNow } from "date-fns"
import { MoreHorizontal, Pencil, Trash2, Archive, ArchiveRestore } from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Badge } from "@/components/ui/badge"
import { ProjectDialog } from "./project-dialog"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  deleteProject,
  archiveProject,
  unarchiveProject,
  listProjects,
} from "@/api/gen/donejournal/projects/v1/projects-ProjectService_connectquery"
import type { ProjectStats } from "@/api/gen/donejournal/projects/v1/projects_pb"
import { toDate } from "@/lib/dates"
import { cn } from "@/lib/utils"

type Props = {
  stats: ProjectStats
}

export function ProjectCard({ stats }: Props) {
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const qc = useQueryClient()
  const project = stats.project!

  const invalidate = () =>
    qc.invalidateQueries({
      queryKey: createConnectQueryKey({
        schema: listProjects,
        cardinality: "finite",
      }),
    })

  const deleteMutation = useMutation(deleteProject, { onSuccess: invalidate })
  const archiveMutation = useMutation(archiveProject, { onSuccess: invalidate })
  const unarchiveMutation = useMutation(unarchiveProject, { onSuccess: invalidate })

  const updatedAt = toDate(project.updatedAt)

  return (
    <>
      <div
        className={cn(
          "flex items-start gap-3 rounded-lg border bg-card p-3 shadow-sm transition-colors hover:bg-accent/30",
          project.archived && "opacity-60"
        )}
      >
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <p className="text-sm leading-snug font-medium">{project.name}</p>
            {project.archived && (
              <Badge variant="secondary" className="text-xs">
                Archived
              </Badge>
            )}
          </div>
          {project.description && (
            <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
              {project.description}
            </p>
          )}
          <div className="mt-1 flex items-center gap-3 text-xs text-muted-foreground">
            <span>{stats.todoCount} todos</span>
            <span>{stats.completedTodoCount} done</span>
            <span>{stats.noteCount} notes</span>
            {updatedAt && (
              <span>{formatDistanceToNow(updatedAt, { addSuffix: true })}</span>
            )}
          </div>
        </div>

        <div className="flex shrink-0 items-center">
          <DropdownMenu>
            <DropdownMenuTrigger className="inline-flex h-7 w-7 items-center justify-center rounded-md transition-colors hover:bg-muted">
              <MoreHorizontal className="h-4 w-4" />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => setEditOpen(true)}>
                <Pencil className="mr-2 h-3.5 w-3.5" />
                Edit
              </DropdownMenuItem>
              {project.archived ? (
                <DropdownMenuItem
                  onClick={() => unarchiveMutation.mutate({ id: project.id })}
                >
                  <ArchiveRestore className="mr-2 h-3.5 w-3.5" />
                  Unarchive
                </DropdownMenuItem>
              ) : (
                <DropdownMenuItem
                  onClick={() => archiveMutation.mutate({ id: project.id })}
                >
                  <Archive className="mr-2 h-3.5 w-3.5" />
                  Archive
                </DropdownMenuItem>
              )}
              <DropdownMenuSeparator />
              <DropdownMenuItem
                variant="destructive"
                onClick={() => setDeleteOpen(true)}
              >
                <Trash2 className="mr-2 h-3.5 w-3.5" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      <ProjectDialog
        mode="edit"
        project={project}
        open={editOpen}
        onOpenChange={setEditOpen}
      />

      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete project?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. Todos and notes in this project will
              become unassigned.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => deleteMutation.mutate({ id: project.id })}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
