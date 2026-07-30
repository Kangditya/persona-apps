import { describe, expect, it } from "vitest";

import { paths } from "./paths";

describe("storefront paths", () => {
  it("keeps the bootstrap routes stable", () => {
    expect(Object.values(paths)).toEqual(["/", "/products", "/cart"]);
  });
});
