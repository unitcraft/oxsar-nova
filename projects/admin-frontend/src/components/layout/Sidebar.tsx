// Sidebar с группировкой пунктов и collapse-режимом (240px / 56px).
// Skin minimalist: line-icons (lucide-react), 32px row height.
//
// Permission-aware: пункты с requiredPermission скрываются, если у юзера
// permission нет. Это UX-fix; backend всё равно перепроверяет (defence
// in depth).
import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  Users,
  ShieldCheck,
  CreditCard,
  Gamepad2,
  Flag,
  ScrollText,
  Settings,
  ChevronsLeft,
  ChevronsRight,
  type LucideIcon,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useAuth } from '@/store/auth';
import { useUiStore } from '@/store/ui';

interface NavItem {
  to: string;
  label: string;
  icon: LucideIcon;
  requiredPermission?: string;
}

interface NavGroup {
  title: string;
  items: NavItem[];
}

const navGroups: NavGroup[] = [
  {
    title: 'overview',
    items: [{ to: '/', label: 'Dashboard', icon: LayoutDashboard }],
  },
  {
    title: 'identity',
    items: [
      { to: '/users', label: 'Users', icon: Users, requiredPermission: 'users:read' },
      { to: '/roles', label: 'Roles', icon: ShieldCheck, requiredPermission: 'users:read' },
      { to: '/audit', label: 'Audit log', icon: ScrollText, requiredPermission: 'audit:read' },
    ],
  },
  {
    title: 'operations',
    items: [
      { to: '/billing', label: 'Billing', icon: CreditCard, requiredPermission: 'billing:read' },
      { to: '/game-ops/events', label: 'Game-ops', icon: Gamepad2, requiredPermission: 'game:events:retry' },
      // План 56 Ф.6: UGC-жалобы (149-ФЗ). Permission moderation:reports:read
      // придёт через identity план 52 Ф.X; до тех пор пункт скрыт от
      // не-admin'ов через Reports.tsx (role-based fallback в самом
      // route-компоненте), а в sidebar пункт виден всем — это известное
      // ограничение, исчезнет с подключением permissions в JWT.
      { to: '/reports', label: 'Жалобы', icon: Flag },
    ],
  },
  {
    title: 'system',
    items: [{ to: '/settings', label: 'Settings', icon: Settings }],
  },
];

export function Sidebar(): React.ReactElement {
  const collapsed = useUiStore((s) => s.sidebarCollapsed);
  const toggle = useUiStore((s) => s.toggleSidebar);
  const hasPermission = useAuth((s) => s.hasPermission);

  return (
    <aside
      className={cn(
        'border-r bg-card text-card-foreground transition-[width] duration-150',
        collapsed ? 'w-sidebar-collapsed' : 'w-sidebar',
      )}
    >
      <div className="flex h-topbar items-center justify-between border-b px-3">
        {!collapsed && (
          <span className="text-xs font-semibold tracking-wide text-muted-foreground">
            navigation
          </span>
        )}
        <button
          type="button"
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          onClick={toggle}
          className="rounded-md p-1.5 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
        >
          {collapsed ? (
            <ChevronsRight className="h-4 w-4" aria-hidden="true" />
          ) : (
            <ChevronsLeft className="h-4 w-4" aria-hidden="true" />
          )}
        </button>
      </div>

      <nav className="space-y-4 p-2">
        {navGroups.map((group) => {
          const visible = group.items.filter(
            (item) => !item.requiredPermission || hasPermission(item.requiredPermission),
          );
          if (visible.length === 0) return null;
          return (
            <div key={group.title}>
              {!collapsed && (
                <p className="mb-1 px-2 text-2xs font-medium uppercase tracking-wide text-muted-foreground">
                  {group.title}
                </p>
              )}
              <ul className="space-y-0.5">
                {visible.map((item) => (
                  <li key={item.to}>
                    <NavLink
                      to={item.to}
                      end={item.to === '/'}
                      title={collapsed ? item.label : undefined}
                      className={({ isActive }) =>
                        cn(
                          'flex h-8 items-center gap-2 rounded-md px-2 text-sm transition-colors',
                          isActive
                            ? 'bg-accent text-accent-foreground'
                            : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground',
                        )
                      }
                    >
                      <item.icon className="h-4 w-4 flex-shrink-0" aria-hidden="true" />
                      {!collapsed && <span className="truncate">{item.label}</span>}
                    </NavLink>
                  </li>
                ))}
              </ul>
            </div>
          );
        })}
      </nav>
    </aside>
  );
}
