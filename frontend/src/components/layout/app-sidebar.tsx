import { Link, useRouterState } from "@tanstack/react-router"
import { Inbox, CheckSquare, Calendar, LogOut } from "lucide-react"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from "@/components/ui/sidebar"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useAuth } from "@/hooks/use-auth"

const navItems = [
  { to: "/inbox", label: "Inbox", icon: Inbox },
  { to: "/todos", label: "Todos", icon: CheckSquare },
  { to: "/calendar", label: "Calendar", icon: Calendar },
] as const

export function AppSidebar() {
  const { user, logout } = useAuth()
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  const initials = user
    ? `${user.firstName.charAt(0)}${user.lastName.charAt(0) || ""}`.toUpperCase()
    : "?"

  return (
    <Sidebar>
      <SidebarHeader className="px-4 py-4">
        <div className="flex items-center gap-2">
          <CheckSquare className="h-5 w-5 text-primary" />
          <span className="font-semibold">DoneJournal</span>
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarMenu>
          {navItems.map(({ to, label, icon: Icon }) => (
            <SidebarMenuItem key={to}>
              <SidebarMenuButton render={<Link to={to} />} isActive={pathname.startsWith(to)}>
                <Icon className="h-4 w-4" />
                <span>{label}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarContent>

      <SidebarFooter className="p-2">
        <DropdownMenu>
          <DropdownMenuTrigger className="flex w-full items-center gap-3 rounded-lg px-2 py-2 text-sm hover:bg-sidebar-accent hover:text-sidebar-accent-foreground transition-colors text-left bg-transparent border-0">
              <Avatar className="h-7 w-7">
                {user?.photoUrl && <AvatarImage src={user.photoUrl} alt={user.firstName} />}
                <AvatarFallback className="text-xs">{initials}</AvatarFallback>
              </Avatar>
              <div className="flex-1 min-w-0 text-left">
                <p className="truncate font-medium leading-none">
                  {user ? `${user.firstName} ${user.lastName}`.trim() : "Loading..."}
                </p>
                {user?.username && (
                  <p className="truncate text-xs text-muted-foreground">@{user.username}</p>
                )}
              </div>
          </DropdownMenuTrigger>
          <DropdownMenuContent side="top" align="start" className="w-48">
            <DropdownMenuItem onClick={logout} className="text-destructive focus:text-destructive">
              <LogOut className="mr-2 h-4 w-4" />
              Logout
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarFooter>
    </Sidebar>
  )
}
