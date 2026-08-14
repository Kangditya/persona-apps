import { describe, expect, it } from "vitest";

import { offeringPath, paths, routePatterns } from "./paths";

describe("storefront paths", () => {
  it("keeps the storefront routes centralized", () => {
    expect(Object.values(paths)).toEqual([
      "/",
      "/offerings",
      "/purchase-tracking",
    ]);
    expect(routePatterns.offering).toBe("/offerings/:offeringId");
    expect(offeringPath("offering/id")).toBe("/offerings/offering%2Fid");
  });
});
