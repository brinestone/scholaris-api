DROP TABLE IF EXISTS enrollment_form_answers;

DROP TABLE IF EXISTS enrollment_form_question_options;

DROP TABLE IF EXISTS enrollment_form_questions;

DROP TYPE IF EXISTS question_type;

DROP TYPE IF EXISTS answer_type;

CREATE TABLE
    enrollment_steps_forms (
        id BIGSERIAL PRIMARY KEY,
        enrollment BIGINT NOT NULL,
        form BIGINT NOT NULL,
        FOREIGN KEY (enrollment) REFERENCES enrollments (id) ON DELETE CASCADE
    );