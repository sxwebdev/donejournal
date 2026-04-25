import { Link, useRouterState } from "@tanstack/react-router"
import {
  Inbox,
  CheckSquare,
  Calendar,
  FileText,
  FolderOpen,
  Tag,
  LogOut,
  Sun,
  Moon,
  Monitor,
} from "lucide-react"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarRail,
  useSidebar,
} from "@/components/ui/sidebar"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useAuth } from "@/hooks/use-auth"
import { useTheme } from "@/components/theme-provider"

const themeOrder = ["system", "light", "dark"] as const
const themeIcons = { light: Sun, dark: Moon, system: Monitor } as const

function ThemeToggle() {
  const { theme, setTheme } = useTheme()
  const nextTheme =
    themeOrder[(themeOrder.indexOf(theme) + 1) % themeOrder.length]
  const Icon = themeIcons[theme]
  return (
    <button
      onClick={() => setTheme(nextTheme)}
      className="flex w-full items-center gap-3 rounded-lg px-2 py-2 text-sm transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
      title="Toggle theme"
    >
      <Icon className="h-4 w-4" />
      <span className="text-muted-foreground capitalize">{theme}</span>
    </button>
  )
}

const navItems = [
  { to: "/inbox", label: "Inbox", icon: Inbox },
  { to: "/todos", label: "Todos", icon: CheckSquare },
  { to: "/notes", label: "Notes", icon: FileText },
  { to: "/calendar", label: "Calendar", icon: Calendar },
  { to: "/workspaces", label: "Workspaces", icon: FolderOpen },
  { to: "/tags", label: "Tags", icon: Tag },
] as const

export function AppSidebar() {
  const { user, logout } = useAuth()
  const { setOpenMobile, state } = useSidebar()
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  const initials = user
    ? `${user.firstName.charAt(0)}${user.lastName.charAt(0) || ""}`.toUpperCase()
    : "?"

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader className="px-4 py-4">
        <div className="flex items-center gap-2">
          <CheckSquare className="h-5 w-5 text-primary" />
          <span className="font-semibold">
            {state === "collapsed" ? "DJ" : "DoneJournal"}
          </span>
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarMenu>
            {navItems.map(({ to, label, icon: Icon }) => (
              <SidebarMenuItem key={to}>
                <SidebarMenuButton
                  render={<Link to={to} onClick={() => setOpenMobile(false)} />}
                  isActive={pathname.startsWith(to)}
                >
                  <Icon className="h-4 w-4" />
                  <span>{label}</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            ))}
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="border-t border-sidebar-border">
        <ThemeToggle />
        <DropdownMenu>
          <DropdownMenuTrigger className="flex w-full items-center gap-3 rounded-lg border-0 bg-transparent px-2 py-2 text-left text-sm transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground">
            <Avatar className="h-8 w-8">
              {user?.photoUrl && (
                <AvatarImage src={user.photoUrl} alt={user.firstName} />
              )}
              <AvatarFallback className="text-xs">{initials}</AvatarFallback>
            </Avatar>
            <div className="min-w-0 flex-1 text-left">
              <p className="truncate leading-none font-medium">
                {user
                  ? `${user.firstName} ${user.lastName}`.trim()
                  : "Loading..."}
              </p>
              {user?.username && (
                <p className="mt-1.5 truncate text-xs text-muted-foreground">
                  @{user.username}
                </p>
              )}
            </div>
          </DropdownMenuTrigger>
          <DropdownMenuContent side="top" align="start" className="w-48">
            <DropdownMenuItem
              onClick={logout}
              className="text-destructive focus:text-destructive"
            >
              <LogOut className="mr-2 h-4 w-4" />
              Logout
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
