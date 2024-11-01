CREATE TABLE
    form_question_groups (
        id BIGSERIAL PRIMARY KEY,
        label TEXT,
        description TEXT
    );

ALTER TABLE enrollment_form_questions ADD group BIGINT;

ALTER TABLE enrollment_form_questions ADD FOREIGN KEY (group) REFERENCES form_question_groups (id) ON DELETE CASCADE;