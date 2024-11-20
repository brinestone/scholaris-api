CREATE VIEW
    vw_AllQuestionOptions AS
SELECT
    f.id,
    f.prompt,
    f.is_required,
    f.type,
    f.layout_variant,
    COALESCE(
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'id',
                fqo.id,
                'caption',
                fqo.caption,
                'value',
                fqo.value,
                'image',
                fqo.image,
                'isDefault',
                fqo.is_default
            )
        ) FILTER (
            WHERE
                fqo.question IS NOT NULL
        ),
        '[]'
    ) AS options,
    f.form_group,
    f.form
FROM
    form_questions f
    LEFT JOIN form_question_options fqo ON fqo.question = f.id
GROUP BY
    f.id;