import { describe, expect, it } from "vitest";

import { paths } from "./paths";

describe("storefront paths", () => {
  it("keeps the qurban placeholder routes centralized", () => {
    expect(Object.values(paths)).toEqual([
      "/",
      "/offerings",
      "/purchase-tracking",
    ]);
  });
});
