import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { WorkspaceForm, type WorkspaceFormValues } from "./workspace-form"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  createWorkspace,
  updateWorkspace,
  listWorkspaces,
} from "@/api/gen/donejournal/workspaces/v1/workspaces-WorkspaceService_connectquery"
import type { Workspace } from "@/api/gen/donejournal/workspaces/v1/workspaces_pb"
import { ConnectError } from "@connectrpc/connect"
import { toast } from "sonner"

type CreateProps = {
  mode: "create"
  open: boolean
  onOpenChange: (open: boolean) => void
}

type EditProps = {
  mode: "edit"
  workspace: Workspace
  open: boolean
  onOpenChange: (open: boolean) => void
}

type Props = CreateProps | EditProps

export function WorkspaceDialog(props: Props) {
  const qc = useQueryClient()

  const invalidate = () =>
    qc.invalidateQueries({
      queryKey: createConnectQueryKey({
        schema: listWorkspaces,
        cardinality: "finite",
      }),
    })

  const createMutation = useMutation(createWorkspace, {
    onSuccess: () => {
      invalidate()
      props.onOpenChange(false)
    },
  })

  const updateMutation = useMutation(updateWorkspace, {
    onSuccess: () => {
      invalidate()
      props.onOpenChange(false)
    },
  })

  const handleSubmit = async (values: WorkspaceFormValues) => {
    try {
      if (props.mode === "create") {
        await createMutation.mutateAsync({
          name: values.name,
          description: values.description ?? "",
        })
      } else {
        await updateMutation.mutateAsync({
          id: props.workspace.id,
          name: values.name,
          description: values.description,
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
          name: props.workspace.name,
          description: props.workspace.description || undefined,
        }
      : undefined

  const isPending = createMutation.isPending || updateMutation.isPending

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {props.mode === "create" ? "New Workspace" : "Edit Workspace"}
          </DialogTitle>
        </DialogHeader>
        <WorkspaceForm
          defaultValues={defaultValues}
          onSubmit={handleSubmit}
          submitLabel={props.mode === "create" ? "Create" : "Save changes"}
          isPending={isPending}
        />
      </DialogContent>
    </Dialog>
  )
}
