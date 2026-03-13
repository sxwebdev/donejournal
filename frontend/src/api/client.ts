import { createClient } from "@connectrpc/connect"
import { createConnectTransport } from "@connectrpc/connect-web"
import { getToken } from "@/lib/auth"
import { TodoService } from "@/api/gen/donejournal/todos/v1/todos_pb"
import { InboxService } from "@/api/gen/donejournal/inbox/v1/inbox_pb"
import { NoteService } from "@/api/gen/donejournal/notes/v1/notes_pb"
import { ProjectService } from "@/api/gen/donejournal/projects/v1/projects_pb"

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

export const todosClient = createClient(TodoService, transport)
export const inboxClient = createClient(InboxService, transport)
export const notesClient = createClient(NoteService, transport)
export const projectsClient = createClient(ProjectService, transport)
