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
import { cn } from "@/lib/utils"

const MAX = 200

const schema = z.object({
  data: z.string().min(1, "Can't be empty").max(MAX, `Max ${MAX} characters`),
})

type FormValues = z.infer<typeof schema>

export function InboxQuickAdd() {
  const qc = useQueryClient()

  const {
    register,
    handleSubmit,
    reset,
    watch,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) })

  const { mutate, isPending } = useMutation(createInboxItem, {
    onSuccess: () => {
      reset()
      qc.invalidateQueries({ queryKey: createConnectQueryKey({ schema: listInboxItems, cardinality: "finite" }) })
    },
  })

  const value = watch("data") ?? ""

  const onSubmit = (values: FormValues) => {
    mutate({ data: values.data, additionalData: "" })
  }

  return (
    <div className="rounded-xl border bg-card p-4 shadow-sm">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-2">
        <Textarea
          placeholder="Capture a thought..."
          rows={2}
          className="resize-none"
          {...register("data")}
          onKeyDown={(e) => {
            if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
              handleSubmit(onSubmit)()
            }
          }}
        />
        <div className="flex items-center justify-between">
          <span className={cn("text-xs text-muted-foreground", value.length > MAX && "text-destructive")}>
            {value.length}/{MAX}
          </span>
          <div className="flex items-center gap-2">
            {errors.data && (
              <p className="text-xs text-destructive">{errors.data.message}</p>
            )}
            <Button type="submit" size="sm" disabled={isPending}>
              {isPending ? "Adding..." : "Add"}
            </Button>
          </div>
        </div>
      </form>
    </div>
  )
}
