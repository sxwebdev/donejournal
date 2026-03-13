import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import { ArrowLeft } from "lucide-react"
import { useQuery } from "@connectrpc/connect-query"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  getNote,
  updateNote,
  listNotes,
} from "@/api/gen/donejournal/notes/v1/notes-NoteService_connectquery"
import { NoteForm, type NoteFormValues } from "@/components/notes/note-form"
import { Skeleton } from "@/components/ui/skeleton"
import { ConnectError } from "@connectrpc/connect"
import { toast } from "sonner"

export const Route = createFileRoute("/_authenticated/notes_/$noteId_/edit")({
  component: EditNotePage,
})

function EditNotePage() {
  const { noteId } = Route.useParams()
  const navigate = useNavigate()
  const qc = useQueryClient()

  const { data, isLoading } = useQuery(getNote, { id: noteId })
  const note = data?.note

  const mutation = useMutation(updateNote, {
    onSuccess: () => {
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: listNotes,
          cardinality: "finite",
        }),
      })
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: getNote,
          input: { id: noteId },
          cardinality: "finite",
        }),
      })
      navigate({ to: "/notes/$noteId", params: { noteId } })
    },
  })

  const handleSubmit = async (values: NoteFormValues) => {
    try {
      await mutation.mutateAsync({
        id: noteId,
        title: values.title,
        body: values.body ?? "",
        workspaceId: values.workspaceId,
        tagIds: values.tagIds ?? [],
      })
    } catch (err) {
      const message =
        err instanceof ConnectError ? err.rawMessage : "Something went wrong"
      toast.error(message)
    }
  }

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

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <Link
        to="/notes/$noteId"
        params={{ noteId }}
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Note
      </Link>

      <h1 className="text-2xl font-semibold">Edit Note</h1>

      <NoteForm
        defaultValues={{
          title: note.title,
          body: note.body || undefined,
          workspaceId: note.workspaceId || undefined,
          tagIds: note.tagIds.length > 0 ? [...note.tagIds] : undefined,
        }}
        onSubmit={handleSubmit}
        submitLabel="Save changes"
        isPending={mutation.isPending}
      />
    </div>
  )
}
