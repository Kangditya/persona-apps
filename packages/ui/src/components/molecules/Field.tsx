import { cloneElement, type ReactElement, type ReactNode } from "react";

import { cn } from "../../lib/cn";
import { Label } from "../atoms/Label";

type FieldControlProps = {
  id?: string;
  required?: boolean;
  "aria-describedby"?: string;
  "aria-invalid"?: boolean | "false" | "true";
};

export type FieldProps = {
  id: string;
  label: ReactNode;
  description?: ReactNode;
  error?: ReactNode;
  required?: boolean;
  className?: string;
  children: ReactElement<FieldControlProps>;
};

export function Field({
  id,
  label,
  description,
  error,
  required,
  className,
  children,
}: FieldProps) {
  const descriptionId = description ? `${id}-description` : undefined;
  const errorId = error ? `${id}-error` : undefined;
  const describedBy =
    [children.props["aria-describedby"], descriptionId, errorId]
      .filter(Boolean)
      .join(" ") || undefined;

  return (
    <div data-slot="field" className={cn("grid gap-2", className)}>
      <Label htmlFor={id}>
        {label}
        {required ? (
          <span className="ml-1 text-destructive" aria-hidden="true">
            *
          </span>
        ) : null}
      </Label>
      {cloneElement(children, {
        id,
        "aria-describedby": describedBy,
        "aria-invalid": error ? true : children.props["aria-invalid"],
        required,
      })}
      {description ? (
        <p id={descriptionId} className="text-sm text-muted-foreground">
          {description}
        </p>
      ) : null}
      {error ? (
        <p id={errorId} role="alert" className="text-sm text-destructive">
          {error}
        </p>
      ) : null}
    </div>
  );
}
