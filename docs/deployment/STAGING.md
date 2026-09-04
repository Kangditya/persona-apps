# Staging Deployment Contract

## Scope

This is a deployment contract for the Go API, its PostgreSQL database, and the
single-instance private filesystem evidence MVP. It does not provision a cloud
account, DNS, certificate, Kubernetes resource, Terraform stack, secret
manager, or storage volume.

## Required configuration

The long-running API requires:

- `APP_ENV=staging`;
- `DATABASE_URL` for a least-privilege PostgreSQL application role;
- `OIDC_ISSUER_URL`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`,
  `OIDC_REDIRECT_URL`, and `OIDC_PERMISSION_CLAIM`;
- `OPERATIONS_WEB_ORIGIN` and comma-separated
  `OPERATIONS_ALLOWED_ORIGINS` containing only exact HTTPS origins;
- optional comma-separated `STOREFRONT_ALLOWED_ORIGINS` containing only exact
  HTTPS origins when Storefront and API are deployed cross-origin; leave empty
  for the preferred same-origin deployment;
- `AUTH_COOKIE_ENCRYPTION_KEY` as base64-encoded 32 bytes;
- `IDEMPOTENCY_RESPONSE_KEYS` as an ordered
  `key-id:base64-32-byte-key` ring;
- an optional bounded `OPERATIONS_SESSION_MAX_LIFETIME`, no more than 24
  hours;
- `PUBLIC_RATE_LIMIT_PER_MINUTE` (default `60`) and
  `PUBLIC_RATE_LIMIT_BURST` (default `20`) for guest catalogue reads;
- optional comma-separated `TRUSTED_PROXY_CIDRS` containing only exact,
  non-global CIDRs for the reverse proxies that are permitted to supply
  forwarded client IPs; leave it empty to use the direct peer IP;
- `EVIDENCE_STORAGE_ROOT` for an existing absolute, non-root, non-symlink,
  owner-private directory outside the checkout and served web roots; leave it
  empty to disable only evidence upload/download;
- `EVIDENCE_STORAGE_MODE=single-instance` and `API_REPLICA_COUNT=1` whenever
  evidence storage is configured. These values are explicit outside
  development/test; invalid values or an unsafe/unavailable root fail API
  startup.

Secret values are injected by the deployment environment. They are never
committed to source control, copied into image layers, logged, or returned by
the API. Keep prior idempotency response keys available until every replay
record encrypted by them has expired.

The long-running API must have `ALLOW_DESTRUCTIVE_DB_COMMANDS=false` or leave
it unset. It refuses to start in staging or production when that value is true.

## Security boundary

- Terminate TLS at the public edge and forward only trusted request metadata.
- Apply an ingress/CDN rate limit in addition to the API's bounded per-process
  public limiter. Do not configure a broad proxy range merely to accept
  forwarded headers.
- Use Secure, HttpOnly, SameSite=Lax host-only Operations cookies.
- Allow credentialed Operations requests only from the configured exact origins.
- Use separate least-privilege database roles for the migration job and API.
- Keep payment evidence private; do not grant public object URLs or bucket
  listing permissions. Mount the evidence volume only into its one API process;
  it uses an owner-private root, `0700` directories, and `0600` files.
- Restrict database credentials and the private evidence volume to staging.
  The API never logs or returns evidence references, digests, paths, bytes, or
  credentials. Operations evidence download is API-mediated and requires an
  Operations session plus `payment.verify`; Storefront has no storage URL.
- Storefront evidence submission CORS is exact-origin and non-credentialed,
  with `Authorization` allowed only when that route is registered.

## Deployment order

1. Build and publish an immutable API artifact.
2. Run the migration job with the migration role.
3. Run reference seeds with the same controlled job.
4. Provision the persistent private evidence volume, verify its ownership and
   root safety, and deploy exactly one API replica with the application role
   and required configuration.
5. Verify `GET /health`, then PostgreSQL-backed `GET /ready`.

Migration failure stops deployment. API startup never runs migrations.

## Rollback

Rollback is an explicit operator action after a backup and compatibility
review. Do not automate a production-style database rollback in this
environment. A rollback requires the matching bounded migration command,
confirmed application compatibility, and post-action readiness verification.

## CI proof

The Validate workflow uses disposable PostgreSQL 18. It validates, applies,
versions, and reports migrations; runs reference seeds twice; executes
database-backed Go tests; rolls back the newest migration once; reapplies it;
and checks the final version. This proves the migration lifecycle, not an
external-provider deployment.

## Deferred

Backup/restore drills, monitoring, alerting, TLS provisioning, evidence
retention (including hold and deletion authority), domains, object-provider
adoption for multi-replica deployment, and production rollout are separate
work. The root lease and `API_REPLICA_COUNT` fail closed for declared or
same-volume concurrency but cannot prove a dishonest separate-volume topology;
deployment owns that enforcement and the backup/restore posture.
