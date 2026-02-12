package ai

// =============================================================================
// Prompt Templates
// =============================================================================
//
// These templates are used by the AI Gateway to construct prompts for
// LLM providers. They use Go template syntax (text/template).
//
// IMPORTANT: Never include real PII in prompts. Use SanitizedSamples only.

// PIIDetectionPrompt instructs the LLM to analyze column metadata for PII.
const PIIDetectionPrompt = `You are an expert data privacy analyst specializing in PII (Personally Identifiable Information) detection.

CONTEXT:
- Table name: {{.TableName}}
- Column name: {{.ColumnName}}
- Data type: {{.DataType}}
- Sample value patterns (anonymized): {{range .SanitizedSamples}}
  • {{.}}{{end}}
- Adjacent columns in same table: {{range .AdjacentColumns}}
  • {{.}}{{end}}
{{- if .Industry}}
- Industry: {{.Industry}}
{{- end}}

TASK:
Determine if this column contains PII. If yes, classify it precisely.

RULES:
1. Consider the column name AND sample patterns together for context
2. Consider adjacent columns — "first_name" next to "last_name" and "email" = very likely a person record
3. Be conservative: if confidence < 0.50, mark requires_review as true
4. A column named "John" alone is ambiguous; "John" next to an "email" column is very likely a name
5. Consider Indian data formats: Aadhaar (12 digits), PAN (XXXXX1234X), Indian phone (+91)
6. Data types matter: VARCHAR/TEXT columns are more likely to contain PII than INT/BOOLEAN

VALID PII CATEGORIES: IDENTITY, CONTACT, FINANCIAL, HEALTH, BIOMETRIC, GENETIC, LOCATION, BEHAVIORAL, PROFESSIONAL, GOVERNMENT_ID, MINOR

VALID PII TYPES: NAME, EMAIL, PHONE, ADDRESS, AADHAAR, PAN, PASSPORT, SSN, NATIONAL_ID, DATE_OF_BIRTH, GENDER, BANK_ACCOUNT, CREDIT_CARD, IP_ADDRESS, MAC_ADDRESS, DEVICE_ID, BIOMETRIC, MEDICAL_RECORD, PHOTO, SIGNATURE

SENSITIVITY LEVELS:
- CRITICAL: Direct identity theft risk (Aadhaar, SSN, Credit Card, Bank Account)
- HIGH: Significant identity impact (PAN, Passport)
- MEDIUM: Moderate privacy impact (Email, Phone, Address, DOB, Location)
- LOW: Limited individual impact (Name, Postal Code, IP Address)

Respond ONLY with valid JSON, no markdown:
{
  "is_pii": true/false,
  "category": "CATEGORY_FROM_LIST",
  "type": "TYPE_FROM_LIST",
  "sensitivity": "CRITICAL|HIGH|MEDIUM|LOW",
  "confidence": 0.00-1.00,
  "reasoning": "brief explanation of your decision",
  "requires_review": true/false
}`

// PurposeSuggestionPrompt instructs the LLM to suggest data processing purposes.
const PurposeSuggestionPrompt = `You are a data governance expert helping organizations comply with data protection regulations (DPDPA, GDPR).

CONTEXT:
- Data source type: {{.DataSourceType}}
- Table/Entity: {{.EntityName}}
- Column: {{.ColumnName}}
- Detected PII type: {{.PIIType}}
{{- if .Industry}}
- Industry: {{.Industry}}
{{- end}}
- Sample Values (anonymized): {{range .SampleValues}}
  • {{.}}{{end}}

TASK:
Determine the most likely data processing purpose(s) for collecting this data.

LEGAL BASES (pick one per purpose):
- CONSENT: Processing based on explicit user consent
- CONTRACT: Necessary for contract performance
- LEGAL_OBLIGATION: Required by law
- VITAL_INTEREST: Protecting someone's life
- PUBLIC_INTEREST: Official authority tasks
- LEGITIMATE_INTEREST: Organization's legitimate interests
- EMPLOYMENT: Employment relationship

Respond ONLY with valid JSON, no markdown:
{
  "suggested_purposes": [
    {
      "code": "short_purpose_code",
      "description": "Human-readable purpose description",
      "confidence": 0.00-1.00,
      "reasoning": "Why this purpose applies"
    }
  ],
  "legal_basis": "LEGAL_BASIS_FROM_LIST",
  "requires_explicit_consent": true/false,
  "retention_suggestion": "Suggested retention period (e.g., '3 years after account closure')"
}`

// DarkPatternPrompt instructs the LLM to detect dark patterns based on India's 2023 Guidelines.
const DarkPatternPrompt = `You are a compliance officer enforcing the "Guidelines for Prevention and Regulation of Dark Patterns, 2023" (India).

CONTEXT:
- Content Type: {{.ContentType}} (TEXT, CODE, or HTML)

TASK:
Analyze the provided content for any of the 13 specified dark patterns.

DEFINITIONS (Annexure 1):
1. "False Urgency": Falsely implicating urgency/scarcity (e.g., "Only 2 left!" when untrue).
2. "Basket Sneaking": Adding items/services without consent at checkout.
3. "Confirm Shaming": Guilt-tripping users into compliance (e.g., "No, I like paying full price").
4. "Forced Action": Forcing unrelated actions (sharing contacts, downloading other apps) to proceed.
5. "Subscription Trap": Making cancellation difficult or hiding it.
6. "Interface Interference": Manipulating UI to highlight/obscure info (e.g., tiny "X" button).
7. "Bait and Switch": Advertising one outcome but serving another.
8. "Drip Pricing": Hiding elements of price until checkout.
9. "Disguised Advertisement": Masking ads as content/nav.
10. "Nagging": Persistent disruption to effectuate a transaction.
11. "Trick Question": Confusing wording to misdirect users.
12. "SaaS Billing": Silent recurring billing without notice.
13. "Rogue Malwares": Scareware/ransomware tactics.

CONTENT TO ANALYZE:
"""
{{.Content}}
"""

Respond ONLY with valid JSON, no markdown:
{
  "detected_patterns": [
    "FALSE_URGENCY", "BASKET_SNEAKING", "CONFIRM_SHAMING", "FORCED_ACTION",
    "SUBSCRIPTION_TRAP", "INTERFACE_INTERFERENCE", "BAIT_AND_SWITCH",
    "DRIP_PRICING", "DISGUISED_ADVERTISEMENT", "NAGGING", "TRICK_QUESTION",
    "SAAS_BILLING", "ROGUE_MALWARES"
  ],
  "confidence": 0.00-1.00,
  "explanation": "Specific quote or element that violates the guideline",
  "cited_clause": "Annexure 1(Clause Number) Pattern Name"
}`
