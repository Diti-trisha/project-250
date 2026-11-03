CREATE TABLE IF NOT EXISTS users (
    reg_no INT PRIMARY KEY,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    dept VARCHAR(50) NOT NULL,
    phone_no VARCHAR(20)
);