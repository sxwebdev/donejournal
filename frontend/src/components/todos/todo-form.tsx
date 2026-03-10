import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { format } from "date-fns"
import { CalendarIcon } from "lucide-react"
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { cn } from "@/lib/utils"
import { TodoStatus } from "@/api/gen/donejournal/todos/v1/todos_pb"

const schema = z.object({
  title: z.string().min(1, "Title is required").max(200),
  description: z.string().max(1000).optional(),
  plannedDate: z.date(),
  status: z.nativeEnum(TodoStatus).optional(),
})

export type TodoFormValues = z.infer<typeof schema>

const STATUS_OPTIONS: { value: TodoStatus; label: string }[] = [
  { value: TodoStatus.PENDING, label: "Pending" },
  { value: TodoStatus.IN_PROGRESS, label: "In Progress" },
  { value: TodoStatus.COMPLETED, label: "Completed" },
  { value: TodoStatus.CANCELLED, label: "Cancelled" },
]

type Props = {
  defaultValues?: Partial<TodoFormValues>
  onSubmit: (values: TodoFormValues) => Promise<void>
  submitLabel?: string
  isPending?: boolean
  showStatus?: boolean
}

export function TodoForm({
  defaultValues,
  onSubmit,
  submitLabel = "Save",
  isPending,
  showStatus,
}: Props) {
  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<TodoFormValues>({
    resolver: zodResolver(schema),
    defaultValues,
  })

  const plannedDate = watch("plannedDate")
  const status = watch("status")

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="space-y-1.5">
        <Label htmlFor="title">Title</Label>
        <Input
          id="title"
          placeholder="What needs to be done?"
          {...register("title")}
        />
        {errors.title && (
          <p className="text-xs text-destructive">{errors.title.message}</p>
        )}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="description">Description</Label>
        <Textarea
          id="description"
          placeholder="Add more details..."
          rows={3}
          {...register("description")}
        />
        {errors.description && (
          <p className="text-xs text-destructive">
            {errors.description.message}
          </p>
        )}
      </div>

      <div className="space-y-1.5">
        <Label>Planned date</Label>
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

      {showStatus && (
        <div className="space-y-1.5">
          <Label>Status</Label>
          <Select
            value={status !== undefined ? String(status) : ""}
            onValueChange={(v) => setValue("status", Number(v) as TodoStatus)}
          >
            <SelectTrigger className="w-full">
              <SelectValue>
                {status !== undefined
                  ? (STATUS_OPTIONS.find((o) => o.value === status)?.label ??
                    "Select status")
                  : "Select status"}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {STATUS_OPTIONS.map((opt) => (
                <SelectItem key={opt.value} value={String(opt.value)}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? "Saving..." : submitLabel}
      </Button>
    </form>
  )
}
