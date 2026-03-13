import { createFileRoute, Link } from "@tanstack/react-router"
import { useState } from "react"
import { useQuery } from "@connectrpc/connect-query"
import { formatDistanceToNow } from "date-fns"
import MDEditor from "@uiw/react-md-editor"
import { ArrowLeft, Pencil, Trash2 } from "lucide-react"
import { getNote } from "@/api/gen/donejournal/notes/v1/notes-NoteService_connectquery"
import { Button, buttonVariants } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
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
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  deleteNote,
  listNotes,
} from "@/api/gen/donejournal/notes/v1/notes-NoteService_connectquery"
import { useTheme } from "@/components/theme-provider"
import { toDate } from "@/lib/dates"
import { useNavigate } from "@tanstack/react-router"

export const Route = createFileRoute("/_authenticated/notes_/$noteId")({
  component: NoteDetailPage,
})

function NoteDetailPage() {
  const { noteId } = Route.useParams()
  const navigate = useNavigate()
  const { theme } = useTheme()
  const qc = useQueryClient()
  const [deleteOpen, setDeleteOpen] = useState(false)

  const { data, isLoading } = useQuery(getNote, { id: noteId })
  const note = data?.note

  const deleteMutation = useMutation(deleteNote, {
    onSuccess: () => {
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: listNotes,
          cardinality: "finite",
        }),
      })
      navigate({ to: "/notes" })
    },
  })

  if (isLoading) {
    return (
      <div className="mx-auto max-w-3xl space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-6 w-32" />
        <Skeleton className="h-64 w-full" />
      </div>
    )
  }

  if (!note) {
    return (
      <div className="mx-auto max-w-3xl space-y-4">
        <Link
          to="/notes"
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Notes
        </Link>
        <p className="text-muted-foreground">Note not found.</p>
      </div>
    )
  }

  const updatedAt = toDate(note.updatedAt)

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <Link
        to="/notes"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Notes
      </Link>

      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0 flex-1">
          <h1 className="text-2xl font-semibold wrap-break-word">
            {note.title}
          </h1>
          {updatedAt && (
            <p className="mt-1 text-sm text-muted-foreground">
              Updated {formatDistanceToNow(updatedAt, { addSuffix: true })}
            </p>
          )}
        </div>
        <div className="flex shrink-0 gap-2">
          <Link
            to="/notes/$noteId/edit"
            params={{ noteId }}
            className={buttonVariants({ variant: "outline", size: "sm" })}
          >
            <Pencil className="mr-1.5 h-3.5 w-3.5" />
            Edit
          </Link>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setDeleteOpen(true)}
          >
            <Trash2 className="mr-1.5 h-3.5 w-3.5" />
            Delete
          </Button>
        </div>
      </div>

      {note.body ? (
        <div data-color-mode={theme === "dark" ? "dark" : "light"}>
          <MDEditor.Markdown
            source={note.body}
            className="rounded-lg border bg-card p-6"
          />
        </div>
      ) : (
        <p className="text-sm text-muted-foreground italic">No content</p>
      )}

      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete note?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => deleteMutation.mutate({ id: note.id })}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
