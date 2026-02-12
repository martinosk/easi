WITH RECURSIVE capability_ancestors AS (
	SELECT tenant_id, id AS descendant_id, parent_id AS ancestor_id
	FROM capabilities
	WHERE parent_id IS NOT NULL

	UNION ALL

	SELECT ca.tenant_id, ca.descendant_id, c.parent_id
	FROM capability_ancestors ca
	INNER JOIN capabilities c
		ON c.tenant_id = ca.tenant_id AND c.id = ca.ancestor_id
	WHERE c.parent_id IS NOT NULL
)
DELETE FROM capability_realizations cr
WHERE cr.origin = 'Inherited'
	AND cr.source_capability_id IS NOT NULL
	AND NOT EXISTS (
		SELECT 1
		FROM capability_ancestors ca
		WHERE ca.tenant_id = cr.tenant_id
			AND ca.descendant_id = cr.source_capability_id
			AND ca.ancestor_id = cr.capability_id
	);
