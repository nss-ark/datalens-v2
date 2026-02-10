# 05. PII Detection Engine

## Overview

The PII Detection Engine is the core scanning capability of the DataLens Agent. It uses a **multi-method approach** to identify personal information with high accuracy.

---

## Detection Methods

```
┌────────────────────────────────────────────────────────────────────────────┐
│                         PII DETECTION FLOW                                  │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Raw Data (column value or text) ─────────────────────────────►            │
│                       │                                                     │
│                       ▼                                                     │
│   ┌─────────────────────────────────────────────────┐                      │
│   │  STEP 1: Column Name Heuristics (70% confidence) │                     │
│   │  • Check if column name matches known PII fields  │                    │
│   │  • Examples: "email", "phone", "aadhaar"          │                    │
│   └─────────────────────────────────────────────────┘                      │
│                       │                                                     │
│         ┌─────────────┼─────────────┐                                      │
│         ▼ Match       │ No Match   ▼                                       │
│  ┌──────────────┐     │     ┌──────────────────────────────────────┐       │
│  │ Boost to 95% │     │     │  STEP 2: Regex Pattern Matching (90%) │      │
│  │ if regex also│     │     │  • 9 built-in patterns                │      │
│  │ matches      │     │     │  • Email, Phone, Aadhaar, PAN, etc.   │      │
│  └──────────────┘     │     └──────────────────────────────────────┘       │
│                       │                     │                               │
│                       │       ┌─────────────┼─────────────┐                │
│                       │       ▼ Match       │ No Match   ▼                 │
│                       │  [Return Result]    │                              │
│                       │                     ▼                               │
│                       │     ┌──────────────────────────────────────┐       │
│                       │     │  STEP 3: NLP Analysis (85% adjusted)  │      │
│                       │     │  • Uses spaCy NER                     │      │
│                       │     │  • Detects names, locations, orgs     │      │
│                       │     └──────────────────────────────────────┘       │
│                       │                     │                               │
│                       │                     ▼                               │
│                       │              [Return Result]                        │
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## Regex Patterns

The engine includes 9 built-in regex patterns:

| PII Category | Pattern | Examples Matched |
|--------------|---------|------------------|
| **EMAIL_ADDRESS** | `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}` | john@example.com, user.name@company.co.in |
| **PHONE_NUMBER** | `(?:\+?91[-\s]?)?[6-9]\d{9}` (India) + US pattern | +91 9876543210, 9876543210, (555) 123-4567 |
| **AADHAAR_NUMBER** | `\b[2-9]\d{3}[\s-]?\d{4}[\s-]?\d{4}\b` | 2345 6789 0123, 234567890123 |
| **PAN_NUMBER** | `[A-Z]{5}[0-9]{4}[A-Z]` | ABCDE1234F |
| **CREDIT_CARD** | Visa, MasterCard, Amex, Discover patterns | 4111111111111111 |
| **SSN** | `\b\d{3}-\d{2}-\d{4}\b` | 123-45-6789 |
| **POSTAL_CODE** | `\b[1-9][0-9]{5}\b` (India) + US ZIP | 560001, 12345-6789 |
| **IP_ADDRESS** | IPv4 pattern | 192.168.1.1 |
| **DATE_OF_BIRTH** | DD/MM/YYYY or YYYY-MM-DD | 01/15/1990, 1990-01-15 |

---

## Column Name Heuristics

Recognizes **77+ column name variations**:

### Email Patterns
| Column Name | PII Category |
|-------------|--------------|
| email, e_mail, email_address | EMAIL_ADDRESS |
| emailaddress, mail | EMAIL_ADDRESS |

### Phone Patterns
| Column Name | PII Category |
|-------------|--------------|
| phone, phonenumber, phone_number | PHONE_NUMBER |
| mobile, mobilenumber, mobile_number | PHONE_NUMBER |
| cell, cellphone, telephone, contact | PHONE_NUMBER |

### Identity Documents
| Column Name | PII Category |
|-------------|--------------|
| aadhaar, aadhar, aadhaar_number | AADHAAR_NUMBER |
| pan, pan_number, pannumber | PAN_NUMBER |
| ssn, socialsecurity, social_security | SSN |

### Name Patterns
| Column Name | PII Category |
|-------------|--------------|
| name, fullname, full_name | PERSON_NAME |
| username, user_name | PERSON_NAME |
| firstname, first_name, fname | FIRST_NAME |
| lastname, last_name, lname, surname | LAST_NAME |

### Address Patterns
| Column Name | PII Category |
|-------------|--------------|
| address, street, street_address | PHYSICAL_ADDRESS |
| postal, postal_code, zip, zipcode | POSTAL_CODE |
| pincode, pin_code | POSTAL_CODE |

### Other Patterns
| Column Name | PII Category |
|-------------|--------------|
| dob, dateofbirth, date_of_birth | DATE_OF_BIRTH |
| creditcard, credit_card, card_number | CREDIT_CARD |
| bank_account, account_number | BANK_ACCOUNT |
| ip, ip_address, ipaddress | IP_ADDRESS |
| location, latitude, longitude | LOCATION |

---

## Sensitivity Levels

Each PII category has an assigned sensitivity level:

| Level | Categories | Impact |
|-------|------------|--------|
| **CRITICAL** | SSN, Aadhaar, Credit Card, Bank Account | Direct identity theft risk |
| **HIGH** | PAN Number | Significant identity impact |
| **MEDIUM** | Email, Phone, Address, DOB, Location | Moderate privacy impact |
| **LOW** | Name, First Name, Last Name, Postal Code, IP | Limited individual impact |

---

## Confidence Scoring

### Base Confidence by Method

| Detection Method | Base Confidence |
|------------------|-----------------|
| Column Heuristic only | 0.70 (70%) |
| Regex Pattern only | 0.90 (90%) |
| Column Heuristic + Regex | 0.95 (95%) |
| NLP Detection | 0.85 × entity confidence |

### Score Boosting

When analyzing multiple samples from a column:

```
adjustedScore = avgScore × (0.7 + 0.3 × detectionRate)

Example:
- 80 out of 100 samples match email pattern
- avgScore = 0.90
- detectionRate = 0.80
- adjustedScore = 0.90 × (0.7 + 0.3 × 0.80) = 0.90 × 0.94 = 0.846
```

---

## NLP Integration

### spaCy Named Entity Recognition

The NLP service uses spaCy for:

| Entity Type | Maps to PII Category |
|-------------|---------------------|
| PERSON | PERSON_NAME |
| GPE (Geo-Political Entity) | LOCATION |
| LOC (Location) | LOCATION |
| ORG (Organization) | ORGANIZATION |

### NLP Service Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     NLP MICROSERVICE                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Agent (Go) ──► HTTP POST /analyze ──► Flask (Python)           │
│                     │                        │                   │
│                     │                        ▼                   │
│                     │              ┌────────────────┐            │
│                     │              │  spaCy NLP     │            │
│                     │              │  Model         │            │
│                     │              │  (en_core_web) │            │
│                     │              └────────────────┘            │
│                     │                        │                   │
│                     ◄─── JSON Response ──────┘                   │
│                     [{entity, label, confidence}]                │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Code Reference

### Main Detection Function

**File**: [`services/pii_detection.go`](file:///e:/Comply%20Ark/Technical/Data%20Lens%20Application/DataLensApplication/DataLensAgent/backend/services/pii_detection.go)

```go
func (s *PIIDetectionService) DetectPII(text, columnName string) []models.DetectionResult {
    // Step 1: Check column name heuristics
    if category, found := s.columnHeuristics[normalizedColumn]; found {
        result := models.DetectionResult{
            PIICategory:     category,
            ConfidenceScore: 0.70,
            DetectionMethod: "COLUMN_HEURISTIC",
        }
        // Boost if regex also matches
        if pattern.MatchString(text) {
            result.ConfidenceScore = 0.95
            result.DetectionMethod = "COLUMN_HEURISTIC+REGEX"
        }
        return results
    }
    
    // Step 2: Apply regex patterns
    for category, pattern := range s.patterns {
        if pattern.MatchString(text) {
            results = append(results, models.DetectionResult{
                PIICategory:     category,
                ConfidenceScore: 0.90,
                DetectionMethod: "REGEX",
            })
        }
    }
    
    // Step 3: Use NLP if no matches
    if len(results) == 0 && s.nlpClient.IsAvailable() {
        nlpResults, _ := s.nlpClient.AnalyzeText(text)
        // ... process NLP results
    }
    
    return s.deduplicateResults(results)
}
```

### Column Analysis Function

```go
func (s *PIIDetectionService) DetectPIIInColumn(samples []string, columnName string) *models.DetectionResult {
    // Analyze multiple samples from a column
    // Calculate category with highest average confidence
    // Boost score based on detection rate
}
```

---

## Text Column Recognition

### PostgreSQL Text Types

| Data Type | Scanned |
|-----------|---------|
| character varying, varchar | ✅ Yes |
| character, char | ✅ Yes |
| text | ✅ Yes |
| name, citext, bpchar | ✅ Yes |

### MySQL Text Types

| Data Type | Scanned |
|-----------|---------|
| varchar, char | ✅ Yes |
| text, tinytext | ✅ Yes |
| mediumtext, longtext | ✅ Yes |
| enum, set | ✅ Yes |

---

## Display Names

Human-readable names for PII categories:

| Code | Display Name |
|------|--------------|
| EMAIL_ADDRESS | Email Address |
| PHONE_NUMBER | Phone Number |
| PERSON_NAME | Person Name |
| AADHAAR_NUMBER | Aadhaar Number |
| PAN_NUMBER | PAN Number |
| CREDIT_CARD | Credit Card |
| SSN | Social Security Number |
| PHYSICAL_ADDRESS | Physical Address |
| DATE_OF_BIRTH | Date of Birth |
| IP_ADDRESS | IP Address |
| BANK_ACCOUNT | Bank Account |
| LOCATION | Location |
