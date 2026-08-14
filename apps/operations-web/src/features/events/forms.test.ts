import { describe, expect, it } from "vitest";

import {
  eventTransitions,
  exactInteger,
  FormValueError,
  localDateTimeToUTC,
  maximumCapacity,
  maximumSafeInteger,
  offeringTransitions,
  utcToLocalDateTime,
} from "./forms";
import { operationsQueryKeys } from "./queries";

describe("Operations Event and Offering form boundaries", () => {
  it("parses exact integer bounds without floating-point syntax", () => {
    expect(
      exactInteger(
        "quota",
        "Quota",
        String(maximumSafeInteger),
        0,
        maximumSafeInteger,
      ),
    ).toBe(maximumSafeInteger);
    expect(
      exactInteger(
        "capacity",
        "Capacity",
        String(maximumCapacity),
        1,
        maximumCapacity,
      ),
    ).toBe(maximumCapacity);
    expect(() =>
      exactInteger("price", "Price", "1.5", 0, maximumSafeInteger),
    ).toThrow(FormValueError);
    expect(() =>
      exactInteger("quota", "Quota", "9007199254740992", 0, maximumSafeInteger),
    ).toThrow(FormValueError);
  });

  it("round-trips valid browser-local minutes and rejects invalid calendar values", () => {
    const local = "2026-08-14T09:30";
    const utc = localDateTimeToUTC("opens", "Registration opens", local);
    expect(utc).toMatch(/Z$/);
    expect(utcToLocalDateTime(utc)).toBe(local);
    expect(() =>
      localDateTimeToUTC("opens", "Registration opens", "2026-02-30T09:30"),
    ).toThrow(FormValueError);
  });

  it("maps only locally plausible transitions and stable hierarchical keys", () => {
    expect(eventTransitions("ACTIVE")).toEqual(["suspend", "close"]);
    expect(eventTransitions("ARCHIVED")).toEqual([]);
    expect(offeringTransitions("UNAVAILABLE")).toEqual(["publish", "archive"]);
    expect(operationsQueryKeys.eventDetail("event-1")).toEqual([
      "operations",
      "private",
      "events",
      "detail",
      "event-1",
    ]);
    expect(operationsQueryKeys.offeringList("event-1", "cursor-1")).toEqual([
      "operations",
      "private",
      "events",
      "event-1",
      "offerings",
      "list",
      "cursor-1",
    ]);
  });
});
