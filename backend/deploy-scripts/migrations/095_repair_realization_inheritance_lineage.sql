UPDATE capability_realizations inherited
SET source_capability_id = source.capability_id,
	source_capability_name = COALESCE(cap.name, inherited.source_capability_name),
	updated_at = CURRENT_TIMESTAMP
FROM capability_realizations source
LEFT JOIN capabilities cap
	ON cap.tenant_id = source.tenant_id
	AND cap.id = source.capability_id
WHERE inherited.origin = 'Inherited'
	AND inherited.tenant_id = source.tenant_id
	AND inherited.source_realization_id = source.id
	AND (
		NULLIF(inherited.source_capability_id, '') IS NULL
		OR inherited.source_capability_id <> source.capability_id
		OR NULLIF(inherited.source_capability_name, '') IS NULL
	);

DELETE FROM capability_realizations inherited
WHERE inherited.origin = 'Inherited'
	AND (
		NULLIF(inherited.source_realization_id, '') IS NULL
		OR NOT EXISTS (
			SELECT 1
			FROM capability_realizations source
			WHERE source.tenant_id = inherited.tenant_id
				AND source.id = inherited.source_realization_id
		)
	);

WITH RECURSIVE capability_ancestors AS (
	SELECT tenant_id, id AS descendant_id, parent_id AS ancestor_id
	FROM capabilities
	WHERE parent_id IS NOT NULL

	UNION ALL

	SELECT ca.tenant_id, ca.descendant_id, c.parent_id
	FROM capability_ancestors ca
	INNER JOIN capabilities c
		ON c.tenant_id = ca.tenant_id
		AND c.id = ca.ancestor_id
	WHERE c.parent_id IS NOT NULL
)
DELETE FROM capability_realizations inherited
WHERE inherited.origin = 'Inherited'
	AND (
		NULLIF(inherited.source_capability_id, '') IS NULL
		OR NOT EXISTS (
			SELECT 1
			FROM capability_ancestors ca
			WHERE ca.tenant_id = inherited.tenant_id
				AND ca.descendant_id = inherited.source_capability_id
				AND ca.ancestor_id = inherited.capability_id
		)
	);
