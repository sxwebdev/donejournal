import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { ProjectForm, type ProjectFormValues } from "./project-form"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  createProject,
  updateProject,
  listProjects,
} from "@/api/gen/donejournal/projects/v1/projects-ProjectService_connectquery"
import type { Project } from "@/api/gen/donejournal/projects/v1/projects_pb"
import { ConnectError } from "@connectrpc/connect"
import { toast } from "sonner"

type CreateProps = {
  mode: "create"
  open: boolean
  onOpenChange: (open: boolean) => void
}

type EditProps = {
  mode: "edit"
  project: Project
  open: boolean
  onOpenChange: (open: boolean) => void
}

type Props = CreateProps | EditProps

export function ProjectDialog(props: Props) {
  const qc = useQueryClient()

  const invalidate = () =>
    qc.invalidateQueries({
      queryKey: createConnectQueryKey({
        schema: listProjects,
        cardinality: "finite",
      }),
    })

  const createMutation = useMutation(createProject, {
    onSuccess: () => {
      invalidate()
      props.onOpenChange(false)
    },
  })

  const updateMutation = useMutation(updateProject, {
    onSuccess: () => {
      invalidate()
      props.onOpenChange(false)
    },
  })

  const handleSubmit = async (values: ProjectFormValues) => {
    try {
      if (props.mode === "create") {
        await createMutation.mutateAsync({
          name: values.name,
          description: values.description ?? "",
        })
      } else {
        await updateMutation.mutateAsync({
          id: props.project.id,
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
          name: props.project.name,
          description: props.project.description || undefined,
        }
      : undefined

  const isPending = createMutation.isPending || updateMutation.isPending

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {props.mode === "create" ? "New Project" : "Edit Project"}
          </DialogTitle>
        </DialogHeader>
        <ProjectForm
          defaultValues={defaultValues}
          onSubmit={handleSubmit}
          submitLabel={props.mode === "create" ? "Create" : "Save changes"}
          isPending={isPending}
        />
      </DialogContent>
    </Dialog>
  )
}
