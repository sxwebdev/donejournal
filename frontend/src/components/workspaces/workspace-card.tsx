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
import { WorkspaceDialog } from "./workspace-dialog"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  deleteWorkspace,
  archiveWorkspace,
  unarchiveWorkspace,
  listWorkspaces,
} from "@/api/gen/donejournal/workspaces/v1/workspaces-WorkspaceService_connectquery"
import type { WorkspaceStats } from "@/api/gen/donejournal/workspaces/v1/workspaces_pb"
import { toDate } from "@/lib/dates"
import { cn } from "@/lib/utils"

type Props = {
  stats: WorkspaceStats
}

export function WorkspaceCard({ stats }: Props) {
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const qc = useQueryClient()
  const workspace = stats.workspace!

  const invalidate = () =>
    qc.invalidateQueries({
      queryKey: createConnectQueryKey({
        schema: listWorkspaces,
        cardinality: "finite",
      }),
    })

  const deleteMutation = useMutation(deleteWorkspace, { onSuccess: invalidate })
  const archiveMutation = useMutation(archiveWorkspace, { onSuccess: invalidate })
  const unarchiveMutation = useMutation(unarchiveWorkspace, { onSuccess: invalidate })

  const updatedAt = toDate(workspace.updatedAt)

  return (
    <>
      <div
        className={cn(
          "flex items-start gap-3 rounded-lg border bg-card p-3 shadow-sm transition-colors hover:bg-accent/30",
          workspace.archived && "opacity-60"
        )}
      >
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <p className="text-sm leading-snug font-medium">{workspace.name}</p>
            {workspace.archived && (
              <Badge variant="secondary" className="text-xs">
                Archived
              </Badge>
            )}
          </div>
          {workspace.description && (
            <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
              {workspace.description}
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
              {workspace.archived ? (
                <DropdownMenuItem
                  onClick={() => unarchiveMutation.mutate({ id: workspace.id })}
                >
                  <ArchiveRestore className="mr-2 h-3.5 w-3.5" />
                  Unarchive
                </DropdownMenuItem>
              ) : (
                <DropdownMenuItem
                  onClick={() => archiveMutation.mutate({ id: workspace.id })}
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

      <WorkspaceDialog
        mode="edit"
        workspace={workspace}
        open={editOpen}
        onOpenChange={setEditOpen}
      />

      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete workspace?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. Todos and notes in this workspace will
              become unassigned.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => deleteMutation.mutate({ id: workspace.id })}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
