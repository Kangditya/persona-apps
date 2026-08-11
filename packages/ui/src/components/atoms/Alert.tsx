import type { HTMLAttributes } from "react";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "../../lib/cn";

const alertVariants = cva("relative w-full rounded-lg border p-4 text-sm", {
    variants: {
        variant: {
            default: "bg-background text-foreground",
            destructive:
                "border-destructive/50 text-destructive dark:border-destructive",
        },
    },
    defaultVariants: { variant: "default" },
});

export type AlertProps = HTMLAttributes<HTMLDivElement> &
    VariantProps<typeof alertVariants> & { role?: "alert" | "status" };

export function Alert({
    className,
    variant,
    role = "status",
    ...props
}: AlertProps) {
    return (
        <div
            data-slot="alert"
            role={role}
            className={cn(alertVariants({ variant }), className)}
            {...props}
        />
    );
}
