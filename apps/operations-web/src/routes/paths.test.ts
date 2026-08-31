import { describe, expect, it } from "vitest";

import {
  eventPath,
  isActivePath,
  offeringPath,
  paths,
  purchasePath,
} from "./paths";

describe("operations paths", () => {
  it("keeps the qurban placeholder routes centralized", () => {
    expect(Object.values(paths)).toEqual([
      "/operator-login",
      "/events",
      "/event-dashboard",
      "/purchasing",
      "/payment-verification",
    ]);
  });

  it("builds Event, Offering, and Purchase detail paths safely", () => {
    expect(eventPath("event/id")).toBe("/events/event%2Fid");
    expect(offeringPath("event/id", "offering id")).toBe(
      "/events/event%2Fid/offerings/offering%20id",
    );
    expect(purchasePath("purchase/id")).toBe("/purchasing/purchase%2Fid");
  });

  it("marks exact routes and their dynamic children active", () => {
    expect(isActivePath("/events", paths.events)).toBe(true);
    expect(isActivePath("/events/event-id", paths.events)).toBe(true);
    expect(isActivePath("/event-dashboard", paths.events)).toBe(false);
    expect(isActivePath("/purchasing/purchase-id", paths.purchasing)).toBe(
      true,
    );
  });
});
