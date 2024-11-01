CREATE TABLE
    form_question_groups (
        id BIGSERIAL PRIMARY KEY,
        institution BIGINT,
        label TEXT,
        description TEXT,
        FOREIGN KEY (institution) REFERENCES institutions(id) ON DELETE CASCADE
    );

ALTER TABLE enrollment_form_questions ADD form_group BIGINT,
ADD CONSTRAINT fk_efq_fqg FOREIGN KEY (form_group) REFERENCES form_question_groups (id) ON DELETE CASCADE;