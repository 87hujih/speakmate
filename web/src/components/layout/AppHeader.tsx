import { History, Home, MessageCircle } from "lucide-react";
import { NavLink } from "react-router-dom";
import { cn } from "../../utils/cn";

const navItems = [
  { label: "场景训练", to: "/", icon: Home, end: true },
  { label: "历史记录", to: "/history", icon: History },
];

export function AppHeader() {
  return (
    <header className="sticky top-0 z-30 flex min-h-[72px] flex-wrap items-center justify-between gap-3 border-b border-line/80 bg-slate-50/80 px-4 py-3 backdrop-blur-xl md:px-10">
      <NavLink to="/" className="flex items-center gap-3 text-[21px] font-black tracking-[-0.02em] text-ink">
        <span className="grid h-10 w-10 place-items-center rounded-2xl bg-gradient-to-br from-brand-blue to-brand-purple text-white shadow-[0_12px_26px_rgba(37,99,235,0.28)]">
          <MessageCircle className="h-5 w-5" />
        </span>
        SpeakMate
      </NavLink>

      <nav className="flex flex-wrap items-center gap-2">
        {navItems.map((item) => {
          const Icon = item.icon;
          return (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                cn(
                  "inline-flex h-10 items-center gap-2 rounded-full px-4 text-sm font-bold text-muted transition hover:bg-white hover:text-brand-blue hover:shadow-soft",
                  isActive && "bg-white text-brand-blue shadow-soft",
                )
              }
            >
              <Icon className="h-4 w-4" />
              {item.label}
            </NavLink>
          );
        })}
      </nav>
    </header>
  );
}
