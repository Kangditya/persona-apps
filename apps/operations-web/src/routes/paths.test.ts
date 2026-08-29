import { describe, expect, it } from "vitest";

import { eventPath, isActivePath, offeringPath, paths } from "./paths";

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

  it("builds Event and Offering detail paths safely", () => {
    expect(eventPath("event/id")).toBe("/events/event%2Fid");
    expect(offeringPath("event/id", "offering id")).toBe(
      "/events/event%2Fid/offerings/offering%20id",
    );
  });

  it("marks exact routes and their dynamic children active", () => {
    expect(isActivePath("/events", paths.events)).toBe(true);
    expect(isActivePath("/events/event-id", paths.events)).toBe(true);
    expect(isActivePath("/event-dashboard", paths.events)).toBe(false);
  });
});
