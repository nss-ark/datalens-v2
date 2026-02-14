// =============================================================================
// SEED DATA: Healthcare / Patient Records (MongoDB)
// Database: patient_records
// =============================================================================

db = db.getSiblingDB('patient_records');

// =========================================================================
// PATIENTS — Deeply nested documents with demographics
// =========================================================================
db.patients.drop();
db.patients.insertMany([
    {
        patient_id: "PAT-10001",
        name: { first: "Ramesh", middle: "Kumar", last: "Agarwal" },
        date_of_birth: new Date("1965-03-20"),
        gender: "Male",
        blood_group: "A+",
        aadhaar_number: "3456 7890 1234",
        contact: {
            phone: "+91-9876543210",
            alternate_phone: "0522-2345678",
            email: "ramesh.agarwal@gmail.com"
        },
        address: {
            line1: "145, Hazratganj",
            line2: "Near GPO",
            city: "Lucknow",
            state: "Uttar Pradesh",
            pincode: "226001",
            coordinates: { lat: 26.8467, lng: 80.9462 }
        },
        emergency_contact: {
            name: "Sunita Agarwal",
            relationship: "Spouse",
            phone: "9876500001"
        },
        insurance: {
            provider: "Star Health Insurance",
            policy_number: "SHI-2024-789012",
            valid_until: new Date("2026-12-31")
        },
        registration_date: new Date("2020-01-15"),
        status: "ACTIVE"
    },
    {
        patient_id: "PAT-10002",
        name: { first: "Fatima", last: "Begum" },
        date_of_birth: new Date("1990-08-14"),
        gender: "Female",
        blood_group: "O+",
        aadhaar_number: "567890123456",
        contact: {
            phone: "9345678901",
            email: "fatima.begum90@yahoo.com"
        },
        address: {
            line1: "23, Charminar Road",
            city: "Hyderabad",
            state: "Telangana",
            pincode: "500002"
        },
        emergency_contact: {
            name: "Ahmed Khan",
            relationship: "Brother",
            phone: "+91-9012345678"
        },
        insurance: null,
        registration_date: new Date("2022-06-10"),
        status: "ACTIVE"
    },
    {
        patient_id: "PAT-10003",
        name: { first: "Deepak", middle: "Singh", last: "Chauhan" },
        date_of_birth: new Date("1978-12-01"),
        gender: "Male",
        blood_group: "B-",
        aadhaar_number: "8901-2345-6789",
        pan_number: "ABCPC1234D",
        contact: {
            phone: "+91 98765 00002",
            email: "deepak.chauhan78@hotmail.com",
            whatsapp: "9876500002"
        },
        address: {
            line1: "Bungalow 5, Civil Lines",
            city: "Jaipur",
            state: "Rajasthan",
            pincode: "302001"
        },
        emergency_contact: {
            name: "Meena Chauhan",
            relationship: "Spouse",
            phone: "9876500003"
        },
        insurance: {
            provider: "ICICI Lombard",
            policy_number: "ICL-2023-456789",
            valid_until: new Date("2025-06-30"),
            nominee: { name: "Meena Chauhan", relationship: "Spouse", aadhaar_last4: "6789" }
        },
        allergies: ["Penicillin", "Sulfa drugs"],
        registration_date: new Date("2019-03-22"),
        status: "ACTIVE"
    },
    {
        patient_id: "PAT-10004",
        name: { first: "Anita", last: "Desai" },
        date_of_birth: new Date("2015-05-10"),
        gender: "Female",
        blood_group: "AB+",
        contact: {
            phone: "9876500004"
        },
        address: {
            line1: "12, Koramangala 5th Block",
            city: "Bangalore",
            state: "Karnataka",
            pincode: "560095"
        },
        guardian: {
            name: "Suresh Desai",
            relationship: "Father",
            phone: "+91-9876500005",
            email: "suresh.desai@gmail.com",
            aadhaar_number: "234567890123"
        },
        registration_date: new Date("2023-01-05"),
        status: "ACTIVE"
    },
    // Edge: deceased patient — PII still in system
    {
        patient_id: "PAT-10005",
        name: { first: "Govind", middle: "Prasad", last: "Tiwari" },
        date_of_birth: new Date("1940-01-15"),
        date_of_death: new Date("2024-11-20"),
        gender: "Male",
        blood_group: "O-",
        aadhaar_number: "1234 5678 9012",
        contact: {
            phone: "9876500006"
        },
        address: {
            line1: "Old City, Varanasi",
            state: "Uttar Pradesh",
            pincode: "221001"
        },
        next_of_kin: {
            name: "Ravi Tiwari",
            relationship: "Son",
            phone: "9876500007",
            email: "ravi.tiwari@gmail.com"
        },
        registration_date: new Date("2015-06-01"),
        status: "DECEASED"
    }
]);

// =========================================================================
// MEDICAL RECORDS — Sensitive health data
// =========================================================================
db.medical_records.drop();
db.medical_records.insertMany([
    {
        record_id: "MR-20001",
        patient_id: "PAT-10001",
        visit_date: new Date("2025-01-05"),
        visit_type: "OPD",
        department: "Cardiology",
        doctor: { name: "Dr. Sanjay Gupta", registration_no: "MCI-12345", phone: "9876500010" },
        vitals: { bp: "140/90", pulse: 82, temperature: 98.6, weight_kg: 78 },
        diagnosis: [
            { code: "I10", description: "Essential (primary) hypertension", severity: "MODERATE" },
            { code: "E11.9", description: "Type 2 diabetes mellitus without complications", severity: "MILD" }
        ],
        prescriptions: [
            { drug: "Amlodipine 5mg", dosage: "Once daily", duration: "30 days" },
            { drug: "Metformin 500mg", dosage: "Twice daily with meals", duration: "30 days" }
        ],
        lab_orders: ["HbA1c", "Lipid Panel", "Creatinine"],
        notes: "Patient Ramesh Agarwal, 59 years old, complains of intermittent headaches. BP elevated. Started on Amlodipine. Review in 2 weeks. Spouse Sunita (phone 9876500001) instructed on diet changes.",
        follow_up_date: new Date("2025-01-19"),
        created_by: "dr.sanjay@hospital.in"
    },
    {
        record_id: "MR-20002",
        patient_id: "PAT-10002",
        visit_date: new Date("2025-01-08"),
        visit_type: "Emergency",
        department: "Obstetrics",
        doctor: { name: "Dr. Priya Menon", registration_no: "KMC-67890" },
        diagnosis: [
            { code: "O80", description: "Single spontaneous delivery", severity: "NORMAL" }
        ],
        notes: "Patient Fatima Begum, 34 years old, admitted for delivery. Normal vaginal delivery. Baby boy, 3.2kg. Brother Ahmed Khan (9012345678) informed. Discharge planned in 48 hours.",
        admission_date: new Date("2025-01-08"),
        discharge_date: new Date("2025-01-10"),
        created_by: "dr.priya@hospital.in"
    },
    {
        record_id: "MR-20003",
        patient_id: "PAT-10003",
        visit_date: new Date("2025-01-10"),
        visit_type: "OPD",
        department: "Orthopedics",
        doctor: { name: "Dr. Vikram Rathore", registration_no: "RMC-11111", email: "vikram.rathore@hospital.in" },
        diagnosis: [
            { code: "M54.5", description: "Low back pain", severity: "MODERATE" }
        ],
        prescriptions: [
            { drug: "Diclofenac 50mg", dosage: "Twice daily after meals", duration: "7 days" },
            { drug: "Thiocolchicoside 4mg", dosage: "Twice daily", duration: "5 days" }
        ],
        imaging: [
            { type: "X-Ray", region: "Lumbar Spine", finding: "Mild degenerative changes at L4-L5", date: new Date("2025-01-10") }
        ],
        notes: "Mr. Deepak Chauhan (Aadhaar ending 6789), chronic back pain. Advised physiotherapy. PAN: ABCPC1234D noted for insurance claim. Contact wife Meena at 9876500003 for follow-up.",
        created_by: "dr.vikram@hospital.in"
    },
    {
        record_id: "MR-20004",
        patient_id: "PAT-10004",
        visit_date: new Date("2025-01-12"),
        visit_type: "OPD",
        department: "Pediatrics",
        doctor: { name: "Dr. Anitha Rao", registration_no: "KMC-22222" },
        diagnosis: [
            { code: "J06.9", description: "Acute upper respiratory infection, unspecified", severity: "MILD" }
        ],
        notes: "Minor patient Anita Desai, 9 years old. Father Suresh Desai (guardian, Aadhaar 234567890123) present. Mild URTI. Prescribed Paracetamol syrup. Follow up if fever persists beyond 3 days. Guardian phone: 9876500005.",
        guardian_consent: true,
        created_by: "dr.anitha@hospital.in"
    }
]);

// =========================================================================
// INSURANCE CLAIMS — Financial + health crossover
// =========================================================================
db.insurance_claims.drop();
db.insurance_claims.insertMany([
    {
        claim_id: "CLM-30001",
        patient_id: "PAT-10001",
        policy_number: "SHI-2024-789012",
        claim_date: new Date("2025-01-06"),
        claim_amount: 4500.00,
        approved_amount: 4500.00,
        status: "APPROVED",
        diagnosis_codes: ["I10", "E11.9"],
        hospital_name: "City Heart Hospital",
        hospital_gstin: "09AAACH1234F1Z5",
        patient_name: "Ramesh Kumar Agarwal",
        patient_aadhaar_last4: "1234",
        bank_details: { account: "10987654321098", ifsc: "SBIN0001234", name: "RAMESH K AGARWAL" },
        documents: ["prescription_scan.pdf", "bill_receipt.pdf"]
    },
    {
        claim_id: "CLM-30002",
        patient_id: "PAT-10003",
        policy_number: "ICL-2023-456789",
        claim_date: new Date("2025-01-11"),
        claim_amount: 8500.00,
        approved_amount: null,
        status: "PENDING",
        diagnosis_codes: ["M54.5"],
        hospital_name: "Fortis Hospital Jaipur",
        patient_name: "Deepak Singh Chauhan",
        patient_pan: "ABCPC1234D",
        bank_details: { account: "9876543210123", ifsc: "ICIC0002222", name: "DEEPAK S CHAUHAN" },
        documents: ["xray_report.pdf", "doctor_prescription.pdf", "id_proof.pdf"]
    }
]);

// =========================================================================
// LAB RESULTS — Test data near PII
// =========================================================================
db.lab_results.drop();
db.lab_results.insertMany([
    {
        lab_id: "LAB-40001",
        patient_id: "PAT-10001",
        patient_name: "Ramesh Agarwal",
        ordered_by: "Dr. Sanjay Gupta",
        sample_date: new Date("2025-01-05"),
        report_date: new Date("2025-01-06"),
        tests: [
            { name: "HbA1c", value: "7.2", unit: "%", reference: "4.0-5.6", flag: "HIGH" },
            { name: "Total Cholesterol", value: "220", unit: "mg/dL", reference: "<200", flag: "HIGH" },
            { name: "Creatinine", value: "1.1", unit: "mg/dL", reference: "0.7-1.3", flag: "NORMAL" }
        ],
        technician: "Tech. Ravi Kumar (ID: TECH-001)",
        verified_by: "Dr. Meena Pathologist"
    },
    {
        lab_id: "LAB-40002",
        patient_id: "PAT-10003",
        patient_name: "Deepak Chauhan",
        ordered_by: "Dr. Vikram Rathore",
        sample_date: new Date("2025-01-10"),
        report_date: new Date("2025-01-10"),
        tests: [
            { name: "ESR", value: "28", unit: "mm/hr", reference: "0-20", flag: "HIGH" },
            { name: "CRP", value: "12.5", unit: "mg/L", reference: "<10", flag: "HIGH" }
        ],
        technician: "Tech. Sunita Verma (ID: TECH-002)"
    }
]);

// =========================================================================
// CONSENT RECORDS — Audit trail
// =========================================================================
db.consent_records.drop();
db.consent_records.insertMany([
    {
        consent_id: "CST-50001",
        patient_id: "PAT-10001",
        patient_name: "Ramesh Kumar Agarwal",
        consent_type: "TREATMENT",
        consent_text: "I, Ramesh Kumar Agarwal (Aadhaar: 3456 7890 1234), consent to the treatment plan as explained by Dr. Sanjay Gupta. I understand the risks.",
        given_date: new Date("2025-01-05"),
        given_by: "PATIENT",
        witness: "Nurse Kavita (ID: NRS-001)",
        ip_address: "192.168.1.50"
    },
    {
        consent_id: "CST-50002",
        patient_id: "PAT-10004",
        patient_name: "Anita Desai",
        consent_type: "TREATMENT_MINOR",
        consent_text: "I, Suresh Desai (Father, Aadhaar: 234567890123), consent to the treatment of my daughter Anita Desai (DOB: 10-May-2015) as explained by Dr. Anitha Rao.",
        given_date: new Date("2025-01-12"),
        given_by: "GUARDIAN",
        guardian_details: { name: "Suresh Desai", aadhaar: "234567890123", relationship: "Father" }
    },
    {
        consent_id: "CST-50003",
        patient_id: "PAT-10001",
        consent_type: "DATA_SHARING",
        consent_text: "I consent to share my medical records with Star Health Insurance (Policy: SHI-2024-789012) for claim processing.",
        given_date: new Date("2025-01-06"),
        given_by: "PATIENT"
    }
]);

// =========================================================================
// DEVICE TELEMETRY (False positive trap — technical data)
// =========================================================================
db.device_telemetry.drop();
db.device_telemetry.insertMany([
    {
        device_id: "MED-ECG-001",
        mac_address: "AA:BB:CC:DD:EE:01",
        timestamp: new Date("2025-01-05T10:30:00Z"),
        location: { ward: "ICU-1", bed: "B3", gps: { lat: 26.8500, lng: 80.9500 } },
        readings: { heart_rate: 78, spo2: 97, ecg_lead: "normal_sinus" },
        firmware_version: "v2.3.1",
        serial_number: "SN-20230001-ECG",
        calibration_date: new Date("2024-06-15"),
        network: { ip: "10.0.1.50", subnet: "255.255.255.0", gateway: "10.0.1.1" }
    },
    {
        device_id: "MED-BP-002",
        mac_address: "AA:BB:CC:DD:EE:02",
        timestamp: new Date("2025-01-05T10:35:00Z"),
        location: { ward: "OPD-Cardiology", room: "R5" },
        readings: { systolic: 140, diastolic: 90, pulse: 82 },
        firmware_version: "v1.8.4",
        serial_number: "SN-20230002-BPM",
        battery_level: 85
    },
    {
        device_id: "MED-VENT-003",
        mac_address: "AA:BB:CC:DD:EE:03",
        timestamp: new Date("2025-01-05T11:00:00Z"),
        location: { ward: "ICU-2", bed: "B1" },
        readings: { fio2: 40, peep: 5, tidal_volume: 450, respiratory_rate: 14 },
        firmware_version: "v3.1.0",
        serial_number: "SN-20220003-VENT",
        maintenance_log: [
            { date: new Date("2024-12-01"), technician: "Rajiv Mehta", action: "Filter replaced" },
            { date: new Date("2024-09-15"), technician: "Anil Sharma", action: "Annual service" }
        ]
    }
]);

// Create indexes
db.patients.createIndex({ patient_id: 1 }, { unique: true });
db.medical_records.createIndex({ patient_id: 1 });
db.medical_records.createIndex({ record_id: 1 }, { unique: true });
db.insurance_claims.createIndex({ claim_id: 1 }, { unique: true });
db.lab_results.createIndex({ patient_id: 1 });
db.consent_records.createIndex({ patient_id: 1 });

print("MongoDB seed complete: patient_records database loaded with 6 collections");
