import { createRoot } from "react-dom/client"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { TransportProvider } from "@connectrpc/connect-query"
import { RouterProvider } from "@tanstack/react-router"
import { ConnectError, Code } from "@connectrpc/connect"

import "./index.css"
import { ThemeProvider } from "@/components/theme-provider"
import { AuthProvider, useAuth } from "@/context/auth-context"
import { transport } from "@/api/client"
import { router } from "@/router"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 30,
      retry: (count, err) => {
        if (err instanceof ConnectError && err.code === Code.Unauthenticated) {
          return false
        }
        return count < 2
      },
    },
  },
})

function AppRouter() {
  const auth = useAuth()
  return <RouterProvider router={router} context={{ auth }} />
}

createRoot(document.getElementById("root")!).render(
  <ThemeProvider>
    <QueryClientProvider client={queryClient}>
      <TransportProvider transport={transport}>
        <AuthProvider>
          <AppRouter />
        </AuthProvider>
      </TransportProvider>
    </QueryClientProvider>
  </ThemeProvider>
)
