ALTER TABLE forms
ADD tags TEXT[] DEFAULT '{}',
ADD owner_type TEXT NOT NULL;
