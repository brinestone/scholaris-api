CREATE TABLE
    users (
        id BIGSERIAL PRIMARY KEY,
        first_name VARCHAR(255) NOT NULL,
        last_name VARCHAR(255) DEFAULT NULL,
        email VARCHAR(100) NOT NULL UNIQUE,
        dob date NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        phone VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        gender VARCHAR(6) NOT NULL,
        avatar VARCHAR(255)
    );