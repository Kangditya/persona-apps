import { describe, expect, it } from "vitest";

import { isActivePath, offeringPath, paths } from "./paths";

describe("storefront paths", () => {
  it("keeps the storefront routes centralized", () => {
    expect(Object.values(paths)).toEqual([
      "/",
      "/offerings",
      "/purchase-tracking",
    ]);
    expect(offeringPath("offering/id")).toBe("/offerings/offering%2Fid");
  });

  it("marks exact routes and their dynamic children active", () => {
    expect(isActivePath("/", paths.home)).toBe(true);
    expect(isActivePath("/offerings/example", paths.offerings)).toBe(true);
    expect(isActivePath("/offerings", paths.home)).toBe(false);
    expect(isActivePath("/purchase-tracking", paths.offerings)).toBe(false);
  });
});
