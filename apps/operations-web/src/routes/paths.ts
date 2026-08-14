export const paths = {
  operatorLogin: "/operator-login",
  events: "/events",
  eventDashboard: "/event-dashboard",
  purchasing: "/purchasing",
  paymentVerification: "/payment-verification",
} as const;

export const routePatterns = {
  event: "/events/:eventId",
  offering: "/events/:eventId/offerings/:offeringId",
} as const;

export function eventPath(eventId: string): string {
  return `/events/${encodeURIComponent(eventId)}`;
}

export function offeringPath(eventId: string, offeringId: string): string {
  return `${eventPath(eventId)}/offerings/${encodeURIComponent(offeringId)}`;
}
