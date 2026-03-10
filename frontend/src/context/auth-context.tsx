import * as React from "react"
import { createClient } from "@connectrpc/connect"
import { transport } from "@/api/client"
import { AuthService } from "@/api/gen/donejournal/auth/v1/auth_pb"
import type { TelegramLoginData, User } from "@/api/gen/donejournal/auth/v1/auth_pb"
import { clearToken, getToken, setToken } from "@/lib/auth"

type AuthContextValue = {
  user: User | null
  isLoading: boolean
  login: (telegramData: TelegramLoginData) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = React.createContext<AuthContextValue | undefined>(undefined)

const authClient = createClient(AuthService, transport)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = React.useState<User | null>(null)
  const [isLoading, setIsLoading] = React.useState(true)

  React.useEffect(() => {
    const token = getToken()
    if (!token) {
      setIsLoading(false)
      return
    }

    authClient
      .getCurrentUser({})
      .then((res) => {
        setUser(res.user ?? null)
      })
      .catch(() => {
        clearToken()
      })
      .finally(() => {
        setIsLoading(false)
      })
  }, [])

  const login = React.useCallback(async (telegramData: TelegramLoginData) => {
    const res = await authClient.loginWithTelegram({ telegramData })
    setToken(res.accessToken)
    setUser(res.user ?? null)
  }, [])

  const logoutFn = React.useCallback(async () => {
    try {
      await authClient.logout({})
    } finally {
      clearToken()
      setUser(null)
    }
  }, [])

  const value = React.useMemo(
    () => ({ user, isLoading, login, logout: logoutFn }),
    [user, isLoading, login, logoutFn]
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = React.useContext(AuthContext)
  if (!ctx) throw new Error("useAuth must be used within AuthProvider")
  return ctx
}
