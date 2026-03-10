import { createConnectTransport } from "@connectrpc/connect-web"
import { getToken } from "@/lib/auth"

export const transport = createConnectTransport({
  baseUrl: "/api/v1",
  interceptors: [
    (next) => async (req) => {
      const token = getToken()
      if (token) {
        req.header.set("Authorization", `Bearer ${token}`)
      }
      return next(req)
    },
  ],
})
