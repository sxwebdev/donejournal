import { createRouter } from "@tanstack/react-router"
import { routeTree } from "./routeTree.gen"
import type { useAuth } from "@/hooks/use-auth"

type RouterContext = {
  auth: ReturnType<typeof useAuth>
}

export const router = createRouter({
  routeTree,
  context: {
    auth: undefined!,
  } as RouterContext,
  defaultPreload: "intent",
})

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router
  }
}
