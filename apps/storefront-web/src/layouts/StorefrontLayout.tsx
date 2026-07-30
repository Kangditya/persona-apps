import { NavLink, Outlet } from "react-router";

import { paths } from "../routes/paths";

const links = [
  [paths.home, "Home"],
  [paths.products, "Products"],
  [paths.cart, "Cart"],
] as const;

export function StorefrontLayout() {
  return (
    <div className="min-h-screen bg-amber-50 text-stone-900">
      <header className="border-b border-amber-200 bg-white">
        <nav
          aria-label="Storefront navigation"
          className="mx-auto flex max-w-5xl items-center gap-6 px-6 py-4"
        >
          <strong className="mr-auto">Persona Storefront</strong>
          {links.map(([to, label]) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                isActive
                  ? "text-amber-700"
                  : "text-stone-600 hover:text-stone-950"
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
