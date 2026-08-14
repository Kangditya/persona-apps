export const paths = {
  home: "/",
  offerings: "/offerings",
  purchaseTracking: "/purchase-tracking",
} as const;

export const routePatterns = {
  offering: "/offerings/:offeringId",
} as const;

export function offeringPath(offeringId: string): string {
  return `/offerings/${encodeURIComponent(offeringId)}`;
}
