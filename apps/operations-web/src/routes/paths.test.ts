import { describe, expect, it } from "vitest";

import { eventPath, offeringPath, paths, routePatterns } from "./paths";

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
    expect(routePatterns.event).toBe("/events/:eventId");
    expect(routePatterns.offering).toBe(
      "/events/:eventId/offerings/:offeringId",
    );
    expect(eventPath("event/id")).toBe("/events/event%2Fid");
    expect(offeringPath("event/id", "offering id")).toBe(
      "/events/event%2Fid/offerings/offering%20id",
    );
  });
});
