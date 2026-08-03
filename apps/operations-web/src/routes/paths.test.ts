import { describe, expect, it } from "vitest";

import { paths } from "./paths";

describe("operations paths", () => {
  it("keeps the qurban placeholder routes centralized", () => {
    expect(Object.values(paths)).toEqual([
      "/operator-login",
      "/event-dashboard",
      "/purchasing",
      "/payment-verification",
    ]);
  });
});
