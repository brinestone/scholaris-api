CREATE TYPE form_status AS ENUM ('draft', 'published');

CREATE TYPE answer_type as ENUM (
    'text',
    'single-choice',
    'multiple-choice',
    'file',
    'date',
    'coords',
    'email',
    'multiline',
    'tel'
);

CREATE TABLE
    forms (
        id BIGSERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        description TEXT,
        meta_background VARCHAR(7),
        meta_bg_img TEXT,
        meta_img TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        owner BIGINT NOT NULL,
        multi_response BOOLEAN NOT NULL DEFAULT true,
        response_resubmission BOOLEAN DEFAULT true,
        status form_status DEFAULT 'draft'
    );

CREATE TABLE
    form_questions (
        id BIGSERIAL PRIMARY KEY,
        prompt TEXT NOT NULL,
        respose_type answer_type DEFAULT 'text',
        form BIGINT NOT NULL,
        is_required BOOLEAN DEFAULT true,
        type answer_type NOT NULL,
        layout_variant TEXT DEFAULT 'default',
        FOREIGN KEY (form) REFERENCES forms (id) ON DELETE CASCADE
    );

CREATE TABLE
    form_question_options (
        id BIGSERIAL PRIMARY KEY,
        caption TEXT NOT NULL,
        value TEXT,
        image TEXT,
        question BIGINT NOT NULL,
        FOREIGN KEY (question) REFERENCES form_questions (id) ON DELETE CASCADE
    );

CREATE TABLE
    form_responses (
        id BIGSERIAL,
        responder BIGINT NOT NULL,
        submitted_at TIMESTAMP,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (id)
    );

CREATE TABLE
    response_answers (
        id BIGSERIAL,
        question BIGINT NOT NULL,
        value TEXT,
        response BIGINT NOT NULL,
        FOREIGN KEY (question) REFERENCES form_questions (id),
        FOREIGN KEY (response) REFERENCES form_responses (id),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );