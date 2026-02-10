# 19. User Feedback & Suggestions

## Source

**Meeting**: Demo session with Kuruvila, Trilegal (Legal Expert)  
**Date**: January 2026  
**Context**: Product demo and feedback session for DPDPA compliance validation

---

## Priority Suggestions Summary

| # | Suggestion | Priority | Source Quote |
|---|------------|----------|--------------|
| 1 | Purpose Mapping Automation | **P0** | "Purpose mapping needs to be automated, prefill, prefill, prefill" |
| 2 | Local Storage Scanning | **P0** | "#2, data sources, local storage" |
| 3 | Sector-Wise Templates | **P1** | "Sector one, airline, hotel... that logic can always be coded" |
| 4 | DSR Auto-Verification | **P1** | "Deletion. We can prompt and make sure it's deleted using our read access" |
| 5 | Ongoing Compliance Messaging | **P1** | "One-time compliance at the end of the ongoing compliance" |
| 6 | Reduce IT Friction | **P2** | "IT department, the problem is they don't like lawyers" |
| 7 | Breach Checklist Module | **P2** | "Twenty-one, twenty-two different types of incidents" |

---

## Detailed Suggestions

### 1. Purpose Mapping Automation (P0) 🎯

**Current State**: Manual purpose assignment for every PII field discovered.

**Feedback**:
> "Purpose mapping needs to be automated, prefill, prefill, prefill, and then second round check."

**Implementation Requirements**:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    PURPOSE MAPPING AUTOMATION                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  STEP 1: Context Analysis                                                   │
│  ├── Table name patterns (hr_, customer_, marketing_)                       │
│  ├── Column relationships (name + email + phone = contact info)             │
│  └── Data source type (CRM, HRMS, ERP)                                      │
│                                                                              │
│  STEP 2: Sector Template Match                                              │
│  ├── Identify client's industry sector                                      │
│  ├── Apply sector-specific rules                                            │
│  └── Reference known privacy policies                                       │
│                                                                              │
│  STEP 3: Auto-Suggest Purpose                                               │
│  ├── HIGH confidence (>90%): Auto-assign                                    │
│  ├── MEDIUM confidence (70-90%): Suggest with one-click confirm             │
│  └── LOW confidence (<70%): Require manual selection                        │
│                                                                              │
│  STEP 4: Second Round Check                                                 │
│  └── Batch review of auto-assigned purposes                                 │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Effort**: 3-4 weeks

---

### 2. Local Storage Data Source Enhancement (P0) 💾

**Feedback**:
> "#2, data sources, local storage, local storage."

**Current Coverage**:
- File System Scanner: ✅ Exists
- S3 Scanner: ✅ Exists

**Gap Identified**: Need to ensure robust local storage handling including:
- Network drives (SMB/CIFS)
- NAS devices
- Windows shared folders
- Linux mount points

**Effort**: 2 weeks

---

### 3. Sector-Wise Templates (P1) 🏨

**Feedback**:
> "Airline industry, different airline... hotel, name, ID card, number, and then stay in India. Predictable, that logic can always be coded as a first base."  
> "ITC privacy policy, like Amazon for e-commerce, baseline information."

**Proposed Sector Templates**:

| Sector | Common Data Sets | Typical Purposes |
|--------|------------------|------------------|
| **Hospitality** | Name, ID, Phone, Stay dates, Preferences | Service delivery, Legal compliance |
| **Airlines** | Name, Passport, Contact, Travel history | Booking, Safety compliance |
| **E-commerce** | Name, Address, Payment, Order history | Order fulfillment, Marketing |
| **HR/Employment** | Employee data, ID, Bank details | Employment purposes (no consent needed) |
| **Healthcare** | Patient ID, Medical records | Treatment, Legal compliance |
| **BFSI** | KYC data, Transaction history | Service, Regulatory compliance |

**Implementation**:
```yaml
# templates/hospitality.yaml
sector: hospitality
common_tables:
  - pattern: "guest*|reservation*|booking*"
    purposes: ["service_delivery", "legal_compliance"]
    
common_columns:
  - pattern: "id_number|passport|aadhaar"
    category: "IDENTITY"
    purpose: "legal_compliance"
    lawful_basis: "legal_obligation"
```

**Effort**: 2-3 weeks (initial 6 sectors)

---

### 4. DSR Auto-Verification (P1) ✅

**Feedback**:
> "Deletion. We can prompt it and then make sure that it's deleted using our read access."

**Current State**: DSR execution completes, but no verification that data was actually deleted/corrected.

**Proposed Flow**:

```
DSR Executed ──► Wait Period ──► Re-Query Source ──► Verify Change
                 (1-5 mins)        (Read Access)      (Compare)
                                        │
                               ┌────────┴────────┐
                               ▼                 ▼
                           VERIFIED          FAILED
                           (Auto-close)      (Alert + Retry)
```

**Verification Logic**:
```go
func (e *DSRExecutor) VerifyExecution(task DSRTask) VerifyResult {
    // Wait for propagation
    time.Sleep(e.config.VerifyDelay)
    
    // Re-query the source
    currentData := e.scanner.QueryRecord(task.DataSubjectID, task.DataSource)
    
    switch task.Type {
    case DSR_ERASURE:
        if currentData == nil {
            return VerifyResult{Status: VERIFIED}
        }
        return VerifyResult{Status: FAILED, Reason: "Data still exists"}
        
    case DSR_CORRECTION:
        if currentData[task.Field] == task.NewValue {
            return VerifyResult{Status: VERIFIED}
        }
        return VerifyResult{Status: FAILED, Reason: "Value not updated"}
    }
}
```

**Effort**: 2 weeks

---

### 5. Ongoing Compliance Messaging (P1) 📊

**Feedback**:
> "Future looking... You have done the mapping as on today, you can't keep on this."  
> "One-time compliance at the end of the ongoing compliance."

**Action Items**:
1. **Product**: Ensure incremental sync is clearly visible in UI
2. **Sales/Demo**: Emphasize ongoing compliance, not just initial mapping
3. **Dashboard**: Show "Last Sync" timestamp prominently
4. **Alerts**: Notify when new data subjects or PII detected

**UI Enhancement**:
```
┌─────────────────────────────────────────────────────────────────┐
│  DATA FRESHNESS INDICATOR                                       │
├─────────────────────────────────────────────────────────────────┤
│  Initial Mapping: ✅ Complete (Jan 20, 2026)                    │
│  Last Sync: 2 hours ago                                         │
│  New Data Subjects (7 days): +143                               │
│  New PII Fields Detected: +8                                    │
│  Next Scheduled Sync: In 22 hours                               │
└─────────────────────────────────────────────────────────────────┘
```

**Effort**: 1 week (UI + messaging)

---

### 6. Reduce IT Department Friction (P2) 🤝

**Feedback**:
> "IT department, the problem is they don't like lawyers. Compliance is not their friend."  
> "As long as you demonstrate I know generally what you do, helps their conversation become smoother."

**Recommendations**:
1. **Minimize Questionnaires**: Pre-fill as much as possible
2. **Technical Language**: Speak in IT terms, not legal terms
3. **Self-Service Setup**: Agent installs without IT hand-holding
4. **Read-Only Emphasis**: We only READ, no write access needed
5. **Sector Knowledge**: Show industry understanding during onboarding

**Implementation**:
- Pre-onboarding sector selection
- Auto-generated technical brief for IT teams
- One-click agent installation scripts

**Effort**: 2-3 weeks (documentation + UX)

---

### 7. Breach/Incident Checklist Module (P2) 🚨

**Feedback**:
> "CERT-In has twenty-one, twenty-two different types of incidents. We'll help you prepare the response."  
> "It's like a compliance checklist. Please make sure you do these five things."

**Note**: DataLens does NOT detect breaches. It helps prepare the response.

**Proposed Module**:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    INCIDENT RESPONSE HELPER                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  CREATE INCIDENT                                                            │
│  ├── Incident Type (21 CERT-In types)                                       │
│  ├── Date/Time of Detection                                                 │
│  ├── Affected Systems                                                       │
│  └── Estimated Impact                                                       │
│                                                                              │
│  CHECKLIST (auto-generated based on type)                                   │
│  ☐ Contain the incident                                                     │
│  ☐ Identify affected data subjects                                         │
│  ☐ Assess if reportable to DPA Board (72 hrs)                              │
│  ☐ Prepare CERT-In notification form                                       │
│  ☐ Notify affected data subjects (if required)                             │
│  ☐ Document remediation steps                                              │
│                                                                              │
│  FORMS                                                                       │
│  └── Pre-filled templates for CERT-In reporting                            │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Effort**: 3-4 weeks

---

### 8. Handle Blank Purpose Fields (P2) ⚠️

**Feedback**:
> "Purpose is blank, they have to figure it out. If they don't figure it, fill it or delete it."  
> "We have to only prompt and tell them that this is incomplete."

**Implementation**:
- Alert/notification for PII without assigned purpose
- Dashboard widget: "X fields need purpose assignment"
- Options presented: Assign Purpose | Delete Data | Snooze

**Effort**: 1 week

---

### 9. White-Label Confirmation ✅

**Feedback**:
> "Third party portal... white label it... makes sense."

**Status**: Already planned. Confirmed as essential requirement.

---

### 10. Target Market Strategy 🎯

**Feedback**:
> "Smaller places, much easier to target. Big companies will have their own portals."  
> "Any questionnaire you send them has 20 departments. Each department will have their own ego."

**Strategic Recommendations**:
1. **Initial Focus**: Startups, SMBs, mid-market
2. **Ideal Sectors**: Hospitality, Restaurant Tech (e.g., PetPooja), Regional businesses
3. **Avoid Initially**: Large enterprises, MNCs
4. **Rationale**: Faster sales cycles, less approval friction, learn and iterate

---

### 11. Employee Data Nuance (Informational) 📝

**Legal Insight from Feedback**:
> "HR for employees under the new law, consent OK... as long as it's restricted to employment purpose."

**Implication**: 
- HR/Employee data may not require consent module
- Lawful basis = "Employment purposes"
- Some companies may still want it for best practice

**Action**: Add "Employment" as lawful basis option with appropriate guidance.

---

### 12. Templatize Over AI (Efficiency) 🔧

**Feedback**:
> "As long as you have fields to fetch, they don't have to use AI... you can templatize things."

**Principle**: 
- Don't over-engineer with AI where templates suffice
- Use templates for predictable flows (consent, DSR responses)
- Reserve AI for genuinely complex/ambiguous cases

---

## Updated Priority Matrix

```
                        IMPACT
                   HIGH │ ████████████████████████████████████████████████
                        │ █ Purpose Mapping █ Sector Templates █ DSR    █
                        │ █ Automation (P0) █ (P1)             █ Verify █
                        │ ████████████████████████████████████████████████
                        │
                        │ ████████████████████████████████████████████████
                   MED  │ █ Local Storage  █ Ongoing Sync  █ Breach    █
                        │ █ (P0)           █ Messaging     █ Checklist █
                        │ ████████████████████████████████████████████████
                        │
                        │ ████████████████████████████████████████████████
                   LOW  │ █ IT Friction Docs █ Blank Purpose Alerts    █
                        │ ████████████████████████████████████████████████
                        └─────────────────────────────────────────────────
                             LOW              MED             HIGH
                                          EFFORT
```

---

## Integration with Existing Roadmap

These suggestions have been integrated into:
- [16_Improvement_Recommendations.md](./16_Improvement_Recommendations.md) - Added purpose automation, sector templates
- [17_V2_Feature_Roadmap.md](./17_V2_Feature_Roadmap.md) - Added to Phase 1 priorities
