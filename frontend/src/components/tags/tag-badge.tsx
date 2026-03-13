import { X } from "lucide-react"
import { cn } from "@/lib/utils"

type Props = {
  name: string
  color: string
  onRemove?: () => void
  className?: string
}

function contrastColor(hex: string): string {
  const r = parseInt(hex.slice(1, 3), 16)
  const g = parseInt(hex.slice(3, 5), 16)
  const b = parseInt(hex.slice(5, 7), 16)
  return r * 0.299 + g * 0.587 + b * 0.114 > 150 ? "#000" : "#fff"
}

export function TagBadge({ name, color, onRemove, className }: Props) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-0.5 rounded-full px-2 py-0.5 text-[10px] font-medium leading-none",
        className
      )}
      style={{ backgroundColor: color, color: contrastColor(color) }}
    >
      {name}
      {onRemove && (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            onRemove()
          }}
          className="ml-0.5 rounded-full p-0.5 transition-opacity hover:opacity-70"
        >
          <X className="h-2.5 w-2.5" />
        </button>
      )}
    </span>
  )
}
