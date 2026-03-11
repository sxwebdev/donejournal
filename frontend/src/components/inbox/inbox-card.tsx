import { useState } from "react"
import { formatDistanceToNow } from "date-fns"
import { ArrowRight, FileText, Pencil, Trash2, Check, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
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
import { ConvertToTodoDialog } from "./convert-to-todo-dialog"
import { ConvertToNoteDialog } from "./convert-to-note-dialog"
import { useMutation, createConnectQueryKey } from "@connectrpc/connect-query"
import { useQueryClient } from "@tanstack/react-query"
import {
  deleteInboxItem,
  updateInboxItem,
  listInboxItems,
} from "@/api/gen/donejournal/inbox/v1/inbox-InboxService_connectquery"
import type { InboxItem } from "@/api/gen/donejournal/inbox/v1/inbox_pb"
import { toDate } from "@/lib/dates"

type Props = {
  item: InboxItem
}

export function InboxCard({ item }: Props) {
  const [convertTodoOpen, setConvertTodoOpen] = useState(false)
  const [convertNoteOpen, setConvertNoteOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [editing, setEditing] = useState(false)
  const [editValue, setEditValue] = useState(item.data)
  const qc = useQueryClient()

  const updateMutation = useMutation(updateInboxItem, {
    onSuccess: () => {
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: listInboxItems,
          cardinality: "finite",
        }),
      })
      setEditing(false)
    },
  })

  const deleteMutation = useMutation(deleteInboxItem, {
    onSuccess: () =>
      qc.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: listInboxItems,
          cardinality: "finite",
        }),
      }),
  })

  const handleSave = () => {
    if (!editValue.trim()) return
    updateMutation.mutate({
      id: item.id,
      data: editValue.trim(),
      additionalData: item.additionalData,
    })
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) handleSave()
    if (e.key === "Escape") {
      setEditValue(item.data)
      setEditing(false)
    }
  }

  const createdAt = toDate(item.createdAt)
  const timeAgo = createdAt
    ? formatDistanceToNow(createdAt, { addSuffix: true })
    : ""

  return (
    <>
      <div className="flex items-start gap-3 rounded-xl border bg-card p-4 shadow-sm">
        <div className="min-w-0 flex-1">
          {editing ? (
            <Textarea
              autoFocus
              value={editValue}
              onChange={(e) => setEditValue(e.target.value)}
              onKeyDown={handleKeyDown}
              rows={3}
              className="text-sm resize-y"
            />
          ) : (
            <p className="text-sm leading-snug whitespace-pre-wrap">{item.data}</p>
          )}
          {!editing && timeAgo && (
            <p className="mt-1 text-xs text-muted-foreground">{timeAgo}</p>
          )}
        </div>

        <div className="flex shrink-0 items-center gap-1">
          {editing ? (
            <>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                onClick={handleSave}
                disabled={updateMutation.isPending}
              >
                <Check className="h-3.5 w-3.5 text-green-600" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                onClick={() => {
                  setEditValue(item.data)
                  setEditing(false)
                }}
              >
                <X className="h-3.5 w-3.5" />
              </Button>
            </>
          ) : (
            <>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                title="Convert to todo"
                onClick={() => setConvertTodoOpen(true)}
              >
                <ArrowRight className="h-3.5 w-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                title="Convert to note"
                onClick={() => setConvertNoteOpen(true)}
              >
                <FileText className="h-3.5 w-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                onClick={() => setEditing(true)}
              >
                <Pencil className="h-3.5 w-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 text-destructive hover:text-destructive"
                onClick={() => setDeleteOpen(true)}
                disabled={deleteMutation.isPending}
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            </>
          )}
        </div>
      </div>

      <ConvertToTodoDialog
        item={item}
        open={convertTodoOpen}
        onOpenChange={setConvertTodoOpen}
      />

      <ConvertToNoteDialog
        item={item}
        open={convertNoteOpen}
        onOpenChange={setConvertNoteOpen}
      />

      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete item?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => deleteMutation.mutate({ id: item.id })}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
