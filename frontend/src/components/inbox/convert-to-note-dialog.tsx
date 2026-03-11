import { useEffect } from "react"
import { useForm, Controller } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import MDEditor from "@uiw/react-md-editor"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  convertToNote,
  listInboxItems,
} from "@/api/gen/donejournal/inbox/v1/inbox-InboxService_connectquery"
import { listNotes } from "@/api/gen/donejournal/notes/v1/notes-NoteService_connectquery"
import type { InboxItem } from "@/api/gen/donejournal/inbox/v1/inbox_pb"
import { useTheme } from "@/components/theme-provider"

const schema = z.object({
  title: z.string().min(1, "Title is required"),
  body: z.string().optional(),
})

type FormValues = z.infer<typeof schema>

type Props = {
  item: InboxItem
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ConvertToNoteDialog({ item, open, onOpenChange }: Props) {
  const qc = useQueryClient()
  const { theme } = useTheme()

  const {
    register,
    handleSubmit,
    control,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { title: item.data, body: "" },
  })

  useEffect(() => {
    if (open) {
      reset({ title: item.data, body: "" })
    }
  }, [open]) // eslint-disable-line react-hooks/exhaustive-deps

  const { mutate, isPending } = useMutation(convertToNote, {
    onSuccess: () => {
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: listInboxItems,
          cardinality: "finite",
        }),
      })
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: listNotes,
          cardinality: "finite",
        }),
      })
      onOpenChange(false)
    },
  })

  const onSubmit = (values: FormValues) => {
    mutate({
      inboxItemId: item.id,
      title: values.title,
      body: values.body ?? "",
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Convert to Note</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="note-title">
              Title <span className="text-destructive">*</span>
            </Label>
            <Input id="note-title" {...register("title")} />
            {errors.title && (
              <p className="text-xs text-destructive">{errors.title.message}</p>
            )}
          </div>

          <div className="space-y-1.5" data-color-mode={theme === "dark" ? "dark" : "light"}>
            <Label>Body</Label>
            <Controller
              name="body"
              control={control}
              render={({ field }) => (
                <MDEditor
                  value={field.value ?? ""}
                  onChange={(val) => field.onChange(val ?? "")}
                  height={300}
                  preview="edit"
                />
              )}
            />
          </div>

          <Button type="submit" className="w-full" disabled={isPending}>
            {isPending ? "Converting..." : "Convert to Note"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  )
}
