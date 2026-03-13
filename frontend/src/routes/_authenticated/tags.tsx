import { createFileRoute } from "@tanstack/react-router"
import { TagManager } from "@/components/tags/tag-manager"

export const Route = createFileRoute("/_authenticated/tags")({
  component: TagsPage,
})

function TagsPage() {
  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Tags</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Manage your tags and labels
        </p>
      </div>
      <TagManager />
    </div>
  )
}
