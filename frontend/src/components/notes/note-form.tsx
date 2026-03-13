import { useForm, Controller } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import MDEditor from "@uiw/react-md-editor"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useTheme } from "@/components/theme-provider"
import { ProjectSelector } from "@/components/projects/project-selector"

const schema = z.object({
  title: z.string().min(1, "Title is required").max(200),
  body: z.string().optional(),
  projectId: z.string().optional(),
})

export type NoteFormValues = z.infer<typeof schema>

type Props = {
  defaultValues?: Partial<NoteFormValues>
  onSubmit: (values: NoteFormValues) => Promise<void>
  submitLabel?: string
  isPending?: boolean
}

export function NoteForm({
  defaultValues,
  onSubmit,
  submitLabel = "Save",
  isPending,
}: Props) {
  const { theme } = useTheme()
  const {
    register,
    handleSubmit,
    control,
    watch,
    setValue,
    formState: { errors },
  } = useForm<NoteFormValues>({
    resolver: zodResolver(schema),
    defaultValues,
  })

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="space-y-1.5">
        <Label htmlFor="title">Title</Label>
        <Input
          id="title"
          placeholder="Note title"
          {...register("title")}
        />
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

      <div className="space-y-1.5">
        <Label>Project</Label>
        <ProjectSelector
          value={watch("projectId")}
          onChange={(v) => setValue("projectId", v)}
        />
      </div>

      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? "Saving..." : submitLabel}
      </Button>
    </form>
  )
}
