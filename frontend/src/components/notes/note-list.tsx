import { useCallback, useRef } from "react"
import { useQuery } from "@connectrpc/connect-query"
import { create } from "@bufbuild/protobuf"
import { listNotes } from "@/api/gen/donejournal/notes/v1/notes-NoteService_connectquery"
import { SubscribeNotesRequestSchema } from "@/api/gen/donejournal/notes/v1/notes_pb"
import { notesClient } from "@/api/client"
import { useSubscriptionRefetch } from "@/hooks/use-subscription-refetch"
import { NoteCard } from "./note-card"
import { Skeleton } from "@/components/ui/skeleton"
import { FileText } from "lucide-react"

type Props = {
  search?: string
  projectId?: string
}

export function NoteList({ search, projectId }: Props) {
  const query = useQuery(listNotes, {
    pageSize: 100,
    search: search || undefined,
    projectId,
  })

  const subRef = useRef<{ abort: () => void } | null>(null)
  const subscribe = useCallback(
    (signal: AbortSignal) =>
      notesClient.subscribeNotes(create(SubscribeNotesRequestSchema), {
        signal,
      }),
    []
  )
  useSubscriptionRefetch({ refetch: query.refetch, subscribe, ref: subRef })

  const { data, isLoading } = query

  if (isLoading) {
    return (
      <div className="space-y-2">
        {[...Array(5)].map((_, i) => (
          <Skeleton key={i} className="h-20 w-full rounded-lg" />
        ))}
      </div>
    )
  }

  const notes = data?.notes ?? []

  if (!notes.length) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <FileText className="mb-3 h-10 w-10 text-muted-foreground/50" />
        <p className="font-medium text-muted-foreground">No notes found</p>
        <p className="mt-1 text-sm text-muted-foreground">
          Create one above or adjust your search
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {notes.map((note) => (
        <NoteCard key={note.id} note={note} />
      ))}
    </div>
  )
}
