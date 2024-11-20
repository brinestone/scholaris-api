CREATE VIEW
    vw_AllSettings AS
SELECT
    s.id,
    s.description,
    s.key,
    s.multi_values,
    s.created_at,
    s.updated_at,
    s.parent,
    s.owner,
    s.owner_type,
    s.created_by,
    s.overridable,
    s.system_generated,
    COALESCE(
        json_agg(
            json_build_object(
                'id',
                so.id,
                'label',
                so.label,
                'value',
                so.value,
                'setting',
                so.setting
            )
        ) FILTER (
            WHERE
                so.setting IS NOT NULL
        ),
        '[]'
    ) AS options_json,
    COALESCE(
        json_agg(
            json_build_object(
                'id',
                sv.id,
                'setting',
                sv.setting,
                'value',
                sv.value,
                'setAt',
                sv.set_at,
                'setBy',
                sv.set_by,
                'index',
                sv.value_index
            )
        ) FILTER (
            WHERE
                sv.setting IS NOT NULL
        ),
        '[]'
    ) AS values_json
FROM
    settings s
    LEFT JOIN setting_options so ON so.setting = s.id
    LEFT JOIN setting_values sv ON sv.setting = s.id
GROUP BY
    s.id;