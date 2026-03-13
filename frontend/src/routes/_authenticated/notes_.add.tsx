import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import { ArrowLeft } from "lucide-react"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  createNote,
  listNotes,
} from "@/api/gen/donejournal/notes/v1/notes-NoteService_connectquery"
import { NoteForm, type NoteFormValues } from "@/components/notes/note-form"
import { ConnectError } from "@connectrpc/connect"
import { toast } from "sonner"

export const Route = createFileRoute("/_authenticated/notes_/add")({
  component: AddNotePage,
})

function AddNotePage() {
  const navigate = useNavigate()
  const qc = useQueryClient()

  const mutation = useMutation(createNote, {
    onSuccess: (data) => {
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: listNotes,
          cardinality: "finite",
        }),
      })
      navigate({ to: "/notes/$noteId", params: { noteId: data.note!.id } })
    },
  })

  const handleSubmit = async (values: NoteFormValues) => {
    try {
      await mutation.mutateAsync({
        title: values.title,
        body: values.body ?? "",
        workspaceId: values.workspaceId,
      })
    } catch (err) {
      const message =
        err instanceof ConnectError ? err.rawMessage : "Something went wrong"
      toast.error(message)
    }
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <Link
        to="/notes"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Notes
      </Link>

      <h1 className="text-2xl font-semibold">New Note</h1>

      <NoteForm
        onSubmit={handleSubmit}
        submitLabel="Create"
        isPending={mutation.isPending}
      />
    </div>
  )
}
