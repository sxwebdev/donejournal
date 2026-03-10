import { createFileRoute } from "@tanstack/react-router"
import { InboxQuickAdd } from "@/components/inbox/inbox-quick-add"
import { InboxList } from "@/components/inbox/inbox-list"

export const Route = createFileRoute("/_authenticated/inbox")({
  component: InboxPage,
})

function InboxPage() {
  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Inbox</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Capture quick thoughts and ideas
        </p>
      </div>
      <InboxQuickAdd />
      <InboxList />
    </div>
  )
}
