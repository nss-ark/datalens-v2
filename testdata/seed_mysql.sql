-- =============================================================================
-- SEED DATA: E-commerce Platform (MySQL) — Database: ecommerce_db
-- =============================================================================

CREATE TABLE customers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    customer_id VARCHAR(20) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    alternate_email VARCHAR(255),
    date_of_birth DATE,
    gender ENUM('Male', 'Female', 'Other', 'Prefer not to say'),
    password_hash VARCHAR(255),
    registration_ip VARCHAR(45),
    last_login_ip VARCHAR(45),
    shipping_address_line1 VARCHAR(255),
    shipping_address_line2 VARCHAR(255),
    shipping_city VARCHAR(100),
    shipping_state VARCHAR(100),
    shipping_pincode VARCHAR(10),
    billing_address_line1 VARCHAR(255),
    billing_city VARCHAR(100),
    billing_pincode VARCHAR(10),
    loyalty_points INT DEFAULT 0,
    account_status ENUM('ACTIVE', 'SUSPENDED', 'DELETED') DEFAULT 'ACTIVE',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO customers (customer_id, first_name, last_name, email, phone, alternate_email, date_of_birth, gender, password_hash, registration_ip, last_login_ip, shipping_address_line1, shipping_address_line2, shipping_city, shipping_state, shipping_pincode, billing_address_line1, billing_city, billing_pincode, loyalty_points, account_status) VALUES
('CUST-1001', 'Arjun', 'Mehta', 'arjun.mehta@gmail.com', '+91-9876501234', 'arjun.m@protonmail.com', '1990-05-12', 'Male', '$2b$12$LJ3m4ks9F7xT0ABCDEFGHIJ', '103.12.45.67', '203.0.113.42', '14, Bandra West', 'Near Linking Road', 'Mumbai', 'Maharashtra', '400050', '14, Bandra West', 'Mumbai', '400050', 2500, 'ACTIVE'),
('CUST-1002', 'Meera', 'Iyer', 'meera.iyer@yahoo.co.in', '9887612345', NULL, '1985-11-28', 'Female', '$2b$12$KLMN5678OPQRSTUVwxyz01', '49.37.128.90', '49.37.128.90', 'Flat 3A, Greenview Apt', 'Adyar', 'Chennai', 'Tamil Nadu', '600020', 'Same as shipping', 'Chennai', '600020', 800, 'ACTIVE'),
('CUST-1003', 'Karan', 'Kapoor', 'karan.kapoor@outlook.com', '+91 98112 34567', 'kk_official@hotmail.com', '1992-03-07', 'Male', '$2b$12$ABCD1234EFGHIJKLmnopqr', '157.48.200.15', '157.48.200.15', '22/B, GK-1', NULL, 'New Delhi', 'Delhi', '110048', '22/B, GK-1', 'New Delhi', '110048', 5200, 'ACTIVE'),
('CUST-1004', 'Lakshmi', 'Narayanan', 'lakshmi.n@gmail.com', '04428765432', NULL, '1978-09-15', 'Female', '$2b$12$STUV5678WXYZabcdef0123', '14.139.42.1', NULL, '67, T Nagar', '1st Street', 'Chennai', 'Tamil Nadu', '600017', '67, T Nagar', 'Chennai', '600017', 100, 'SUSPENDED'),
('CUST-1005', 'Abdullah', 'Al-Rashid', 'abdullah.r@domain.ae', '+971-50-1234567', NULL, '1988-01-30', 'Male', '$2b$12$GHIJ9012KLMNOPqrstu567', '94.200.77.103', '94.200.77.103', 'Villa 12, Palm Jumeirah', NULL, 'Dubai', 'Dubai', '00000', 'P.O. Box 54321', 'Dubai', '00000', 0, 'ACTIVE'),
('CUST-1006', 'Pooja', 'Deshmukh', 'pooja.d97@gmail.com', '7720012345', NULL, '1997-06-20', 'Female', '$2b$12$UVWXyz1234abcdefGHIJKL', '106.51.234.56', '106.51.234.89', 'Row House 5, Green Meadows', 'Kothrud', 'Pune', 'Maharashtra', '411038', 'Row House 5', 'Pune', '411038', 3100, 'ACTIVE'),
('CUST-1007', 'Suresh', 'Menon', 'suresh.menon@company.com', '+91-484-2345678', NULL, '1970-12-01', 'Male', '$2b$12$MNOPqr5678stuvWXYZ0123', '59.92.186.2', NULL, '45, MG Road', 'Marine Drive', 'Kochi', 'Kerala', '682031', '45, MG Road', 'Kochi', '682031', 15000, 'ACTIVE'),
('CUST-1008', 'Nisha', 'Aggarwal', 'nisha.agg@gmail.com', '9560012345', NULL, '1993-04-10', 'Female', '$2b$12$CDEF3456GHIJklmnOPQRst', '103.59.75.21', '103.59.75.21', '12, Rajouri Garden', NULL, 'New Delhi', 'Delhi', '110027', '12, Rajouri Garden', 'New Delhi', '110027', 0, 'DELETED');

-- =========================================================================
-- ORDERS
-- =========================================================================
CREATE TABLE orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id VARCHAR(30) UNIQUE NOT NULL,
    customer_id INT,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status ENUM('PENDING','CONFIRMED','SHIPPED','DELIVERED','CANCELLED','RETURNED') DEFAULT 'PENDING',
    total_amount DECIMAL(12,2),
    shipping_name VARCHAR(200),
    shipping_phone VARCHAR(20),
    shipping_address TEXT,
    billing_name VARCHAR(200),
    delivery_instructions TEXT,
    tracking_number VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO orders (order_id, customer_id, order_date, status, total_amount, shipping_name, shipping_phone, shipping_address, billing_name, delivery_instructions, tracking_number) VALUES
('ORD-2025-0001', 1, '2025-01-05 14:30:00', 'DELIVERED', 4598.00, 'Arjun Mehta', '+91-9876501234', '14, Bandra West, Near Linking Road, Mumbai 400050', 'Arjun Mehta', 'Ring doorbell twice. Ask for Arjun.', 'DELHUB123456789'),
('ORD-2025-0002', 1, '2025-01-10 19:15:00', 'SHIPPED', 3500.00, 'Riya Mehta', '+91-9876509999', '22, Andheri East, Mumbai 400069', 'Arjun Mehta', 'Gift for my sister Riya. Her phone: 9876509999', 'DELHUB987654321'),
('ORD-2025-0003', 3, '2025-01-08 10:00:00', 'DELIVERED', 15999.00, 'Karan Kapoor', '+91 98112 34567', '22/B, GK-1, New Delhi 110048', 'Karan Kapoor', NULL, 'BLRHUB456789012'),
('ORD-2025-0004', 6, '2025-01-12 16:45:00', 'CONFIRMED', 2200.00, 'Pooja Deshmukh', '7720012345', 'Row House 5, Kothrud, Pune 411038', 'Pooja Deshmukh', 'Leave with watchman Ramesh if not home. My number 7720012345.', 'PUNHUB111222333'),
('ORD-2025-0005', 2, '2025-01-15 08:30:00', 'PENDING', 899.00, 'Dr. Meera Iyer', '9887612345', 'Flat 3A, Greenview Apt, Adyar, Chennai 600020', 'Meera Iyer', 'Deliver 4-6 PM only. I am Dr. Meera at Apollo Hospital during day.', NULL);

-- =========================================================================
-- PAYMENT METHODS (masked card data)
-- =========================================================================
CREATE TABLE payment_methods (
    id INT AUTO_INCREMENT PRIMARY KEY,
    customer_id INT,
    card_type ENUM('VISA','MASTERCARD','RUPAY','AMEX'),
    card_holder_name VARCHAR(200),
    masked_card_number VARCHAR(20),
    expiry_month INT,
    expiry_year INT,
    billing_address VARCHAR(255),
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO payment_methods (customer_id, card_type, card_holder_name, masked_card_number, expiry_month, expiry_year, billing_address, is_default) VALUES
(1, 'VISA', 'ARJUN MEHTA', '****-****-****-4532', 12, 2027, '14 Bandra West Mumbai', TRUE),
(2, 'MASTERCARD', 'MEERA IYER', '****-****-****-1234', 3, 2028, 'Flat 3A Greenview Chennai', TRUE),
(3, 'VISA', 'KARAN KAPOOR', '****-****-****-5678', 9, 2026, '22/B GK-1 New Delhi', TRUE),
(5, 'AMEX', 'ABDULLAH AL-RASHID', '****-******-*-9012', 1, 2029, 'Villa 12 Palm Jumeirah Dubai', TRUE),
(6, 'RUPAY', 'POOJA A DESHMUKH', '****-****-****-3456', 5, 2027, 'Row House 5 Kothrud Pune', TRUE);

-- =========================================================================
-- SUPPORT TICKETS (Unstructured text heavy with PII)
-- =========================================================================
CREATE TABLE support_tickets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    ticket_id VARCHAR(20) UNIQUE NOT NULL,
    customer_id INT,
    subject VARCHAR(255),
    body TEXT,
    agent_name VARCHAR(200),
    agent_notes TEXT,
    status ENUM('OPEN','IN_PROGRESS','RESOLVED','CLOSED') DEFAULT 'OPEN',
    priority ENUM('LOW','MEDIUM','HIGH','CRITICAL') DEFAULT 'MEDIUM',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO support_tickets (ticket_id, customer_id, subject, body, agent_name, agent_notes, status, priority, resolved_at) VALUES
('TKT-001', 1, 'Wrong item delivered', 'Hi, I am Arjun Mehta (order ORD-2025-0001). I received a wrong item. My phone is 9876501234 and email arjun.mehta@gmail.com. I paid via card ending 4532.', 'Deepika Menon', 'Customer verified via Aadhaar last 4: 1234. Replacement shipped to 14 Bandra West Mumbai.', 'RESOLVED', 'HIGH', '2025-01-08 10:00:00'),
('TKT-002', 3, 'Refund not received', 'This is Karan Kapoor. Returned order ORD-2025-0003 on Jan 10 but no refund of Rs 15999 to VISA ending 5678. Email: karan.kapoor@outlook.com. UPI: karan@upi.', 'Rahul Verma', 'Refund processed to card ending 5678. Customer PAN: ABCPK9999K.', 'IN_PROGRESS', 'CRITICAL', NULL),
('TKT-003', 6, 'Account hacked', 'Someone accessed my account! Name: Pooja Deshmukh. Unauthorized orders from IP 185.220.101.42. My Aadhaar: 9876-5432-1098. Lock account! Contact: 7720012345 or pooja.d97@gmail.com.', 'Anil Kumar', 'SECURITY INCIDENT. DOB: 20-Jun-1997. Account locked. Suspicious IPs: 185.220.101.42 (TOR). Aadhaar ends 1098.', 'OPEN', 'CRITICAL', NULL);

-- =========================================================================
-- NEWSLETTER SUBSCRIBERS
-- =========================================================================
CREATE TABLE newsletter_subscribers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(200),
    subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_at_signup VARCHAR(45),
    consent_text TEXT,
    unsubscribed_at TIMESTAMP NULL,
    source VARCHAR(50)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO newsletter_subscribers (email, full_name, subscribed_at, ip_at_signup, consent_text, source) VALUES
('arjun.mehta@gmail.com', 'Arjun Mehta', '2024-06-15 10:00:00', '103.12.45.67', 'I consent to marketing emails per DPDPA 2023.', 'CHECKOUT'),
('meera.iyer@yahoo.co.in', 'Dr. Meera Iyer', '2024-08-20 14:30:00', '49.37.128.90', 'I consent to marketing emails.', 'FOOTER'),
('pooja.d97@gmail.com', 'Pooja Deshmukh', '2024-11-10 18:45:00', '106.51.234.56', 'I consent. I can unsubscribe anytime.', 'POPUP'),
('nisha.agg@gmail.com', 'Nisha Aggarwal', '2024-05-01 12:00:00', '103.59.75.21', 'I consent.', 'CHECKOUT');

-- =========================================================================
-- PRODUCT REVIEWS (UGC with PII leakage)
-- =========================================================================
CREATE TABLE product_reviews (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_sku VARCHAR(30),
    customer_id INT,
    reviewer_display_name VARCHAR(100),
    rating INT,
    review_title VARCHAR(255),
    review_body TEXT,
    helpful_votes INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO product_reviews (product_sku, customer_id, reviewer_display_name, rating, review_title, review_body, helpful_votes) VALUES
('SKU-001', 1, 'Arjun M.', 5, 'Best mouse ever!', 'Using this at my office in Bandra Mumbai. My colleague Priya also bought one!', 12),
('SKU-003', 3, 'KK_Delhi', 4, 'Great pen set', 'Gift for my father Mr. Ramesh Kapoor, 60th birthday. Delivered to GK-1 Delhi in 2 days.', 5),
('SKU-006', 6, 'Pooja D', 3, 'Good but expensive', 'Comfortable but overpriced at Rs 15999. Delivery guy called 7720012345 multiple times.', 8),
('SKU-002', 2, 'BookLover_Chennai', 5, 'Perfect notebook', 'Using for medical notes at Apollo Hospital. Recommend to all doctors. - Dr. Meera', 3),
('SKU-005', NULL, 'GuitarGuru42', 4, 'Decent strings', 'I play at open mics in Koramangala Bangalore. Contact guitarguru42@gmail.com to jam!', 1);

-- =========================================================================
-- AUDIT TRAIL (False positive trap — system UUIDs, hashes)
-- =========================================================================
CREATE TABLE audit_trail (
    id INT AUTO_INCREMENT PRIMARY KEY,
    event_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(50),
    event_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    service_name VARCHAR(100),
    request_id VARCHAR(36),
    response_code INT,
    payload_hash VARCHAR(64),
    processing_time_ms INT,
    partition_key VARCHAR(50),
    shard_id INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO audit_trail (event_id, event_type, service_name, request_id, response_code, payload_hash, processing_time_ms, partition_key, shard_id) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'ORDER_CREATED', 'order-service', 'f47ac10b-58cc-4372-a567-0e02b2c3d479', 200, 'a3f5b67890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 45, 'orders-2025-01', 3),
('550e8400-e29b-41d4-a716-446655440002', 'PAYMENT_PROCESSED', 'payment-service', 'e12ab34c-56de-7890-f123-456789abcdef', 200, 'b4e6c78901bcdef02345678901bcdef02345678901bcdef02345678901bcdef0', 120, 'payments-2025-01', 1),
('550e8400-e29b-41d4-a716-446655440003', 'INVENTORY_UPDATED', 'inventory-service', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 200, 'c5f7d89012cdef13456789012cdef13456789012cdef13456789012cdef1234', 30, 'inventory-2025-01', 2);
