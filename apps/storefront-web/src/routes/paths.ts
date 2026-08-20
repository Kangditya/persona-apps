export const paths = {
  home: "/",
  offerings: "/offerings",
  purchaseTracking: "/purchase-tracking",
} as const;

export function offeringPath(offeringId: string): string {
  return `/offerings/${encodeURIComponent(offeringId)}`;
}

export function isActivePath(pathname: string, target: string): boolean {
  return (
    pathname === target ||
    (target !== paths.home && pathname.startsWith(`${target}/`))
  );
}
