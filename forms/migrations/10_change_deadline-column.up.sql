ALTER TABLE forms
ADD response_window INTERVAL,
ADD response_start DATE;

CREATE OR REPLACE VIEW
    vw_AllForms AS
SELECT
    f.id,
    f.title,
    f.description,
    f.meta_background,
    f.meta_bg_img,
    f.meta_img,
    f.created_at,
    f.updated_at,
    f.owner,
    f.owner_type,
    f.multi_response,
    f.response_resubmission,
    f.status,
    COALESCE((f.response_start + f.response_window), NULL) AS deadline,
    ARRAY_TO_JSON(
        COALESCE(
            (
                ARRAY_AGG(DISTINCT fq.id) FILTER (
                    WHERE
                        fq.id IS NOT NULL
                )
            ),
            '{}'
        )
    ) AS question_ids,
    ARRAY_TO_JSON(
        COALESCE(
            (
                ARRAY_AGG(DISTINCT fqg.id) FILTER (
                    WHERE
                        fqg.id IS NOT NULL
                )
            ),
            '{}'
        )
    ) AS question_group_ids,
    COUNT(fr.id) AS response_count,
    (
        SELECT
            COUNT(_fr.id)
        FROM
            form_responses _fr
        WHERE
            _fr.id = ANY (ARRAY_AGG(fr.id))
            AND _fr.submitted_at IS NOT NULL
    ) AS submission_count,
    ARRAY_TO_JSON(COALESCE(f.tags, '{}')) AS tags,
    f.max_responses,
    f.response_start,
    COALESCE(
        EXTRACT(
            EPOCH
            FROM
                f.response_window
        ) * 1000000,
        NULL
    ) AS response_window
FROM
    forms f
    LEFT JOIN form_questions fq ON fq.form = f.id
    LEFT JOIN form_question_groups fqg ON fqg.form = f.id
    LEFT JOIN form_responses fr ON fr.form = f.id
GROUP BY
    f.id;

ALTER TABLE forms
DROP deadline;