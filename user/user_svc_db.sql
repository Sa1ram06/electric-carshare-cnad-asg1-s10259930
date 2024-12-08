-- Creating Database for users
CREATE DATABASE user_svc_db;

USE user_svc_db;

-- Attributes of the table (membership_id, hourly_rate_discount, priority_access, booking_limit)
CREATE TABLE memberships (
    membership_id VARCHAR(20) PRIMARY KEY CHECK (membership_id IN ('Basic', 'Premium', 'VIP')),
    hourly_rate_discount DECIMAL(5, 2) NOT NULL DEFAULT 0.00, 
    booking_limit INT NOT NULL DEFAULT 0 
);

-- Attributes of the table (user_id, name, email, phone, dob, hashed-password, membership_id, verification_code, verified) 
CREATE TABLE users (
	user_id INT PRIMARY KEY auto_increment,
	email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
	phone CHAR(8) NOT NULL UNIQUE,
	dob DATE NOT NULL,  
	password VARCHAR(255) NOT NULL,
	membership_id VARCHAR(20) DEFAULT 'Basic',
	license_number VARCHAR(50),
    license_expiry DATE,        
	verification_code VARCHAR(6),
    verified BOOLEAN DEFAULT FALSE,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT fk_memberships FOREIGN KEY (membership_id) REFERENCES memberships(membership_id) -- reference to membership table
);

-- Insert values into memberships table
INSERT INTO memberships (membership_id, hourly_rate_discount, booking_limit) 
VALUES 
    ('Basic', 0.00, 3),  -- No discount for Basic membership
    ('Premium', 10.00, 6),  -- 10% discount for Premium
    ('VIP', 20.00, 10);  -- 20% discount for VIP

-- Insert values into users table
INSERT INTO users (email, name, phone, dob, password, membership_id, license_number, license_expiry, verification_code, verified) 
VALUES 
('john.doe@example.com', 'John Doe', '98765432', '1990-05-12', '$2a$08$xfW2Yas5NJXl1scqBSLef.Evm8FwrXYmQlZAqqYpoZIFBfYssp5wO', 'Basic', 'SG12345678', '2025-05-12', '123456', TRUE), -- password: p@ssw0rd
('jane.smith@example.com', 'Jane Smith', '91234567', '1985-09-23', '$2a$08$ZvJIeHkCQb25vDGtgPR6deL6.L5nSOwQs8.2F0K8qd64Y32DtO5nm', 'Premium', 'SG87654321', '2026-03-15', '654321', TRUE), -- password789
('alice.johnson@example.com', 'Alice Johnson', '92345678', '2000-02-18', '$2a$08$Ak5mmhVaLwLmrGd54wCQJOFf3tMG.ViZwe2WUNiHX0Iony2ZF9KnG', 'VIP', 'SG13579246', '2024-12-31', '987654', FALSE); -- password456
