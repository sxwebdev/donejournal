import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { TodoForm, type TodoFormValues } from "./todo-form"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  createTodo,
  updateTodo,
  listTodos,
  getCalendarEntries,
} from "@/api/gen/donejournal/todos/v1/todos-TodoService_connectquery"
import type { Todo } from "@/api/gen/donejournal/todos/v1/todos_pb"
import { TodoStatus } from "@/api/gen/donejournal/todos/v1/todos_pb"
import { fromDateOnly, toDate } from "@/lib/dates"
import { ConnectError } from "@connectrpc/connect"
import { toast } from "sonner"

type CreateProps = {
  mode: "create"
  open: boolean
  onOpenChange: (open: boolean) => void
  initialDate?: Date
}

type EditProps = {
  mode: "edit"
  todo: Todo
  open: boolean
  onOpenChange: (open: boolean) => void
}

type Props = CreateProps | EditProps

export function TodoDialog(props: Props) {
  const qc = useQueryClient()

  const invalidate = () =>
    Promise.all([
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: listTodos,
          cardinality: "finite",
        }),
      }),
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: getCalendarEntries,
          cardinality: "finite",
        }),
      }),
    ])

  const createMutation = useMutation(createTodo, {
    onSuccess: () => {
      invalidate()
      props.onOpenChange(false)
    },
  })

  const updateMutation = useMutation(updateTodo, {
    onSuccess: () => {
      invalidate()
      props.onOpenChange(false)
    },
  })

  const handleSubmit = async (values: TodoFormValues) => {
    try {
      if (props.mode === "create") {
        await createMutation.mutateAsync({
          title: values.title,
          description: values.description ?? "",
          plannedDate: values.plannedDate
            ? fromDateOnly(values.plannedDate)
            : undefined,
          workspaceId: values.workspaceId,
        })
      } else {
        await updateMutation.mutateAsync({
          id: props.todo.id,
          title: values.title,
          description: values.description,
          plannedDate: values.plannedDate
            ? fromDateOnly(values.plannedDate)
            : undefined,
          status: values.status !== undefined ? values.status : undefined,
          workspaceId: values.workspaceId,
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
          title: props.todo.title,
          description: props.todo.description || undefined,
          plannedDate: toDate(props.todo.plannedDate),
          status: props.todo.status as TodoStatus,
          workspaceId: props.todo.workspaceId || undefined,
        }
      : props.initialDate
        ? { plannedDate: props.initialDate }
        : undefined

  const isPending = createMutation.isPending || updateMutation.isPending

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {props.mode === "create" ? "New Todo" : "Edit Todo"}
          </DialogTitle>
        </DialogHeader>
        <TodoForm
          defaultValues={defaultValues}
          onSubmit={handleSubmit}
          submitLabel={props.mode === "create" ? "Create" : "Save changes"}
          isPending={isPending}
          showStatus={props.mode === "edit"}
        />
      </DialogContent>
    </Dialog>
  )
}
