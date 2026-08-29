import { describe, expect, it } from "vitest";

import { createQueryClient } from "./queryClient";

describe("Operations query provider", () => {
  it("creates isolated clients with bounded default retry", () => {
    const first = createQueryClient();
    const second = createQueryClient();
    expect(first).not.toBe(second);
    expect(first.getDefaultOptions().queries?.retry).toBe(false);
  });
});
