import { useState } from "react"
import { useNavigate } from "@tanstack/react-router"
import { formatDistanceToNow } from "date-fns"
import { MoreHorizontal, Pencil, Trash2 } from "lucide-react"
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
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  deleteNote,
  listNotes,
} from "@/api/gen/donejournal/notes/v1/notes-NoteService_connectquery"
import type { Note } from "@/api/gen/donejournal/notes/v1/notes_pb"
import { toDate } from "@/lib/dates"
import { NoteTagBadges } from "./note-tag-badges"

type Props = {
  note: Note
}

function stripMarkdown(text: string): string {
  return text
    .replace(/#{1,6}\s/g, "")
    .replace(/[*_~`]/g, "")
    .replace(/\[([^\]]+)\]\([^)]+\)/g, "$1")
    .replace(/\n/g, " ")
    .trim()
}

export function NoteCard({ note }: Props) {
  const navigate = useNavigate()
  const [deleteOpen, setDeleteOpen] = useState(false)
  const qc = useQueryClient()

  const invalidateNotes = () =>
    qc.invalidateQueries({
      queryKey: createConnectQueryKey({
        schema: listNotes,
        cardinality: "finite",
      }),
    })

  const deleteMutation = useMutation(deleteNote, {
    onSuccess: invalidateNotes,
  })

  const updatedAt = toDate(note.updatedAt)
  const bodyPreview = note.body ? stripMarkdown(note.body).slice(0, 120) : ""

  return (
    <>
      <div
        className="flex cursor-pointer items-start gap-3 rounded-lg border bg-card p-3 shadow-sm transition-colors hover:bg-accent/30"
        onClick={() =>
          navigate({ to: "/notes/$noteId", params: { noteId: note.id } })
        }
      >
        <div className="min-w-0 flex-1">
          <p className="text-sm leading-snug font-medium">{note.title}</p>
          {bodyPreview && (
            <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
              {bodyPreview}
            </p>
          )}
          <NoteTagBadges tagIds={note.tagIds} />
          {updatedAt && (
            <p className="mt-1 text-xs text-muted-foreground">
              {formatDistanceToNow(updatedAt, { addSuffix: true })}
            </p>
          )}
        </div>

        <div className="flex shrink-0 items-center">
          <DropdownMenu>
            <DropdownMenuTrigger
              className="inline-flex h-7 w-7 items-center justify-center rounded-md transition-colors hover:bg-muted"
              onClick={(e) => e.stopPropagation()}
            >
              <MoreHorizontal className="h-4 w-4" />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation()
                  navigate({
                    to: "/notes/$noteId/edit",
                    params: { noteId: note.id },
                  })
                }}
              >
                <Pencil className="mr-2 h-3.5 w-3.5" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                variant="destructive"
                onClick={(e) => {
                  e.stopPropagation()
                  setDeleteOpen(true)
                }}
              >
                <Trash2 className="mr-2 h-3.5 w-3.5" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

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
    </>
  )
}
