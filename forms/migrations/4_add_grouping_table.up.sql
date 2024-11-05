CREATE TABLE
    form_question_groups (
        id BIGSERIAL PRIMARY KEY,
        form BIGINT NOT NULL,
        label TEXT DEFAULT 'default',
        description TEXT DEFAULT NULL,
        image TEXT DEFAULT NULL,
        FOREIGN KEY (form) REFERENCES forms (id) ON DELETE CASCADE
    );

ALTER TABLE form_questions ADD form_group BIGINT DEFAULT NULL,
ADD CONSTRAINT fk_fq_fqg FOREIGN KEY (form_group) REFERENCES form_question_groups (id) ON DELETE CASCADE;