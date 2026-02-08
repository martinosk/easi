UPDATE view_element_positions
SET element_id = CASE
    WHEN element_id LIKE 'acq-%' THEN SUBSTRING(element_id FROM 5)
    WHEN element_id LIKE 'vendor-%' THEN SUBSTRING(element_id FROM 8)
    WHEN element_id LIKE 'team-%' THEN SUBSTRING(element_id FROM 6)
    ELSE element_id
END,
updated_at = NOW()
WHERE element_type = 'origin_entity'
AND (element_id LIKE 'acq-%' OR element_id LIKE 'vendor-%' OR element_id LIKE 'team-%');
