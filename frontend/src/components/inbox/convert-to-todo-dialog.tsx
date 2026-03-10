import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { format } from "date-fns"
import { CalendarIcon } from "lucide-react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button, buttonVariants } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Calendar } from "@/components/ui/calendar"
import { cn } from "@/lib/utils"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  convertToTodo,
  listInboxItems,
} from "@/api/gen/donejournal/inbox/v1/inbox-InboxService_connectquery"
import type { InboxItem } from "@/api/gen/donejournal/inbox/v1/inbox_pb"
import {
  listTodos,
  getCalendarEntries,
} from "@/api/gen/donejournal/todos/v1/todos-TodoService_connectquery"
import { fromDate } from "@/lib/dates"

const schema = z.object({
  title: z.string().min(1, "Title is required").max(200),
  description: z.string().max(1000).optional(),
  plannedDate: z.date(),
})

type FormValues = z.infer<typeof schema>

type Props = {
  item: InboxItem
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ConvertToTodoDialog({ item, open, onOpenChange }: Props) {
  const qc = useQueryClient()

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { title: item.data },
  })

  useEffect(() => {
    if (open) {
      reset({ title: item.data, description: "", plannedDate: undefined })
    }
  }, [open]) // eslint-disable-line react-hooks/exhaustive-deps

  const { mutate, isPending } = useMutation(convertToTodo, {
    onSuccess: () => {
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: listInboxItems,
          cardinality: "finite",
        }),
      })
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: listTodos,
          cardinality: "finite",
        }),
      })
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: getCalendarEntries,
          cardinality: "finite",
        }),
      })
      onOpenChange(false)
    },
  })

  const plannedDate = watch("plannedDate")

  const onSubmit = (values: FormValues) => {
    mutate({
      inboxItemId: item.id,
      title: values.title,
      description: values.description ?? "",
      plannedDate: fromDate(values.plannedDate),
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Convert to Todo</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="title">
              Title <span className="text-destructive">*</span>
            </Label>
            <Input id="title" {...register("title")} />
            {errors.title && (
              <p className="text-xs text-destructive">{errors.title.message}</p>
            )}
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="description">Description</Label>
            <Textarea id="description" rows={2} {...register("description")} />
          </div>

          <div className="space-y-1.5">
            <Label>
              Planned date <span className="text-destructive">*</span>
            </Label>
            <Popover>
              <PopoverTrigger
                type="button"
                className={cn(
                  buttonVariants({ variant: "outline" }),
                  "w-full justify-start text-left font-normal",
                  !plannedDate && "text-muted-foreground"
                )}
              >
                <CalendarIcon className="mr-2 h-4 w-4" />
                {plannedDate ? format(plannedDate, "PPP") : "Pick a date"}
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0" align="start">
                <Calendar
                  mode="single"
                  selected={plannedDate}
                  onSelect={(d) =>
                    setValue("plannedDate", d!, { shouldValidate: true })
                  }
                />
              </PopoverContent>
            </Popover>
            {errors.plannedDate && (
              <p className="text-xs text-destructive">
                {errors.plannedDate.message}
              </p>
            )}
          </div>

          <Button type="submit" className="w-full" disabled={isPending}>
            {isPending ? "Converting..." : "Convert to Todo"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  )
}
