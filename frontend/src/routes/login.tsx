import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useEffect, useState } from "react"
import { TelegramLoginButton } from "@/components/auth/telegram-login-button"
import type { TelegramLoginData } from "@/api/gen/donejournal/auth/v1/auth_pb"
import { useAuth } from "@/hooks/use-auth"
import { CheckSquare } from "lucide-react"

export const Route = createFileRoute("/login")({
  component: LoginPage,
})

function LoginPage() {
  const { user, login } = useAuth()
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)

  // Redirect if already authenticated
  useEffect(() => {
    if (user) {
      navigate({ to: "/inbox" })
    }
  }, [user, navigate])

  const handleAuth = async (data: TelegramLoginData) => {
    try {
      setError(null)
      await login(data)
      // Navigation will happen via the useEffect above when user state updates
    } catch {
      setError("Authentication failed. Please try again.")
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <div className="w-full max-w-sm space-y-8 text-center">
        <div className="space-y-3">
          <div className="flex justify-center">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary">
              <CheckSquare className="h-8 w-8 text-primary-foreground" />
            </div>
          </div>
          <h1 className="text-3xl font-bold tracking-tight">DoneJournal</h1>
          <p className="text-muted-foreground">Your Telegram task journal</p>
        </div>

        <div className="rounded-xl border bg-card p-6 shadow-sm">
          <p className="mb-4 text-sm text-muted-foreground">
            Sign in with your Telegram account to continue
          </p>
          <div className="flex justify-center">
            <TelegramLoginButton onAuth={handleAuth} />
          </div>
          {error && (
            <p className="mt-3 text-sm text-destructive">{error}</p>
          )}
        </div>
      </div>
    </div>
  )
}
