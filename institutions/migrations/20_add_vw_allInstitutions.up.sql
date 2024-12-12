CREATE VIEW
    vw_AllInstitutions AS
SELECT
    i.id,
    i.name,
    i.description,
    i.logo,
    i.visible, 
    i.slug,
    i.tenant,
    i.verified,
    i.created_at,
    i.updated_at,
    (
        SELECT
            id
        FROM
            func_get_academic_year (i.id, NULL)
    ) as current_year,
    (
        SELECT
            id
        FROM
            func_get_academic_term (i.id, NULL)
    ) as current_term
FROM 
    institutions i
ORDER BY
    i.updated_at DESC;