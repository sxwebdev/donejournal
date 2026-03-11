import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  createInboxItem,
  listInboxItems,
} from "@/api/gen/donejournal/inbox/v1/inbox-InboxService_connectquery"
import { Textarea } from "@/components/ui/textarea"
import { Button } from "@/components/ui/button"

const schema = z.object({
  data: z.string().min(1, "Can't be empty"),
})

type FormValues = z.infer<typeof schema>

export function InboxQuickAdd() {
  const qc = useQueryClient()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) })

  const { mutate, isPending } = useMutation(createInboxItem, {
    onSuccess: () => {
      reset()
      qc.invalidateQueries({ queryKey: createConnectQueryKey({ schema: listInboxItems, cardinality: "finite" }) })
    },
  })

  const onSubmit = (values: FormValues) => {
    mutate({ data: values.data, additionalData: "" })
  }

  return (
    <div className="rounded-xl border bg-card p-4 shadow-sm">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-2">
        <Textarea
          placeholder="Capture a thought..."
          rows={3}
          className="resize-y"
          {...register("data")}
          onKeyDown={(e) => {
            if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
              handleSubmit(onSubmit)()
            }
          }}
        />
        <div className="flex items-center justify-end gap-2">
          {errors.data && (
            <p className="text-xs text-destructive">{errors.data.message}</p>
          )}
          <Button type="submit" size="sm" disabled={isPending}>
            {isPending ? "Adding..." : "Add"}
          </Button>
        </div>
      </form>
    </div>
  )
}
