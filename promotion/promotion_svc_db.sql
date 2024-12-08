-- Creating Database for promotion
CREATE DATABASE promotion_svc_db;
USE promotion_svc_db;

-- Attributes of the table (promotion_id, promo_code, promotion_name, discount_percentage,valid_from, valid_to)
CREATE TABLE promotion (
    promo_code VARCHAR(20) PRIMARY KEY,              
    promotion_name VARCHAR(100) NOT NULL,                
    discount_percentage DECIMAL(5, 2) NOT NULL,          
    valid_from DATE NOT NULL,                            
    valid_to DATE NOT NULL                               
);

INSERT INTO promotion (promo_code, promotion_name, discount_percentage, valid_from, valid_to)
VALUES
('DECEMBERHOLIDAY', 'December Holiday Promotion - 20%', 20.00, '2024-12-01', '2024-12-20'),
('CHRISTMAS15', 'Christmas Sale - 15%', 15.00, '2024-12-15', '2024-12-25');

