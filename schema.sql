CREATE TABLE IF NOT EXISTS users (
    reg_no INT PRIMARY KEY,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    dept VARCHAR(50) NOT NULL,
    phone_no VARCHAR(20),
    role VARCHAR(20) NOT NULL DEFAULT 'student'
);

CREATE TABLE IF NOT EXISTS calendar_days (
    calendar_date DATE PRIMARY KEY,
    day_type VARCHAR(20) NOT NULL CHECK (day_type IN ('regular', 'holiday', 'event')),
    day_description TEXT NOT NULL DEFAULT ''
);