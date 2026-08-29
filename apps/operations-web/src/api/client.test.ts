import { describe, expect, it } from "vitest";

import { createOperationsApi } from "./client";
import { operationsQueryKeys } from "./diagnostics";

describe("operations API boundary", () => {
  it("uses its own deterministic development provider", async () => {
    const api = createOperationsApi({
      baseUrl: "https://api.example.test",
      provider: "development",
      fetch: () => {
        throw new Error("development provider must not call fetch");
      },
    });

    await expect(api.diagnostics.health()).resolves.toEqual({
      status: "development",
    });
  });

  it("keeps operations query keys separate from storefront keys", () => {
    expect(operationsQueryKeys.health).toEqual([
      "operations",
      "diagnostics",
      "health",
    ]);
  });
});
