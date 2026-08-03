import { NavLink, Outlet } from "react-router";

import { paths } from "../routes/paths";

const links = [
  [paths.operatorLogin, "Operator Login"],
  [paths.eventDashboard, "Event Dashboard"],
  [paths.purchasing, "Purchasing"],
  [paths.paymentVerification, "Payment Verification"],
] as const;

export function OperationsLayout() {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      <header className="border-b border-slate-800 bg-slate-900">
        <nav
          aria-label="Operations navigation"
          className="mx-auto flex max-w-5xl items-center gap-6 px-6 py-4"
        >
          <strong className="mr-auto">Qurban Operations</strong>
          {links.map(([to, label]) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                isActive ? "text-cyan-300" : "text-slate-300 hover:text-white"
              }
            >
              {label}
            </NavLink>
          ))}
        </nav>
      </header>
      <main className="mx-auto max-w-5xl px-6 py-16">
        <Outlet />
      </main>
    </div>
  );
}
