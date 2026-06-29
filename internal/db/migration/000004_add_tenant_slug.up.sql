-- Add a unique public slug to tenants for the public booking flow.
-- The slug identifies a tenant in shareable public URLs (e.g. /public/barberia-juan).
ALTER TABLE tenants ADD COLUMN slug VARCHAR(63);

-- Backfill existing tenants by slugifying their name. For same-base collisions
-- we derive the suffix from the tenant's unique id (first 8 hex chars) instead
-- of a per-base counter. A counter like `base-2` could collide with a tenant
-- literally named "Base 2" (whose own base is already `base-2`); an id-derived
-- suffix cannot, so the final unique index can never fail to build.
WITH slugified AS (
    SELECT id,
           trim(both '-' from regexp_replace(lower(name), '[^a-z0-9]+', '-', 'g')) AS base
    FROM tenants
),
counts AS (
    SELECT base, count(*) AS n
    FROM slugified
    GROUP BY base
)
UPDATE tenants t
SET slug = CASE
    WHEN s.base = '' THEN 'tenant-' || left(t.id::text, 8)
    WHEN c.n = 1     THEN left(s.base, 63)
    -- left(base, 54) + '-' + 8 id chars = 63 max, within the column limit.
    ELSE left(s.base, 54) || '-' || left(t.id::text, 8)
END
FROM slugified s
JOIN counts c ON c.base = s.base
WHERE t.id = s.id;

-- Unique slug. NULL is allowed (multiple NULLs permitted), so a tenant without a
-- slug simply is not publicly bookable.
CREATE UNIQUE INDEX tenants_slug_idx ON tenants(slug);
