import { useRouterState } from "@tanstack/react-router"
import { SidebarTrigger } from "@/components/ui/sidebar"

const pageTitles: Record<string, string> = {
  "/inbox": "Inbox",
  "/todos": "Todos",
  "/calendar": "Calendar",
}

export function Header() {
  const pathname = useRouterState({ select: (s) => s.location.pathname })
  const title = pageTitles[pathname] ?? "DoneJournal"

  return (
    <header className="flex h-12 items-center gap-3 border-b px-4 md:hidden">
      <SidebarTrigger />
      <span className="font-medium">{title}</span>
    </header>
  )
}
