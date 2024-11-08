ALTER TABLE forms
ADD deadline TIMESTAMP,
ADD max_responses INT,
ADD max_submissions INT;

ALTER TABLE form_responses
ADD form BIGINT NOT NULL,
ADD CONSTRAINT fk_form_resonses_form_1 FOREIGN KEY (form) REFERENCES forms (id);

ALTER TABLE response_answers
ADD CONSTRAINT uq_response_queston_1 UNIQUE (response, question);