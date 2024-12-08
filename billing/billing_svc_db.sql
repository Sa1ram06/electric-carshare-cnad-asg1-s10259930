CREATE DATABASE billing_svc_db;
USE billing_svc_db;

-- Attributes of the table (card_id, card_number, card_expiry, cvv, card_balance, user_id)
CREATE TABLE card (
    card_id INT PRIMARY KEY AUTO_INCREMENT,
    card_number VARCHAR(16) NOT NULL,
    card_expiry VARCHAR(5) NOT NULL,
    cvv VARCHAR(3) NOT NULL,
    card_balance DECIMAL(10, 2) NOT NULL, 
    user_id INT NOT NULL
);
-- Attributes of the table (invoice_id, booking_id, user_id, issue_date, base_cost, promo_code, discount_applied, total_amount, details, status)
CREATE TABLE invoice (
    invoice_id INT AUTO_INCREMENT PRIMARY KEY,
    booking_id INT NOT NULL,
    user_id INT NOT NULL,
    issue_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    base_cost DECIMAL(5, 2) NOT NULL,
    promo_code VARCHAR(20), 
    discount_applied DECIMAL(5, 2) DEFAULT 0.00,  
    total_amount DECIMAL(5, 2) NOT NULL, 
    details TEXT,  
	status ENUM('Pending', 'Paid') DEFAULT 'Pending'  
);

-- Attributes of the table (billing_id, invoice_id, card_id, transaction_amount, transaction_date)
CREATE TABLE billing (
    billing_id INT AUTO_INCREMENT PRIMARY KEY,
    invoice_id INT NOT NULL,  -- Reference to invoice
    card_id INT NOT NULL,  -- Payment card used
    transaction_amount DECIMAL(5, 2) NOT NULL,
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- Payment timestamp
    FOREIGN KEY (invoice_id) REFERENCES invoice(invoice_id),  -- Reference to the invoice
    FOREIGN KEY (card_id) REFERENCES card(card_id)  -- Reference to payment card
);

-- Attributes of the table (receipt_id, billing_id, card_id, amount, date, description)
CREATE TABLE receipt (
    receipt_id INT AUTO_INCREMENT PRIMARY KEY,   
    billing_id INT NOT NULL,
    card_id INT NOT NULL, 
    amount DECIMAL(5, 2) NOT NULL,              
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,                     
    description TEXT,                            
    FOREIGN KEY (billing_id) REFERENCES billing(billing_id),
    FOREIGN KEY (card_id) REFERENCES card(card_id)
);

-- Insert cards for the three users
INSERT INTO card (card_number, card_expiry, cvv, card_balance, user_id)
VALUES 
('1234567812345678', '12/25', '123', 2000.00, 1), -- Card for John Doe
('2345678923456789', '11/25', '456', 3000.50, 2), -- Card for Jane Smith
('3456789034567890', '10/26', '789', 5500.75, 3); -- Card for Alice Johnson

-- Invoice for Booking 1: John Doe
INSERT INTO invoice (booking_id, user_id, base_cost, promo_code, discount_applied, total_amount, details, status)
VALUES 
(1, 1, 80.00, 'DECEMBERHOLIDAY', 16.00, 64.00, 'Reserved the Toyota Corolla on 2024-12-04 from 08:00 AM to 12:00 PM', 'Paid');

-- Invoice for Booking 2: John Doe
INSERT INTO invoice (booking_id, user_id, base_cost, total_amount, details, status)
VALUES 
(2, 1, 80.00, 80.00, 'Reserved the Toyota Corolla on 2024-12-10 from 06:00 PM to 10:00 PM', 'Paid');

-- Invoice for Booking 3: John Doe
INSERT INTO invoice (booking_id, user_id, base_cost, promo_code, discount_applied, total_amount, details, status)
VALUES 
(3, 1, 360.00, 'CHRISTMAS15', 54.00, 306.00, 'Reserved the Honda CR-V on 2024-12-18 from 08:00 AM to 08:00 PM', 'Paid');

-- Invoice for Booking 4: Jane Smith
INSERT INTO invoice (booking_id, user_id, base_cost, discount_applied, total_amount, details, status)
VALUES 
(4, 2, 300.00, 30.00, 270.00, 'Reserved the BMW 5 Series on 2024-12-20 from 04:00 PM to 10:00 PM', 'Paid');

-- Invoice for Booking 5: Jane Smith
INSERT INTO invoice (booking_id, user_id, base_cost, promo_code, discount_applied, total_amount, details, status)
VALUES 
(5, 2, 300.00, 'CHRISTMAS15', 70.50, 229.50, 'Reserved the BMW 5 Series on 2024-12-22 from 04:00 PM to 06:00 PM', 'Paid');

-- Invoice for Booking 6: Alice Johnson
INSERT INTO invoice (booking_id, user_id, base_cost, discount_applied, total_amount, details, status)
VALUES 
(6, 3, 480.00, 96.00, 384.00, 'Reserved the Volkswagen Golf on 2024-11-16 from 08:00 AM to 08:00 PM', 'Paid');

-- Invoice for Booking 7: Alice Johnson
INSERT INTO invoice (booking_id, user_id, base_cost, discount_applied, total_amount, details, status)
VALUES 
(7, 3, 240.00, 48.00, 192.00, 'Reserved the Mercedes C-Class on 2024-11-20 from 04:00 PM to 08:00 PM', 'Paid');


-- Billing for Booking 1: John Doe (card_id = 1)
INSERT INTO billing (invoice_id, card_id, transaction_amount)
VALUES 
(1, 1, 64.00);

-- Billing for Booking 2: John Doe (card_id = 1)
INSERT INTO billing (invoice_id, card_id, transaction_amount)
VALUES 
(2, 1, 80.00);

-- Billing for Booking 3: John Doe (card_id = 1)
INSERT INTO billing (invoice_id, card_id, transaction_amount)
VALUES 
(3, 1, 306.00);

-- Billing for Booking 4: Jane Smith (card_id = 2)
INSERT INTO billing (invoice_id, card_id, transaction_amount)
VALUES 
(4, 2, 270.00);

-- Billing for Booking 5: Jane Smith (card_id = 2)
INSERT INTO billing (invoice_id, card_id, transaction_amount)
VALUES 
(5, 2, 229.50);

-- Billing for Booking 6: Alice Johnson (card_id = 3)
INSERT INTO billing (invoice_id, card_id, transaction_amount)
VALUES 
(6, 3, 384.00);

-- Billing for Booking 7: Alice Johnson (card_id = 3)
INSERT INTO billing (invoice_id, card_id, transaction_amount)
VALUES 
(7, 3, 192.00);

-- Receipt for Billing 1: John Doe (card_id = 1)
INSERT INTO receipt (billing_id, card_id, amount, description)
VALUES 
(1, 1, 64.00, 'Payment for booking 1: Toyota Corolla, 2024-12-04, 08:00 AM to 12:00 PM');

-- Receipt for Billing 2: John Doe (card_id = 1)
INSERT INTO receipt (billing_id, card_id, amount, description)
VALUES 
(2, 1, 80.00, 'Payment for booking 2: Toyota Corolla, 2024-12-10, 06:00 PM to 10:00 PM');

-- Receipt for Billing 3: John Doe (card_id = 1)
INSERT INTO receipt (billing_id, card_id, amount, description)
VALUES 
(3, 1, 306.00, 'Payment for booking 3: Honda CR-V, 2024-12-18, 08:00 AM to 08:00 PM');

-- Receipt for Billing 4: Jane Smith (card_id = 2)
INSERT INTO receipt (billing_id, card_id, amount, description)
VALUES 
(4, 2, 270.00, 'Payment for booking 4: BMW 5 Series, 2024-12-20, 04:00 PM to 10:00 PM');

-- Receipt for Billing 5: Jane Smith (card_id = 2)
INSERT INTO receipt (billing_id, card_id, amount, description)
VALUES 
(5, 2, 229.50, 'Payment for booking 5: BMW 5 Series, 2024-12-22, 04:00 PM to 06:00 PM');

-- Receipt for Billing 6: Alice Johnson (card_id = 3)
INSERT INTO receipt (billing_id, card_id, amount, description)
VALUES 
(6, 3, 384.00, 'Payment for booking 6: Volkswagen Golf, 2024-11-16, 08:00 AM to 08:00 PM');

-- Receipt for Billing 7: Alice Johnson (card_id = 3)
INSERT INTO receipt (billing_id, card_id, amount, description)
VALUES 
(7, 3, 192.00, 'Payment for booking 7: Mercedes C-Class, 2024-11-20, 04:00 PM to 08:00 PM');


-- TAfter trigger for successful insertion in billing 
DELIMITER $$

CREATE TRIGGER after_billing_insert
AFTER INSERT ON billing
FOR EACH ROW
BEGIN
    -- Update the invoice status to 'Paid'
    UPDATE invoice 
    SET status = 'Paid'
    WHERE invoice_id = NEW.invoice_id;

    -- Insert corresponding details into the receipt table
    INSERT INTO receipt (billing_id, card_id, amount, description)
    VALUES (
        NEW.billing_id, 
        NEW.card_id, 
        NEW.transaction_amount, 
        CONCAT('Payment for Invoice ID ', NEW.invoice_id, ', Amount: ', NEW.transaction_amount)
    );
END$$

DELIMITER ;
