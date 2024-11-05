CREATE TYPE answer_type as ENUM (
	'text',
	'single-choice',
	'multiple-choice',
	'file'
);

CREATE TYPE question_type AS ENUM ('open-ended', 'multiple-choice');

CREATE TYPE enrollment_status AS ENUM ('draft', 'pending', 'rejected', 'approved');

CREATE TABLE
	enrollments (
		id BIGSERIAL PRIMARY KEY,
		owner BIGINT,
		approved_by BIGINT,
		approved_at TIMESTAMP,
		payment_transaction BIGINT,
		service_transaction BIGINT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		status enrollment_status DEFAULT 'draft',
		destination BIGINT NOT NULL
	);

CREATE TABLE
	enrollment_documents (
		id BIGSERIAL PRIMARY KEY,
		enrollment BIGINT NOT NULL,
		url text NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (enrollment) REFERENCES enrollments (id) ON DELETE CASCADE
	);

-- deleted in v7
CREATE TABLE
	enrollment_form_questions (
		id BIGSERIAL PRIMARY KEY,
		institution BIGINT NOT NULL,
		prompt TEXT NOT NULL,
		q_type question_type DEFAULT 'open-ended',
		a_type answer_type DEFAULT 'text',
		is_required boolean DEFAULT true,
		choice_delimiter CHAR DEFAULT ',',
		FOREIGN KEY (institution) REFERENCES institutions (id) ON DELETE CASCADE
	);

-- deleted in v7
CREATE TABLE
	enrollment_form_question_options (
		question BIGINT NOT NULL,
		label TEXT NOT NULL CHECK (TRIM(label) <> ''),
		value TEXT NOT NULL CHECK (TRIM(label) <> ''),
		is_default BOOLEAN DEFAULT false,
		FOREIGN KEY (question) REFERENCES enrollment_form_questions (id) ON DELETE CASCADE
	);

-- deleted in v7
CREATE TABLE
	enrollment_form_answers (
		question BIGINT NOT NULL,
		enrollment BIGINT NOT NULL,
		ans TEXT ARRAY,
		answered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (question) REFERENCES enrollment_form_questions (id) ON DELETE SET NULL,
		FOREIGN KEY (enrollment) REFERENCES enrollments (id) ON DELETE CASCADE
	);