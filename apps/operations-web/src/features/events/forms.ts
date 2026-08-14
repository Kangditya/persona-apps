import type {
  EventStatus,
  EventTransition,
  OfferingStatus,
  OfferingTransition,
} from "../../api/events";

export const maximumSafeInteger = 9_007_199_254_740_991;
export const maximumCapacity = 2_147_483_647;

export class FormValueError extends Error {
  readonly field: string;

  constructor(field: string, message: string) {
    super(message);
    this.name = "FormValueError";
    this.field = field;
  }
}

export function exactInteger(
  field: string,
  label: string,
  raw: FormDataEntryValue | null,
  minimum: number,
  maximum: number,
): number {
  const value = typeof raw === "string" ? raw.trim() : "";
  if (!/^\d+$/.test(value)) {
    throw new FormValueError(field, `${label} must be a whole number.`);
  }
  const parsed = Number(value);
  if (!Number.isSafeInteger(parsed) || parsed < minimum || parsed > maximum) {
    throw new FormValueError(
      field,
      `${label} must be between ${minimum} and ${maximum}.`,
    );
  }
  return parsed;
}

export function optionalExactInteger(
  field: string,
  label: string,
  raw: FormDataEntryValue | null,
  minimum = 0,
  maximum = maximumSafeInteger,
): number | null {
  if (typeof raw !== "string" || raw.trim() === "") return null;
  return exactInteger(field, label, raw, minimum, maximum);
}

export function localDateTimeToUTC(
  field: string,
  label: string,
  value: string,
): string | null {
  if (value === "") return null;
  const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})$/.exec(value);
  if (!match)
    throw new FormValueError(field, `${label} is not a valid local time.`);
  const [, year, month, day, hour, minute] = match.map(Number);
  const date = new Date(year, month - 1, day, hour, minute, 0, 0);
  if (
    date.getFullYear() !== year ||
    date.getMonth() !== month - 1 ||
    date.getDate() !== day ||
    date.getHours() !== hour ||
    date.getMinutes() !== minute
  ) {
    throw new FormValueError(
      field,
      `${label} does not exist in the browser timezone.`,
    );
  }
  return date.toISOString();
}

export function utcToLocalDateTime(value: string | null): string {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const part = (number: number) => String(number).padStart(2, "0");
  return `${date.getFullYear()}-${part(date.getMonth() + 1)}-${part(date.getDate())}T${part(date.getHours())}:${part(date.getMinutes())}`;
}

export function utcPreview(value: string): string {
  if (!value) return "Not set";
  try {
    return localDateTimeToUTC("datetime", "Date and time", value) ?? "Not set";
  } catch {
    return "Invalid local date and time";
  }
}

export function browserTimeZone(): string {
  return (
    Intl.DateTimeFormat().resolvedOptions().timeZone || "browser local time"
  );
}

export function eventTransitions(
  status: EventStatus,
): readonly EventTransition[] {
  switch (status) {
    case "DRAFT":
      return ["publish"];
    case "PUBLISHED":
      return ["activate"];
    case "ACTIVE":
      return ["suspend", "close"];
    case "SUSPENDED":
      return ["activate", "close"];
    case "CLOSED":
      return ["archive"];
    case "ARCHIVED":
      return [];
  }
}

export function offeringTransitions(
  status: OfferingStatus,
): readonly OfferingTransition[] {
  switch (status) {
    case "DRAFT":
      return ["publish", "archive"];
    case "PUBLISHED":
      return ["unavailable", "archive"];
    case "UNAVAILABLE":
      return ["publish", "archive"];
    case "ARCHIVED":
      return [];
  }
}
