-- Add a unique public slug to tenants for the public booking flow.
-- The slug identifies a tenant in shareable public URLs (e.g. /public/barberia-juan).
ALTER TABLE tenants ADD COLUMN slug VARCHAR(63);

-- Backfill existing tenants by slugifying their name, appending a numeric suffix
-- for collisions so every existing row gets a unique, usable slug.
WITH slugified AS (
    SELECT id,
           trim(both '-' from regexp_replace(lower(name), '[^a-z0-9]+', '-', 'g')) AS base
    FROM tenants
),
numbered AS (
    SELECT id, base,
           row_number() OVER (PARTITION BY base ORDER BY id) AS rn
    FROM slugified
)
UPDATE tenants t
SET slug = CASE
    WHEN n.base = '' THEN 'tenant-' || left(t.id::text, 8)
    WHEN n.rn = 1     THEN left(n.base, 63)
    ELSE left(n.base, 55) || '-' || n.rn
END
FROM numbered n
WHERE t.id = n.id;

-- Unique slug. NULL is allowed (multiple NULLs permitted), so a tenant without a
-- slug simply is not publicly bookable.
CREATE UNIQUE INDEX tenants_slug_idx ON tenants(slug);
