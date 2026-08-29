import { ApiError } from "@persona-apps/api-client";

import type { PublicEvent } from "../../api/catalogue";

export type RegistrationPresentation = {
  label: string;
  details: readonly string[];
  timeZone: string | null;
};

export type PublicErrorPresentation = {
  message: string;
  requestId: string | null;
  retryable: boolean;
};

export function browserTimeZone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone || "Local time";
}

export function registrationPresentation(
  event: PublicEvent,
  now = Date.now(),
  timeZone = browserTimeZone(),
): RegistrationPresentation {
  const opens = event.registrationOpensAt
    ? Date.parse(event.registrationOpensAt)
    : null;
  const closes = event.registrationClosesAt
    ? Date.parse(event.registrationClosesAt)
    : null;
  if (opens === null && closes === null) {
    return {
      label: "Registration schedule not configured",
      details: ["No public registration window has been configured."],
      timeZone: null,
    };
  }
  const label =
    opens !== null && now < opens
      ? "Registration upcoming"
      : closes !== null && now >= closes
        ? "Registration closed"
        : "Registration open";
  return {
    label,
    details: [
      ...(event.registrationOpensAt
        ? [`Opens ${formatInstant(event.registrationOpensAt, timeZone)}`]
        : []),
      ...(event.registrationClosesAt
        ? [`Closes ${formatInstant(event.registrationClosesAt, timeZone)}`]
        : []),
    ],
    timeZone,
  };
}

export function exactMinorPrice(
  currencyCode: string,
  priceMinor: number,
): string {
  return `${currencyCode} ${String(priceMinor)} minor units`;
}

export function availabilityText(value: number | null): string {
  if (value === null) {
    return "No finite quota is configured; availability is still advisory.";
  }
  if (value === 0) {
    return "Currently unavailable (advisory snapshot).";
  }
  return `${String(value)} participant units available (advisory snapshot).`;
}

export function publicErrorPresentation(
  error: unknown,
): PublicErrorPresentation {
  if (!(error instanceof ApiError)) {
    return {
      message: "The public catalogue could not be loaded.",
      requestId: null,
      retryable: false,
    };
  }
  const message = (() => {
    switch (error.kind) {
      case "rate_limited":
        if (!error.retryAfterSeconds) {
          return "Too many catalogue requests. Try again shortly.";
        }
        return `Too many catalogue requests. Try again in ${String(error.retryAfterSeconds)} ${error.retryAfterSeconds === 1 ? "second" : "seconds"}.`;
      case "service_unavailable":
      case "unavailable":
        return "The public catalogue is temporarily unavailable.";
      case "not_found":
        return "The requested catalogue item is unavailable.";
      case "cancelled":
        return "The catalogue request was cancelled.";
      case "unknown":
        return "The catalogue response could not be read.";
      default:
        return "The public catalogue request could not be completed.";
    }
  })();
  return {
    message,
    requestId: error.requestId ?? null,
    retryable:
      error.kind === "service_unavailable" || error.kind === "unavailable",
  };
}

function formatInstant(value: string, timeZone: string): string {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: timeZone === "Local time" ? undefined : timeZone,
  }).format(new Date(value));
}
