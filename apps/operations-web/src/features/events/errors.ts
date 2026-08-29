import { ApiError } from "@persona-apps/api-client";

export function safeErrorMessage(error: unknown): string {
  if (!(error instanceof ApiError))
    return "The request could not be completed.";
  const message = (() => {
    switch (error.kind) {
      case "unauthorized":
        return "Your operator session has expired. Sign in and try again.";
      case "forbidden":
        return "Your operator session is not allowed to perform this action.";
      case "validation":
        return "The server rejected one or more values. Review the form and try again.";
      case "conflict":
        return error.code === "stale_version"
          ? "This resource changed after it was loaded. Preserve your entries, reload the latest version, and reconcile manually."
          : "The requested change conflicts with the current server state.";
      case "not_found":
        return "The requested resource is no longer available.";
      case "service_unavailable":
      case "unavailable":
        return "The API is temporarily unavailable. Your entries have been preserved.";
      case "cancelled":
        return "The request was cancelled.";
      default:
        return "The request could not be completed.";
    }
  })();
  return error.requestId
    ? `${message} Request ID: ${error.requestId}.`
    : message;
}

export function isStaleVersion(error: unknown): boolean {
  return error instanceof ApiError && error.code === "stale_version";
}

export function needsSessionRefresh(error: unknown): boolean {
  return (
    error instanceof ApiError &&
    (error.kind === "unauthorized" || error.kind === "forbidden")
  );
}
