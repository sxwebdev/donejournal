import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { NoteForm, type NoteFormValues } from "./note-form"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  createNote,
  updateNote,
  listNotes,
} from "@/api/gen/donejournal/notes/v1/notes-NoteService_connectquery"
import type { Note } from "@/api/gen/donejournal/notes/v1/notes_pb"
import { ConnectError } from "@connectrpc/connect"
import { toast } from "sonner"

type CreateProps = {
  mode: "create"
  open: boolean
  onOpenChange: (open: boolean) => void
}

type EditProps = {
  mode: "edit"
  note: Note
  open: boolean
  onOpenChange: (open: boolean) => void
}

type Props = CreateProps | EditProps

export function NoteDialog(props: Props) {
  const qc = useQueryClient()

  const invalidate = () =>
    qc.invalidateQueries({
      queryKey: createConnectQueryKey({
        schema: listNotes,
        cardinality: "finite",
      }),
    })

  const createMutation = useMutation(createNote, {
    onSuccess: () => {
      invalidate()
      props.onOpenChange(false)
    },
  })

  const updateMutation = useMutation(updateNote, {
    onSuccess: () => {
      invalidate()
      props.onOpenChange(false)
    },
  })

  const handleSubmit = async (values: NoteFormValues) => {
    try {
      if (props.mode === "create") {
        await createMutation.mutateAsync({
          title: values.title,
          body: values.body ?? "",
          projectId: values.projectId,
        })
      } else {
        await updateMutation.mutateAsync({
          id: props.note.id,
          title: values.title,
          body: values.body ?? "",
          projectId: values.projectId,
        })
      }
    } catch (err) {
      const message =
        err instanceof ConnectError ? err.rawMessage : "Something went wrong"
      toast.error(message)
    }
  }

  const defaultValues =
    props.mode === "edit"
      ? {
          title: props.note.title,
          body: props.note.body || undefined,
          projectId: props.note.projectId || undefined,
        }
      : undefined

  const isPending = createMutation.isPending || updateMutation.isPending

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {props.mode === "create" ? "New Note" : "Edit Note"}
          </DialogTitle>
        </DialogHeader>
        {props.open && (
          <NoteForm
            defaultValues={defaultValues}
            onSubmit={handleSubmit}
            submitLabel={props.mode === "create" ? "Create" : "Save changes"}
            isPending={isPending}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}
