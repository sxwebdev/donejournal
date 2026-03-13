import { useState } from "react"
import { useQuery } from "@connectrpc/connect-query"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  listTags,
  createTag,
  updateTag,
  deleteTag,
} from "@/api/gen/donejournal/tags/v1/tags-TagService_connectquery"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Pencil, Trash2, Plus, Check, X } from "lucide-react"
import { Skeleton } from "@/components/ui/skeleton"
import { TagBadge } from "./tag-badge"

const PRESET_COLORS = [
  "#ef4444",
  "#f97316",
  "#eab308",
  "#22c55e",
  "#06b6d4",
  "#3b82f6",
  "#6366f1",
  "#8b5cf6",
  "#ec4899",
  "#64748b",
]

export function TagManager() {
  const qc = useQueryClient()
  const { data, isLoading } = useQuery(listTags, { pageSize: 100 })
  const tags = data?.tags ?? []

  const [newName, setNewName] = useState("")
  const [newColor, setNewColor] = useState(PRESET_COLORS[6])
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editName, setEditName] = useState("")
  const [editColor, setEditColor] = useState("")
  const [deleteId, setDeleteId] = useState<string | null>(null)

  const invalidate = () =>
    qc.invalidateQueries({
      queryKey: createConnectQueryKey({
        schema: listTags,
        cardinality: "finite",
      }),
    })

  const createMutation = useMutation(createTag, {
    onSuccess: () => {
      invalidate()
      setNewName("")
      setNewColor(PRESET_COLORS[6])
    },
  })

  const updateMutation = useMutation(updateTag, {
    onSuccess: () => {
      invalidate()
      setEditingId(null)
    },
  })

  const deleteMutation = useMutation(deleteTag, {
    onSuccess: () => {
      invalidate()
      setDeleteId(null)
    },
  })

  const startEdit = (id: string, name: string, color: string) => {
    setEditingId(id)
    setEditName(name)
    setEditColor(color)
  }

  if (isLoading) {
    return (
      <div className="space-y-2">
        {[...Array(3)].map((_, i) => (
          <Skeleton key={i} className="h-12 w-full rounded-lg" />
        ))}
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <Input
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          placeholder="New tag name"
          className="flex-1"
          onKeyDown={(e) => {
            if (e.key === "Enter" && newName.trim()) {
              createMutation.mutate({ name: newName.trim(), color: newColor })
            }
          }}
        />
        <div className="flex gap-1">
          {PRESET_COLORS.map((c) => (
            <button
              key={c}
              type="button"
              onClick={() => setNewColor(c)}
              className="h-6 w-6 shrink-0 rounded-full border-2 transition-transform hover:scale-110"
              style={{
                backgroundColor: c,
                borderColor: c === newColor ? "currentColor" : "transparent",
              }}
            />
          ))}
        </div>
        <Button
          size="sm"
          disabled={!newName.trim() || createMutation.isPending}
          onClick={() =>
            createMutation.mutate({ name: newName.trim(), color: newColor })
          }
        >
          <Plus className="mr-1 h-3.5 w-3.5" />
          Add
        </Button>
      </div>

      {tags.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">
          No tags yet. Create one above.
        </p>
      ) : (
        <div className="space-y-2">
          {tags.map((tag) => (
            <div
              key={tag.id}
              className="flex items-center gap-3 rounded-lg border bg-card p-3"
            >
              {editingId === tag.id ? (
                <>
                  <Input
                    value={editName}
                    onChange={(e) => setEditName(e.target.value)}
                    className="h-8 flex-1"
                    onKeyDown={(e) => {
                      if (e.key === "Enter" && editName.trim()) {
                        updateMutation.mutate({
                          id: tag.id,
                          name: editName.trim(),
                          color: editColor,
                        })
                      }
                      if (e.key === "Escape") setEditingId(null)
                    }}
                    autoFocus
                  />
                  <div className="flex gap-1">
                    {PRESET_COLORS.map((c) => (
                      <button
                        key={c}
                        type="button"
                        onClick={() => setEditColor(c)}
                        className="h-5 w-5 shrink-0 rounded-full border-2 transition-transform hover:scale-110"
                        style={{
                          backgroundColor: c,
                          borderColor:
                            c === editColor ? "currentColor" : "transparent",
                        }}
                      />
                    ))}
                  </div>
                  <button
                    onClick={() =>
                      updateMutation.mutate({
                        id: tag.id,
                        name: editName.trim(),
                        color: editColor,
                      })
                    }
                    disabled={!editName.trim()}
                    className="rounded p-1 transition-colors hover:bg-accent"
                  >
                    <Check className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => setEditingId(null)}
                    className="rounded p-1 transition-colors hover:bg-accent"
                  >
                    <X className="h-4 w-4" />
                  </button>
                </>
              ) : (
                <>
                  <TagBadge name={tag.name} color={tag.color} />
                  <span className="flex-1" />
                  <button
                    onClick={() => startEdit(tag.id, tag.name, tag.color)}
                    className="rounded p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                  >
                    <Pencil className="h-3.5 w-3.5" />
                  </button>
                  <button
                    onClick={() => setDeleteId(tag.id)}
                    className="rounded p-1 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </>
              )}
            </div>
          ))}
        </div>
      )}

      <AlertDialog
        open={!!deleteId}
        onOpenChange={(open) => !open && setDeleteId(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete tag?</AlertDialogTitle>
            <AlertDialogDescription>
              This will remove the tag from all todos and notes. This action
              cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => deleteId && deleteMutation.mutate({ id: deleteId })}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
