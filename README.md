# electric-carshare-cnad-asg1-s10259930
# Electric Car Sharing System

This project implements a microservices-based Electric Car Sharing System in Go. It features user management, vehicle reservation, and billing with a focus on scalability, security, and user experience.

---

## Table of Contents
1. [Overview](#overview)
2. [System Features](#system-features)
3. [Design Considerations and Microservices Architecture](#design-considerations-and-microservices-architecture)
4. [Database Schema](#database-schema)
5. [Installation Steps](#installation-steps)
7. [Conclusion](#conclusion)


---

## Overview

The Electric Car Sharing System aims to provide a seamless platform for car rentals, incorporating tiered memberships, real-time vehicle availability, and efficient billing. The architecture adopts microservices for scalability, maintainability, and fault tolerance.

---

## System Features

### User Management
- **Registration & Authentication**: Secure registration and login with email/phone verification.
- **Membership Tiers**: Basic, Premium, VIP levels with varied benefits such as hourly rates and increased booking limits. 
- **Profile Management**: Update personal details, view membership status, and rental history.

### Vehicle Reservation System
- **Real-Time Availability**: Book vehicles for specified time ranges on a specific date(Eg: 21/12/2024
08.00 to 20.00).
- **Modification & Cancellation**: Update or cancel bookings per policy. (Eg: Modification or Cancellation of booking is not allowed within 24 hours of rental)

### Billing and Payment Processing
- **Dynamic Pricing**: Calculate costs based on membership and promo code discounts.
- **Real-Time Updates**: Provide cost estimates and updates during rentals.
- **Invoicing**: Auto-generate and email invoices post-rental.

---

## Design Considerations

- **Separation of Concerns**: Each service handles a specific domain (User, Vehicle, Billing, Promotion).
- **Scalability**: Designed to handle increasing users and data load with minimal impact.
- **Security**: Implemented authentication, encrypted storage, and secure API communication.
- **Performance**: Optimized database queries and caching mechanisms for fast response times.

---

# Design Considerations and Microservices Architecture

## Services Overview

### 1. **User Service** 
This service is responsible for managing user registration, authentication, and profile management. It handles user data such as `user_id`, `name`, `email`, and `phone`. Additionally, it manages the user's membership, stored in the `users` table, which impacts their benefits (e.g., hourly rate discounts, booking limits) as per the `memberships` table. This service ensures secure user authentication by hashing passwords before storage, providing secure access to the application.

### 2. **Vehicle Service**
The service manages all vehicle-related information, including vehicle type, brand, model, and availability. It utilizes the `vehicles` table to store details and the `schedules` table to manage vehicle reservations. The service supports scheduling, checking availability, and ensuring that vehicles are reserved based on user demand, which is stored in the `schedules` table along with reservation times and statuses (`is_reserved`).

### 3. **Billing Service**
This service handles all aspects of pricing, payments, and invoice management. It processes bookings by interacting with the `bookings`, `invoice`, `billing`, and `receipt` tables. When a booking is made, the service generates an invoice, calculates the total amount, and processes payment through the `card` table. It ensures that payments are properly recorded and updates the invoice status to 'Paid' once the transaction is completed. The system also manages discounts (membership and promotional) to adjust the final amount.

### 4. **Promotion Service**
The service manages promotional codes and discount offers. It stores promotion details in the `promotion` table, including the promo code, discount percentage, and valid dates. This service ensures that active promotions are applied during booking and billing to calculate the final amount, reflecting the correct discount in the `bookings` and `invoice` tables.

## Separation of Concerns

Each service is designed with a clear responsibility, ensuring separation of concerns:

- The **User Service** is responsible for managing user data and authentication (e.g., `users` table).
- The **Vehicle Service** focuses on vehicle details and reservations (e.g., `vehicles`,`schedules` and       `booking` tables).
- The **Billing Service** handles financial transactions and payment-related data (e.g., `invoice`, `billing`, and `receipt` tables).
- The **Promotion Service** is solely dedicated to managing promotions (e.g., `promotion` table).

This separation allows each service to operate independently, improving maintainability, scalability, and ease of updates without affecting other services.

## Scalability

The microservices architecture enables each service to scale independently based on demand. For example, during high demand for vehicle reservations, the **Vehicle Service** can scale to handle more booking requests, while the **Billing Service** can scale during payment processing to manage increased transaction volume. This flexibility ensures the system can handle spikes in activity without overloading any single service, ensuring optimal performance across all services.

## Security

Security is implemented at multiple levels in the system:

- **Authentication**: The **User Service** hashes passwords using secure algorithms (e.g., bcrypt) before storing them in the database. This ensures that even if the database is compromised, user passwords remain secure.
- **Verification**: The **User Service** uses a **verification code** mechanism to confirm user identity. After registration or certain changes (e.g., email updates), the system sends a verification code to the user, which must be entered to confirm their identity. This ensures that only legitimate users can access their accounts and perform actions, adding an extra layer of security before granting full access.

## Performance

The system optimizes performance through the use of database triggers. For example, the `after_billing_insert` trigger automatically updates the invoice status to "Paid" and creates a corresponding receipt entry when a payment is processed. This reduces the need for additional API calls and ensures that data remains consistent and synchronized without additional application logic.

## Architecture diagram
![Architecture Diagram](C:\NP_IT\Sem_2.2\CNAD\Assg1\v1\electric-carshare-cnad-asg1-s10259930\images\Microservice.drawio.png)
---
# Database Schema

This project consists of multiple databases for managing users, vehicles, bookings, promotions, and billing.

---

## Databases and Tables

### **`user_svc_db`**
- **`memberships`**: Stores membership types (Basic, Premium, VIP) with discounts and booking limits.  
- **`users`**: Contains user details and links to membership types.

### **`vehicle_svc_db`**
- **`vehicles`**: Holds vehicle information like type, brand, and hourly rates.  
- **`schedules`**: Tracks vehicle availability and reservations.  
- **`bookings`**: Manages bookings, costs, and discounts.

### **`promotion_svc_db`**
- **`promotion`**: Stores promotional offers and discounts.

### **`billing_svc_db`**
- **`card`**: Contains payment card details linked to users.  
- **`invoice`**: Tracks booking invoices, discounts, and payments.  
- **`billing`**: Logs payment transactions for invoices.

---

## Key Relationships
- **`users` ↔ `memberships`**: Links users to their membership benefits.  
- **`vehicles` ↔ `schedules` ↔ `bookings`**: Connects vehicles, schedules, and bookings.  
- **`billing` ↔ `invoices` ↔ `cards`**: Tracks payments and invoices.

This schema is designed for modular, scalable data management.

---

# Installation Steps

## Option 1: Running with Batch File
1. Clone the repository:
   ```bash
   git clone https://github.com/Sa1ram06/electric-carshare-cnad-asg1-s10259930.git
2. Navigate to each service folder (user, vehicle, promotion, and billing) and copy the SQL files for each service to create the respective databases (user_svc_db, vehicle_svc_db, promotion_svc_db, billing_svc_db).
3. After copying the SQL files for each service, run the SQL commands in MySQL to create the databases. Ensure the MySQL username is user and the password is password when setting up the connection.
4. Modify the connection string in each service's Go code to point to your database host. For example:
    ```bash
    db, err = sql.Open("mysql", "user:password@tcp(your-database-host:3306)/promotion_svc_db")
    ```
    Replace your-database-host with the appropriate database host and port.
5. Navigate to the root folder of the cloned repository.
6. Run the servers by executing the following command: 
    ```bash
    .\run_servers.bat
7. A series of pop-up windows will appear. Click Allow on all four pop-ups to enable the services to run. This will start all four services required for the application to function.
8. Navigate to index page, and start a live server. 

## Option 2: Running with Docker (Currently Not Functional)
1. Clone the repository:
   ```bash
   git clone https://github.com/Sa1ram06/electric-carshare-cnad-asg1-s10259930.git
2. Navigate to each service folder (user, vehicle, promotion, and billing) and copy the SQL files for each service to create the respective databases (user_svc_db, vehicle_svc_db, promotion_svc_db, billing_svc_db).
3. After copying the SQL files for each service, run the SQL commands in MySQL to create the databases. Ensure the MySQL username is user and the password is password when setting up the connection.
4. In the root folder of the cloned repository, run the following command to build the Docker containers:
    ```bash
    docker compose build
5. Run the Docker containers with the following command:
    ```bash
   docker compose up -d
6. Navigate to index page, and start a live server. 
7. To stop the docker containers, run the following command:
    ```bash
    docker compose down 
---

## Conclusion

The Electric Car Sharing System offers a robust and scalable solution for car rentals, ensuring a seamless user experience with enhanced security, performance, and scalability. Through the use of microservices, the system is designed to grow with increasing user demands, while keeping each service independent for easier maintenance. The system focuses on securing user data, providing dynamic pricing models, and ensuring fast, real-time performance, all while maintaining flexibility for future enhancements and expansions.
