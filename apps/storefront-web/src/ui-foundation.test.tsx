import { describe, expect, it } from "vitest";

import { Button, Card, Field, Input, cn } from "@persona-apps/ui";

describe("shared UI foundation", () => {
  it("resolves stable workspace exports and merges Tailwind classes", () => {
    expect(cn("px-2", "px-4")).toBe("px-4");
    expect(Button({ children: "Save" }).props["data-slot"]).toBe("button");
    expect(Card({ children: "Content" }).props["data-slot"]).toBe("card");
  });

  it("keeps field labels and validation descriptions explicit", () => {
    const field = Field({
      id: "email",
      label: "Email",
      description: "We will not share it.",
      error: "Enter a valid email.",
      children: Input({}),
    });
    const children = field.props.children as Array<{
      props?: Record<string, unknown>;
    }>;
    const input = children[1];

    expect(input.props?.id).toBe("email");
    expect(input.props?.["aria-invalid"]).toBe(true);
    expect(input.props?.["aria-describedby"]).toBe(
      "email-description email-error",
    );
  });
});
