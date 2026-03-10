import { useCallback, useRef } from "react"
import { useQuery } from "@connectrpc/connect-query"
import { create } from "@bufbuild/protobuf"
import { listInboxItems } from "@/api/gen/donejournal/inbox/v1/inbox-InboxService_connectquery"
import { SubscribeInboxRequestSchema } from "@/api/gen/donejournal/inbox/v1/inbox_pb"
import { inboxClient } from "@/api/client"
import { useSubscriptionRefetch } from "@/hooks/use-subscription-refetch"
import { InboxCard } from "./inbox-card"
import { Skeleton } from "@/components/ui/skeleton"
import { Inbox } from "lucide-react"

export function InboxList() {
  const query = useQuery(listInboxItems, { pageSize: 50, pageToken: "" })

  const subRef = useRef<{ abort: () => void } | null>(null)
  const subscribe = useCallback(
    (signal: AbortSignal) =>
      inboxClient.subscribeInbox(create(SubscribeInboxRequestSchema), { signal }),
    [],
  )
  useSubscriptionRefetch({ refetch: query.refetch, subscribe, ref: subRef })

  const { data, isLoading } = query

  if (isLoading) {
    return (
      <div className="space-y-3">
        {[...Array(4)].map((_, i) => (
          <Skeleton key={i} className="h-20 w-full rounded-xl" />
        ))}
      </div>
    )
  }

  const items = data?.items ?? []

  if (!items.length) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <Inbox className="mb-3 h-10 w-10 text-muted-foreground/50" />
        <p className="font-medium text-muted-foreground">Your inbox is empty</p>
        <p className="mt-1 text-sm text-muted-foreground">Start capturing ideas above</p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {items.map((item) => (
        <InboxCard key={item.id} item={item} />
      ))}
    </div>
  )
}
