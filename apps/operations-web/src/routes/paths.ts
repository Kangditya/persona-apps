export const paths = {
  operatorLogin: "/operator-login",
  events: "/events",
  eventDashboard: "/event-dashboard",
  purchasing: "/purchasing",
  paymentVerification: "/payment-verification",
} as const;

export function eventPath(eventId: string): string {
  return `/events/${encodeURIComponent(eventId)}`;
}

export function offeringPath(eventId: string, offeringId: string): string {
  return `${eventPath(eventId)}/offerings/${encodeURIComponent(offeringId)}`;
}

export function isActivePath(pathname: string, target: string): boolean {
  return pathname === target || pathname.startsWith(`${target}/`);
}
