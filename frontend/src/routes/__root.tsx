import { createRootRouteWithContext, Outlet } from "@tanstack/react-router"
import type { useAuth } from "@/hooks/use-auth"

type RouterContext = {
  auth: ReturnType<typeof useAuth>
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: () => <Outlet />,
})
