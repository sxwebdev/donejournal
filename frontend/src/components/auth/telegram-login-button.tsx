import { useEffect, useRef } from "react"
import type { TelegramLoginData } from "@/api/gen/donejournal/auth/v1/auth_pb"

// Raw data from the Telegram widget (uses snake_case and number types)
type TelegramWidgetData = {
  id: number
  first_name: string
  last_name?: string
  username?: string
  photo_url?: string
  auth_date: number
  hash: string
}

type Props = {
  onAuth: (data: TelegramLoginData) => void
}

declare global {
  interface Window {
    onTelegramAuth: (data: TelegramWidgetData) => void
  }
}

export function TelegramLoginButton({ onAuth }: Props) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const container = ref.current
    if (!container) return

    window.onTelegramAuth = (raw: TelegramWidgetData) => {
      onAuth({
        id: BigInt(raw.id),
        firstName: raw.first_name,
        lastName: raw.last_name ?? "",
        username: raw.username ?? "",
        photoUrl: raw.photo_url ?? "",
        authDate: BigInt(raw.auth_date),
        hash: raw.hash,
      } as unknown as TelegramLoginData)
    }

    const script = document.createElement("script")
    script.src = "https://telegram.org/js/telegram-widget.js?22"
    script.setAttribute(
      "data-telegram-login",
      import.meta.env.VITE_TELEGRAM_BOT_USERNAME ?? "your_bot"
    )
    script.setAttribute("data-size", "large")
    script.setAttribute("data-radius", "8")
    script.setAttribute("data-onauth", "onTelegramAuth(user)")
    script.async = true
    container.appendChild(script)

    return () => {
      if (container.contains(script)) {
        container.removeChild(script)
      }
    }
  }, [onAuth])

  return <div ref={ref} />
}
