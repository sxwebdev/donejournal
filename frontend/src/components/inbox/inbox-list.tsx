import { useQuery } from "@connectrpc/connect-query"
import { listInboxItems } from "@/api/gen/donejournal/inbox/v1/inbox-InboxService_connectquery"
import { InboxCard } from "./inbox-card"
import { Skeleton } from "@/components/ui/skeleton"
import { Inbox } from "lucide-react"

export function InboxList() {
  const { data, isLoading } = useQuery(listInboxItems, { pageSize: 50, pageToken: "" })

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
