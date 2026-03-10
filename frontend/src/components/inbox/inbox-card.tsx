import { useState } from "react"
import { formatDistanceToNow } from "date-fns"
import { ArrowRight, Pencil, Trash2, Check, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ConvertToTodoDialog } from "./convert-to-todo-dialog"
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
  const [convertOpen, setConvertOpen] = useState(false)
  const [editing, setEditing] = useState(false)
  const [editValue, setEditValue] = useState(item.data)
  const qc = useQueryClient()

  const updateMutation = useMutation(updateInboxItem, {
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: createConnectQueryKey({ schema: listInboxItems, cardinality: "finite" }) })
      setEditing(false)
    },
  })

  const deleteMutation = useMutation(deleteInboxItem, {
    onSuccess: () => qc.invalidateQueries({ queryKey: createConnectQueryKey({ schema: listInboxItems, cardinality: "finite" }) }),
  })

  const handleSave = () => {
    if (!editValue.trim()) return
    updateMutation.mutate({ id: item.id, data: editValue.trim(), additionalData: item.additionalData })
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") handleSave()
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
            <Input
              autoFocus
              value={editValue}
              onChange={(e) => setEditValue(e.target.value)}
              onKeyDown={handleKeyDown}
              maxLength={200}
              className="h-7 text-sm"
            />
          ) : (
            <p className="text-sm leading-snug">{item.data}</p>
          )}
          {!editing && timeAgo && (
            <p className="mt-1 text-xs text-muted-foreground">{timeAgo}</p>
          )}
        </div>

        <div className="flex flex-shrink-0 items-center gap-1">
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
                onClick={() => setConvertOpen(true)}
              >
                <ArrowRight className="h-3.5 w-3.5" />
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
                onClick={() => deleteMutation.mutate({ id: item.id })}
                disabled={deleteMutation.isPending}
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            </>
          )}
        </div>
      </div>

      <ConvertToTodoDialog item={item} open={convertOpen} onOpenChange={setConvertOpen} />
    </>
  )
}
