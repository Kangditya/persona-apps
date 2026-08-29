# Shared UI foundation

`@persona-apps/ui` owns framework- and domain-agnostic presentational
components. Applications import from the package root; they must not import
deep source paths or add duplicate primitive sets.

The source follows shadcn/ui's editable `new-york` and Tailwind v4 conventions.
`components.json` is configured for future generation from `packages/ui`:

```bash
pnpm dlx shadcn@latest add card -c packages/ui
```

Generated patterns remain editable source in this package. The current
components use Tailwind and `class-variance-authority`, `clsx`, and
`tailwind-merge` for composition. No Radix behavior is installed.

Accessibility ownership is deliberately single-source:

- native HTML owns `Button`, `Input`, `Label`, `Textarea`, `Field`, and simple
  semantic atoms;
- React Aria Components owns keyboard navigation, focus management, Escape
  dismissal, and ARIA behavior for `Dialog`, `DropdownMenu`, and `Sheet`;
- `PwaNotice` is presentational; applications own service-worker lifecycle.

Icon convention:

- use `@radix-ui/react-icons` for compact utility and control icons;
- use `@phosphor-icons/react` for semantic, status, navigation, and qurban visuals;
- import named icons directly for tree-shaking and size them with Tailwind (`size-3.5`, `size-4`, `size-5`, or `size-6`);
- hide decorative icons with `aria-hidden="true"`; icon-only controls must retain an accessible name from their control or `aria-label`.

Add a component by classifying it as an atom, molecule, or pattern, writing its
accessible API and behavior owner, adding a root export, and adding a focused
test from an application package. Keep API access, routing, authentication,
environment configuration, and qurban vocabulary out of this package.
