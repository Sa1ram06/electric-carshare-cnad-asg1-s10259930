-- Creating Database for vehicles
CREATE DATABASE vehicle_svc_db;
USE vehicle_svc_db;

-- attributes of the table(vehicle_id, type, brand, model, license_plate, hourly_rate)
CREATE TABLE vehicles (
    vehicle_id INT PRIMARY KEY AUTO_INCREMENT,
    type VARCHAR(50) NOT NULL,
    brand VARCHAR(50) NOT NULL,
    model VARCHAR(50) NOT NULL,
    license_plate VARCHAR(20) UNIQUE NOT NULL,
    hourly_rate DECIMAL(8, 2) NOT NULL
);

-- attributes of the table (schedule_id, vehicle_id, date, end_time, start_time, is_reserved)
CREATE TABLE schedules (
    schedule_id INT PRIMARY KEY AUTO_INCREMENT,
    vehicle_id INT NOT NULL,
    date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_reserved BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(vehicle_id)
);

-- attributes of the table (booking_id, schedule_id, user_id, status, base_cost, promotion_id, membership_discount, promotion_discount, discount_applied, total_amount, last_updated)
CREATE TABLE bookings (
    booking_id INT PRIMARY KEY AUTO_INCREMENT,
    schedule_id INT NOT NULL,
	user_id INT NOT NULL,
    status ENUM('Confirmed', 'Pending','Cancelled','Completed','SessionExpired') DEFAULT 'Pending',
    base_cost DECIMAL(5, 2) NOT NULL,
	promo_code VARCHAR(20),
    membership_discount DECIMAL(5, 2) DEFAULT 0.00,
    promotion_discount DECIMAL(5, 2) DEFAULT 0.00,
    discount_applied DECIMAL(5, 2) DEFAULT 0.00,
    total_amount DECIMAL(5, 2) NOT NULL,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (schedule_id) REFERENCES schedules(schedule_id)
);

-- insert values into vehcile table
INSERT INTO vehicles (type, brand, model, license_plate, hourly_rate) 
VALUES 
('Sedan', 'Toyota', 'Corolla', 'SG1234A', 20.00),
('SUV', 'Honda', 'CR-V', 'SG5678B', 30.00),
('Sedan', 'BMW', '5 Series', 'SG9101C', 50.00),
('Hatchback', 'Volkswagen', 'Golf', 'SG1122D', 40.00),
('Coupe', 'Mercedes', 'C-Class', 'SG3344E', 60.00);

-- insert values into schedule tab
-- Adding schedules for Vehicle 1 (Toyota Corolla)
INSERT INTO schedules (vehicle_id, date, start_time, end_time, is_reserved)
VALUES
(1, '2024-12-04', '08:00:00', '12:00:00', TRUE), -- booked 
(1, '2024-12-07', '08:00:00', '12:00:00', FALSE), 
(1, '2024-12-07', '14:00:00', '18:00:00', FALSE),
(1, '2024-12-09', '14:00:00', '18:00:00', FALSE), 
(1, '2024-12-09', '08:00:00', '12:00:00', TRUE);  -- booked 

-- Adding schedules for Vehicle 2 (Honda CR-V)
INSERT INTO schedules (vehicle_id, date, start_time, end_time, is_reserved)
VALUES
(2, '2024-12-15', '08:00:00', '20:00:00', FALSE),   
(2, '2024-12-16', '08:00:00', '20:00:00', FALSE),
(2, '2024-12-17', '08:00:00', '20:00:00', FALSE), 
(2, '2024-12-18', '08:00:00', '20:00:00', TRUE),   -- booked 
(2, '2024-12-19', '08:00:00', '20:00:00', FALSE);

-- Adding schedules for Vehicle 3 (BMW 5 Series)
INSERT INTO schedules (vehicle_id, date, start_time, end_time, is_reserved)
VALUES
(3, '2024-12-20', '08:00:00', '14:00:00', FALSE), 
(3, '2024-12-20', '16:00:00', '22:00:00', TRUE),  -- booked  
(3, '2024-12-21', '10:00:00', '14:00:00', FALSE), 
(3, '2024-12-21', '16:00:00', '20:00:00', FALSE), 
(3, '2024-12-22', '16:00:00', '18:00:00', FALSE); -- booked 

-- Adding schedules for Vehicle 4 (Volkswagen Golf)
INSERT INTO schedules (vehicle_id, date, start_time, end_time, is_reserved)
VALUES
(4, '2024-11-16', '08:00:00', '20:00:00', TRUE), -- booked 
(4, '2024-12-17', '08:00:00', '20:00:00', FALSE), 
(4, '2024-12-18', '08:00:00', '20:00:00', FALSE), 
(4, '2024-12-19', '08:00:00', '20:00:00', FALSE),  
(4, '2024-12-20', '08:00:00', '20:00:00', FALSE);  

-- Adding schedules for Vehicle 5 (Mercedes C-Class)
INSERT INTO schedules (vehicle_id, date, start_time, end_time, is_reserved)
VALUES
(5, '2024-12-20', '10:00:00', '14:00:00', FALSE), 
(5, '2024-11-20', '16:00:00', '20:00:00', TRUE),  -- booked 
(5, '2024-12-21', '10:00:00', '14:00:00', FALSE), 
(5, '2024-12-21', '16:00:00', '20:00:00', FALSE),  
(5, '2024-12-22', '08:00:00', '18:00:00', FALSE);


-- Booking 1: John Doe reserves the Toyota Corolla on 2024-12-04 from 08:00 to 12:00
INSERT INTO bookings (schedule_id, user_id, status, base_cost, promo_code, promotion_discount, discount_applied, total_amount) 
VALUES 
(1, 1, 'Completed', 80.00, 'DECEMBERHOLIDAY', 16.00, 16.00, 64.00);


-- Booking 2: John Doe reserves the Toyota Corolla on 2024-12-10 from 18:00 to 22:00
INSERT INTO bookings (schedule_id, user_id, status, base_cost, total_amount) 
VALUES 
(5, 1, 'Confirmed', 80.00, 80.00);

-- Booking 3: John Doe reserves the Honda CR-V on 2024-12-18 from 08:00 to 20:00
INSERT INTO bookings (schedule_id, user_id, status, base_cost, promo_code, promotion_discount, discount_applied, total_amount) 
VALUES 
(9, 1, 'Confirmed', 360.00, 'CHRISTMAS15', 54.00, 54.00, 306.00);


-- Booking 4: Jane smith reserves the BMW 5 Series on 2024-12-20 from 16:00 to 22:00
INSERT INTO bookings (schedule_id, user_id, status, base_cost, membership_discount, discount_applied, total_amount) 
VALUES 
(12, 2, 'Confirmed', 300.00, 30.00, 30.00, 270.00);

-- Booking 5: Jane smith reserves the BMW 5 Series on 2024-12-22 from 16:00 to 18:00
INSERT INTO bookings (schedule_id, user_id, status, base_cost, promo_code, membership_discount, promotion_discount, discount_applied, total_amount) 
VALUES 
(15, 2, 'Confirmed', 300.00, 'CHRISTMAS15', 30.00, 40.50, 70.50, 229.50);


-- Booking 6: Alice reserves the Volkswagen Golf on 2024-11-16 from 08:00 to 20:00
INSERT INTO bookings (schedule_id, user_id, status, base_cost, membership_discount, discount_applied, total_amount) 
VALUES 
(16, 3, 'Completed', 480.00, 96.00, 96.00, 384.00);

-- Booking 7: Alice Johnson reserves the Mercedes C-Class on 2024-11-20 from 16:00 to 20:00
INSERT INTO bookings (schedule_id, user_id, status, base_cost, membership_discount, discount_applied, total_amount) 
VALUES 
(22, 3, 'Completed', 240.00, 48.00, 48.00, 192.00);

