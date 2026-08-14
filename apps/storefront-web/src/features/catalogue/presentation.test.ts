import { ApiError } from "@persona-apps/api-client";
import { describe, expect, it } from "vitest";

import type { PublicEvent } from "../../api/catalogue";
import {
  availabilityText,
  exactMinorPrice,
  publicErrorPresentation,
  registrationPresentation,
} from "./presentation";

const event: PublicEvent = {
  id: "11111111-1111-1111-1111-111111111111",
  eventYear: 2027,
  name: "Qurban 2027",
  status: "ACTIVE",
  registrationOpensAt: "2027-05-01T00:00:00Z",
  registrationClosesAt: "2027-06-01T00:00:00Z",
};

describe("catalogue presentation", () => {
  it("presents registration boundaries in an explicit timezone", () => {
    expect(
      registrationPresentation(
        event,
        Date.parse("2027-04-30T23:59:59Z"),
        "UTC",
      ),
    ).toMatchObject({ label: "Registration upcoming", timeZone: "UTC" });
    expect(
      registrationPresentation(event, Date.parse("2027-05-01T00:00:00Z"), "UTC")
        .label,
    ).toBe("Registration open");
    expect(
      registrationPresentation(event, Date.parse("2027-06-01T00:00:00Z"), "UTC")
        .label,
    ).toBe("Registration closed");
    expect(
      registrationPresentation(
        {
          ...event,
          registrationOpensAt: null,
          registrationClosesAt: null,
        },
        Date.now(),
        "UTC",
      ),
    ).toEqual({
      label: "Registration schedule not configured",
      details: ["No public registration window has been configured."],
      timeZone: null,
    });
  });

  it("keeps price and advisory availability exact", () => {
    expect(exactMinorPrice("IDR", Number.MAX_SAFE_INTEGER)).toBe(
      "IDR 9007199254740991 minor units",
    );
    expect(availabilityText(7)).toBe(
      "7 participant units available (advisory snapshot).",
    );
    expect(availabilityText(0)).toBe(
      "Currently unavailable (advisory snapshot).",
    );
    expect(availabilityText(null)).toBe(
      "No finite quota is configured; availability is still advisory.",
    );
  });

  it("uses safe public errors without automatic rate-limit retry", () => {
    expect(
      publicErrorPresentation(
        new ApiError("rate_limited", "raw server text", 429, {
          requestId: "request_1",
          retryAfterSeconds: 1,
        }),
      ),
    ).toEqual({
      message: "Too many catalogue requests. Try again in 1 second.",
      requestId: "request_1",
      retryable: false,
    });
    expect(
      publicErrorPresentation(
        new ApiError("service_unavailable", "database password", 503),
      ),
    ).toMatchObject({
      message: "The public catalogue is temporarily unavailable.",
      retryable: true,
    });
  });
});
