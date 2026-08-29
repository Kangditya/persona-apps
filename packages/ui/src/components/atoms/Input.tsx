import type { InputHTMLAttributes } from "react";

import { cn } from "../../lib/cn";

export type InputProps = InputHTMLAttributes<HTMLInputElement>;

export function Input({
    className,
    "aria-invalid": invalid,
    ...props
}: InputProps) {
    return (
        <input
            data-slot="input"
            className={cn(
                "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground shadow-sm outline-none placeholder:text-muted-foreground focus-visible:border-ring disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-destructive",
                className,
            )}
            aria-invalid={invalid}
            {...props}
        />
    );
}
