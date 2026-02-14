-- =============================================================================
-- SEED DATA: Corporate HR & Finance System (PostgreSQL)
-- Database: hr_finance_db
-- =============================================================================

-- =========================================================================
-- DEPARTMENTS (Non-PII reference table — scanner should skip)
-- =========================================================================
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    code VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    cost_center VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO departments (code, name, cost_center) VALUES
('ENG', 'Engineering', 'CC-1001'),
('HR', 'Human Resources', 'CC-1002'),
('FIN', 'Finance & Accounting', 'CC-1003'),
('MKT', 'Marketing', 'CC-1004'),
('OPS', 'Operations', 'CC-1005'),
('LEGAL', 'Legal & Compliance', 'CC-1006');

-- =========================================================================
-- EMPLOYEES (Heavy PII — Aadhaar, PAN, phone, email, DOB)
-- =========================================================================
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    employee_id VARCHAR(20) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    date_of_birth DATE NOT NULL,
    gender VARCHAR(20),
    personal_email VARCHAR(255),
    work_email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    alternate_phone VARCHAR(20),
    aadhaar_number VARCHAR(14),       -- Edge: with/without spaces
    pan_number VARCHAR(12),           -- Edge: some masked like ABCPX1234X
    passport_number VARCHAR(20),
    nationality VARCHAR(50),
    marital_status VARCHAR(20),
    blood_group VARCHAR(5),
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    pincode VARCHAR(10),
    department_id INT REFERENCES departments(id),
    designation VARCHAR(100),
    date_of_joining DATE,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO employees (employee_id, first_name, last_name, middle_name, date_of_birth, gender, personal_email, work_email, phone_number, alternate_phone, aadhaar_number, pan_number, passport_number, nationality, marital_status, blood_group, address_line1, address_line2, city, state, pincode, department_id, designation, date_of_joining, status) VALUES
('EMP001', 'Rajesh', 'Kumar', 'Shankar', '1985-03-15', 'Male', 'rajesh.kumar85@gmail.com', 'rajesh.kumar@acmecorp.in', '+91-9876543210', '011-23456789', '2345 6789 0123', 'ABCPK1234K', 'J1234567', 'Indian', 'Married', 'B+', '42, Nehru Nagar', 'Near SBI Bank', 'New Delhi', 'Delhi', '110001', 1, 'Senior Software Engineer', '2015-06-01', 'ACTIVE'),
('EMP002', 'Priya', 'Sharma', NULL, '1990-07-22', 'Female', 'priya.sharma90@yahoo.co.in', 'priya.sharma@acmecorp.in', '+919887654321', NULL, '9876 5432 1098', 'BCDPS5678L', NULL, 'Indian', 'Single', 'O+', 'Flat 12B, Sunshine Apartments', 'MG Road', 'Bangalore', 'Karnataka', '560001', 1, 'Tech Lead', '2018-01-15', 'ACTIVE'),
('EMP003', 'Amit', 'Patel', 'Ramesh', '1988-11-10', 'Male', 'amit.patel@hotmail.com', 'amit.patel@acmecorp.in', '9123456789', '+91-22-28765432', '456789012345', 'CDEPX9012M', 'K9876543', 'Indian', 'Married', 'A-', '15-A, Vasant Kunj', '', 'Mumbai', 'Maharashtra', '400001', 3, 'Finance Manager', '2014-04-01', 'ACTIVE'),
('EMP004', 'Deepa', 'Nair', NULL, '1992-02-28', 'Female', 'deepa.nair92@outlook.com', 'deepa.nair@acmecorp.in', '08041234567', NULL, '1234-5678-9012', 'DEFPN3456P', NULL, 'Indian', 'Single', 'AB+', 'House No 7, 3rd Cross', 'Jayanagar 4th Block', 'Bangalore', 'Karnataka', '560011', 2, 'HR Business Partner', '2019-08-12', 'ACTIVE'),
('EMP005', 'Mohammed', 'Ali', 'Husain', '1980-05-01', 'Male', 'mohd.ali80@gmail.com', 'mohammed.ali@acmecorp.in', '+91 98765 43210', '040-23456789', '567890123456', 'EFGPA7890Q', 'L5432109', 'Indian', 'Married', 'B-', '23, Jubilee Hills', 'Road No 5', 'Hyderabad', 'Telangana', '500033', 5, 'Operations Director', '2010-02-01', 'ACTIVE'),
('EMP006', 'Sneha', 'Reddy', 'Lakshmi', '1995-12-18', 'Female', 'sneha.reddy95@protonmail.com', 'sneha.reddy@acmecorp.in', '7890123456', NULL, '890 123 456 789', 'FGHPR2345R', NULL, 'Indian', 'Single', 'O-', 'Plot 45, Cyber City', 'Phase 2', 'Hyderabad', 'Telangana', '500081', 4, 'Marketing Analyst', '2021-07-01', 'ACTIVE'),
('EMP007', 'Vikram', 'Singh', NULL, '1987-09-05', 'Male', 'vikram.s87@gmail.com', 'vikram.singh@acmecorp.in', '+919012345678', '+91-141-2345678', '345678901234', 'GHIPS6789S', 'M8765432', 'Indian', 'Married', 'A+', '12, Civil Lines', NULL, 'Jaipur', 'Rajasthan', '302001', 6, 'Legal Counsel', '2016-11-15', 'ACTIVE'),
('EMP008', 'Ananya', 'Gupta', 'Devi', '1993-04-14', 'Female', 'ananya.g@icloud.com', 'ananya.gupta@acmecorp.in', '6789012345', NULL, '678901234567', 'HIJPG0123T', NULL, 'Indian', 'Single', NULL, '8/2, Salt Lake', 'Sector V', 'Kolkata', 'West Bengal', '700091', 1, 'Software Developer', '2020-03-01', 'ACTIVE'),
-- Edge: terminated employee — PII should still be detected
('EMP009', 'Ravi', 'Verma', NULL, '1975-01-20', 'Male', 'ravi.verma75@gmail.com', 'ravi.verma@acmecorp.in', '9345678901', NULL, '901234567890', 'IJKPV4567U', 'N6543210', 'Indian', 'Divorced', 'B+', '67, Gomti Nagar', 'Near Hazratganj', 'Lucknow', 'Uttar Pradesh', '226010', 3, 'Chief Financial Officer', '2005-01-01', 'TERMINATED'),
-- Edge: minimal data — many NULLs
('EMP010', 'Kavitha', 'Sundaram', NULL, '1998-08-30', 'Female', NULL, 'kavitha.s@acmecorp.in', NULL, NULL, NULL, NULL, NULL, 'Indian', NULL, NULL, NULL, NULL, 'Chennai', 'Tamil Nadu', NULL, 1, 'Intern', '2024-01-15', 'ACTIVE');

-- =========================================================================
-- PAYROLL (Financial PII — bank accounts, salary, tax IDs)
-- =========================================================================
CREATE TABLE payroll (
    id SERIAL PRIMARY KEY,
    employee_id INT REFERENCES employees(id),
    month VARCHAR(7) NOT NULL,          -- YYYY-MM
    basic_salary DECIMAL(12,2),
    hra DECIMAL(10,2),
    special_allowance DECIMAL(10,2),
    pf_deduction DECIMAL(10,2),
    tax_deducted DECIMAL(10,2),
    net_salary DECIMAL(12,2),
    bank_name VARCHAR(100),
    bank_account_number VARCHAR(30),    -- PII
    ifsc_code VARCHAR(15),              -- Indirect PII
    payment_mode VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO payroll (employee_id, month, basic_salary, hra, special_allowance, pf_deduction, tax_deducted, net_salary, bank_name, bank_account_number, ifsc_code, payment_mode) VALUES
(1, '2025-01', 85000.00, 34000.00, 21250.00, 10200.00, 15000.00, 115050.00, 'State Bank of India', '10987654321098', 'SBIN0001234', 'NEFT'),
(2, '2025-01', 95000.00, 38000.00, 23750.00, 11400.00, 18500.00, 126850.00, 'HDFC Bank', '50100012345678', 'HDFC0001111', 'NEFT'),
(3, '2025-01', 110000.00, 44000.00, 27500.00, 13200.00, 25000.00, 143300.00, 'ICICI Bank', '012345678901', 'ICIC0002222', 'RTGS'),
(4, '2025-01', 72000.00, 28800.00, 18000.00, 8640.00, 9500.00, 100660.00, 'Axis Bank', '917020012345678', 'UTIB0003333', 'NEFT'),
(5, '2025-01', 150000.00, 60000.00, 37500.00, 18000.00, 45000.00, 184500.00, 'Kotak Mahindra Bank', '1234567890123456', 'KKBK0004444', 'RTGS'),
(6, '2025-01', 55000.00, 22000.00, 13750.00, 6600.00, 5000.00, 79150.00, 'Bank of Baroda', '33456789012345', 'BARB0005555', 'NEFT'),
(7, '2025-01', 120000.00, 48000.00, 30000.00, 14400.00, 30000.00, 153600.00, 'Punjab National Bank', '4567890123456789', 'PUNB0006666', 'NEFT'),
(8, '2025-01', 65000.00, 26000.00, 16250.00, 7800.00, 7000.00, 92450.00, 'Yes Bank', '001234567890123', 'YESB0007777', 'NEFT');

-- =========================================================================
-- EMERGENCY CONTACTS (Indirect PII — linked persons)
-- =========================================================================
CREATE TABLE emergency_contacts (
    id SERIAL PRIMARY KEY,
    employee_id INT REFERENCES employees(id),
    contact_name VARCHAR(200) NOT NULL,
    relationship VARCHAR(50),
    phone_number VARCHAR(20),
    alternate_phone VARCHAR(20),
    email VARCHAR(255),
    address TEXT
);

INSERT INTO emergency_contacts (employee_id, contact_name, relationship, phone_number, alternate_phone, email, address) VALUES
(1, 'Sunita Kumar', 'Spouse', '+91-9876500000', NULL, 'sunita.k@gmail.com', '42, Nehru Nagar, New Delhi 110001'),
(1, 'Shankar Kumar', 'Father', '011-23456000', NULL, NULL, '42, Nehru Nagar, New Delhi 110001'),
(2, 'Ramesh Sharma', 'Father', '9887600000', '0141-2345000', 'ramesh.sharma@yahoo.com', 'Jaipur, Rajasthan'),
(3, 'Meena Patel', 'Spouse', '9123400000', NULL, 'meena.patel@gmail.com', 'Mumbai, Maharashtra'),
(5, 'Fatima Ali', 'Spouse', '9876500001', NULL, NULL, '23, Jubilee Hills, Hyderabad'),
(7, 'Manpreet Singh', 'Brother', '9012300000', NULL, 'manpreet.s@outlook.com', 'Jaipur, Rajasthan');

-- =========================================================================
-- PERFORMANCE REVIEWS (Free-text PII embedded in comments)
-- =========================================================================
CREATE TABLE performance_reviews (
    id SERIAL PRIMARY KEY,
    employee_id INT REFERENCES employees(id),
    review_period VARCHAR(20),
    reviewer_name VARCHAR(200),         -- PII (reviewer identity)
    rating INT CHECK (rating BETWEEN 1 AND 5),
    strengths TEXT,
    areas_for_improvement TEXT,         -- Edge: may contain PII about colleagues
    manager_comments TEXT,              -- Edge: free-text with embedded names
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO performance_reviews (employee_id, review_period, reviewer_name, rating, strengths, areas_for_improvement, manager_comments) VALUES
(1, '2024-H2', 'Vikram Singh', 4, 'Excellent technical skills. Rajesh consistently delivers high-quality code and mentors junior developers effectively.', 'Could improve on documentation. Sometimes Rajesh spends too much time on Amit Patel''s deliverables instead of focusing on his own.', 'Rajesh Kumar has been an exceptional performer this quarter. His collaboration with Priya Sharma on the payment gateway was outstanding. Recommend for promotion. Contact HR partner Deepa Nair for processing.'),
(2, '2024-H2', 'Mohammed Ali', 5, 'Priya has shown outstanding leadership. She successfully led the migration project and mentored Ananya Gupta.', 'N/A - Priya exceeds expectations in all areas.', 'Priya Sharma deserves immediate promotion to Principal Engineer. Her Aadhaar integration project saved us 2 months. cc: Rajesh Kumar for handover.'),
(8, '2024-H2', 'Priya Sharma', 3, 'Ananya is a quick learner.', 'Needs to improve communication with stakeholders. During the Q3 client call with Mr. Suresh Menon (client: Reliance Digital), she struggled to present technical details.', 'Ananya Gupta shows potential but needs more exposure. Assign her to Vikram Singh''s compliance project next quarter. Her phone 6789012345 is often unreachable during working hours.');

-- =========================================================================
-- ACCESS LOGS (Indirect PII — IP, user agent, device ID)
-- =========================================================================
CREATE TABLE access_logs (
    id SERIAL PRIMARY KEY,
    employee_id INT REFERENCES employees(id),
    login_timestamp TIMESTAMP NOT NULL,
    logout_timestamp TIMESTAMP,
    ip_address INET,                    -- Edge: INET type
    user_agent TEXT,
    device_id VARCHAR(100),             -- Edge: device fingerprint
    geo_location VARCHAR(100),          -- Edge: lat/long
    action VARCHAR(50),
    resource_accessed VARCHAR(255),
    session_token VARCHAR(255)          -- Edge: session data
);

INSERT INTO access_logs (employee_id, login_timestamp, logout_timestamp, ip_address, user_agent, device_id, geo_location, action, resource_accessed, session_token) VALUES
(1, '2025-01-10 09:15:00', '2025-01-10 18:30:00', '192.168.1.101', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', 'DEV-A1B2C3D4E5F6-RAJESH-LAPTOP', '28.6139,77.2090', 'LOGIN', '/dashboard', 'eyJhbGciOiJIUzI1NiJ9.abc123'),
(1, '2025-01-11 09:00:00', '2025-01-11 17:45:00', '10.0.0.55', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)', 'DEV-A1B2C3D4E5F6-RAJESH-LAPTOP', '28.6139,77.2090', 'LOGIN', '/payroll/view', 'eyJhbGciOiJIUzI1NiJ9.def456'),
(2, '2025-01-10 08:45:00', '2025-01-10 19:00:00', '172.16.0.42', 'Mozilla/5.0 (X11; Linux x86_64)', 'DEV-X7Y8Z9W0V1U2-PRIYA-DESKTOP', '12.9716,77.5946', 'LOGIN', '/code-review', 'eyJhbGciOiJIUzI1NiJ9.ghi789'),
(3, '2025-01-10 10:00:00', NULL, '203.0.113.50', 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0)', 'DEV-M3N4O5P6Q7R8-AMIT-IPHONE', '19.0760,72.8777', 'LOGIN', '/finance/reports', 'eyJhbGciOiJIUzI1NiJ9.jkl012'),
(5, '2025-01-10 07:30:00', '2025-01-10 20:00:00', '100.64.0.1', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Edge/120.0', 'DEV-S9T0U1V2W3X4-ALI-SURFACE', '17.385,78.4867', 'LOGIN', '/admin/settings', 'eyJhbGciOiJIUzI1NiJ9.mno345');

-- =========================================================================
-- PRODUCTS (False positive trap — names that look like person names)
-- =========================================================================
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(30) UNIQUE NOT NULL,
    product_name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    price DECIMAL(10,2),
    manufacturer VARCHAR(200),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO products (sku, product_name, description, category, price, manufacturer) VALUES
('SKU-001', 'Alex Pro Wireless Mouse', 'Ergonomic wireless mouse with 6 programmable buttons', 'Electronics', 2499.00, 'Max Technologies Pvt Ltd'),
('SKU-002', 'Victoria Secret Notebook', 'Premium leather-bound A5 notebook, 200 pages', 'Stationery', 899.00, 'Oliver Writing Instruments'),
('SKU-003', 'James Bond 007 Pen Set', 'Luxury pen set with gold trim', 'Stationery', 3500.00, 'Parker Pens India'),
('SKU-004', 'Mr. Clean All-Purpose Cleaner', 'Multi-surface cleaning solution, 1L', 'Household', 350.00, 'Procter & Gamble India'),
('SKU-005', 'Martin Guitar Strings Set', 'Acoustic guitar strings, phosphor bronze, medium gauge', 'Musical', 450.00, 'Martin & Co.'),
('SKU-006', 'Ruby Red Designer Chair', 'Ergonomic office chair with lumbar support', 'Furniture', 15999.00, 'Herman Miller India'),
('SKU-007', 'Grace Period Insurance Plan', 'Extended warranty plan for electronics, 2 years', 'Services', 1999.00, 'AcmeCorp Financial Services'),
('SKU-008', 'Lily White Paint 5L', 'Interior wall emulsion paint, washable', 'Paint', 2200.00, 'Asian Paints Ltd');

-- =========================================================================
-- COMPLIANCE AUDIT LOGS (Mix of PII and system data)
-- =========================================================================
CREATE TABLE compliance_audit (
    id SERIAL PRIMARY KEY,
    audit_timestamp TIMESTAMP DEFAULT NOW(),
    action_type VARCHAR(50),
    actor_email VARCHAR(255),           -- PII
    actor_ip VARCHAR(45),               -- PII
    target_entity VARCHAR(100),
    target_id VARCHAR(50),
    old_value TEXT,                      -- Edge: may contain serialized PII
    new_value TEXT,                      -- Edge: may contain serialized PII
    justification TEXT
);

INSERT INTO compliance_audit (audit_timestamp, action_type, actor_email, actor_ip, target_entity, target_id, old_value, new_value, justification) VALUES
('2025-01-05 14:30:00', 'DATA_ACCESS', 'deepa.nair@acmecorp.in', '192.168.1.104', 'employees', 'EMP001', NULL, NULL, 'Routine HR audit of employee records'),
('2025-01-06 10:15:00', 'DATA_MODIFY', 'amit.patel@acmecorp.in', '10.0.0.55', 'payroll', '3', '{"salary": 100000}', '{"salary": 110000}', 'Annual increment approved by Mohammed Ali'),
('2025-01-07 16:45:00', 'DATA_EXPORT', 'ravi.verma@acmecorp.in', '203.0.113.50', 'employees', '*', NULL, 'exported 250 records', 'Year-end tax filing with PAN details for all employees. Approved by legal: vikram.singh@acmecorp.in'),
('2025-01-08 09:00:00', 'DATA_DELETE', 'deepa.nair@acmecorp.in', '172.16.0.42', 'emergency_contacts', '12', '{"name": "Old Contact", "phone": "9999999999"}', NULL, 'DSAR request from employee EMP003 - Amit Patel requested removal'),
('2025-01-09 11:30:00', 'CONSENT_UPDATE', 'system@acmecorp.in', '127.0.0.1', 'consent_records', 'CST-445', NULL, '{"consent": true, "purpose": "payroll_processing"}', 'Auto-collected during onboarding of Kavitha Sundaram');
