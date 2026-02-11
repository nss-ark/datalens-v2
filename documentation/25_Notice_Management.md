# 25. Notice Management

## Overview

Notice Management is a **separate, modular sub-system** within the Consent Module. It handles the creation, versioning, translation, and lifecycle management of consent notices — the legal text presented to Data Principals when collecting consent.

Each notice is authored in English and automatically translated into all 22 Eighth Schedule languages via the **HuggingFace API** translation service, triggered from within the application.

---

## Notice Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│  Notice Management                                                     │
├──────────────────────────────────────────────────────────────────────┤
│                                                                        │
│  ┌─────────────┐    ┌─────────────────┐    ┌──────────────────┐      │
│  │  Author     │    │  Translate      │    │  Publish         │      │
│  │  (English)  │───►│  (HuggingFace)  │───►│  (Bind to Widget)│      │
│  └─────────────┘    └─────────────────┘    └──────────────────┘      │
│        │                    │                       │                  │
│        │                    │                       │                  │
│        ▼                    ▼                       ▼                  │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │  IMMUTABLE VERSION HISTORY                                       │  │
│  │  v1 (en, hi, bn...) → v2 (en, hi, bn...) → v3 (en, hi, bn...) │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                        │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Notice Lifecycle

```
  DRAFT ──► PUBLISHED ──► ACTIVE ──► ARCHIVED
    │           │            │           │
    ▼           ▼            ▼           ▼
  Editable   Locked,      Bound to    Historical
  by admin   translated   widget(s)   record only
```

| State | Description | Editable? | Translations? |
|-------|-------------|-----------|---------------|
| **DRAFT** | Being authored, not visible to users | ✅ Yes | Not yet |
| **PUBLISHED** | Content locked, translations triggered | ❌ No | ✅ In progress/complete |
| **ACTIVE** | Bound to one or more consent widgets | ❌ No | ✅ Complete |
| **ARCHIVED** | Replaced by a newer version, kept for audit | ❌ No | ✅ Frozen |

### State Transitions

| From | To | Trigger | Side Effects |
|------|----|---------|-------------|
| DRAFT | PUBLISHED | Admin clicks "Publish" | Content locked, version number incremented |
| PUBLISHED | ACTIVE | Admin binds notice to widget | Widget starts using this notice version |
| ACTIVE | ARCHIVED | New notice version becomes ACTIVE | Previous version archived, audit log entry |
| DRAFT | DRAFT | Admin edits content | No version change (same draft) |

---

## Translation Management

### HuggingFace API Integration

Translations are triggered from within the DataLens application. The English source text is sent to the HuggingFace translation API and translated into all 22 Eighth Schedule languages.

```
  Control Centre Admin              DataLens Backend              HuggingFace API
  ────────────────────            ────────────────              ───────────────
         │                              │                             │
         │  1. Click "Translate"        │                             │
         │─────────────────────────────►│                             │
         │                              │  2. For each language:      │
         │                              │     POST /translate         │
         │                              │────────────────────────────►│
         │                              │  3. Receive translation     │
         │                              │◄────────────────────────────│
         │                              │  4. Store in DB             │
         │                              │  (consent_notice_translations)
         │  5. Translation complete     │                             │
         │◄─────────────────────────────│                             │
```

### Translation Record

```json
{
  "id": "uuid",
  "notice_id": "uuid",
  "notice_version": 2,
  "language_code": "hi",
  "translated_text": "हम आपकी व्यक्तिगत डेटा को निम्नलिखित उद्देश्यों के लिए...",
  "translation_source": "HUGGINGFACE",
  "translated_at": "2026-02-10T10:35:00Z",
  "reviewed_by": null,
  "reviewed_at": null
}
```

### Manual Override

Admins can manually override any machine translation via:

```http
PUT /consent/notices/:id/translations/:lang
Authorization: Bearer {token}
Content-Type: application/json

{
  "translated_text": "Corrected translation text...",
  "translation_source": "MANUAL"
}
```

The `translation_source` field tracks whether a translation is machine-generated (`HUGGINGFACE`) or human-reviewed (`MANUAL`).

---

## Notice-Widget Binding

Each consent widget references a specific notice version. When a new notice version is published and activated, widgets are updated to point to the new version.

```
  Widget: "Website Banner"           Notice: "Privacy Notice"
  ─────────────────────           ─────────────────────
  widget_id: wdg_abc123            notice_id: ntc_xyz789
  notice_id: ntc_xyz789  ◄────►   version: 3
  notice_version: 3                status: ACTIVE
                                   languages: [en, hi, bn, ta...]
```

### Binding Rules

| Rule | Description |
|------|-------------|
| One notice per widget | Each widget is bound to exactly one notice at a time |
| Version-specific | Widget references a specific notice version, not "latest" |
| Re-consent on change | If notice version changes, existing consents remain valid but new interactions use the new notice |
| Multi-widget support | Same notice version can be bound to multiple widgets |

---

## Notice Data Model

### Entity Relationship

```
  consent_notices (1) ──── (N) consent_notice_translations
       │
       │ (1)
       │
       ▼ (N)
  consent_widgets.notice_id
```

### Schema Reference

See [09_Database_Schema.md](./09_Database_Schema.md) — **Consent Module** section for full table definitions:
- `consent_notices` — Versioned notice definitions with English source text
- `consent_notice_translations` — Per-language translations (HuggingFace output or manual)

---

## Notice API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/consent/notices` | List all notices for tenant |
| POST | `/consent/notices` | Create new notice (DRAFT) |
| GET | `/consent/notices/:id` | Get notice with all translations |
| PUT | `/consent/notices/:id` | Update notice (DRAFT only) |
| POST | `/consent/notices/:id/publish` | Publish notice (locks content, increments version) |
| POST | `/consent/notices/:id/archive` | Archive notice |
| POST | `/consent/notices/:id/translate` | Trigger HuggingFace translation for all 22 languages |
| GET | `/consent/notices/:id/translations` | List translations for a notice version |
| PUT | `/consent/notices/:id/translations/:lang` | Override a specific translation |

See [10_API_Reference.md](./10_API_Reference.md) — **Notice Management** section for request/response examples.

---

## Notice Audit Trail

Every notice action is logged immutably:

| Action | Logged Data |
|--------|-------------|
| Notice created | Author, timestamp, initial content |
| Notice edited | Author, timestamp, diff of changes |
| Notice published | Author, timestamp, version number |
| Translation triggered | Author, timestamp, languages requested |
| Translation completed | Language code, source (HUGGINGFACE/MANUAL), timestamp |
| Translation overridden | Author, language code, old text hash, new text hash |
| Notice bound to widget | Widget ID, notice version, timestamp |
| Notice archived | Author, timestamp, replacement notice ID |

---

## Related Documents

- [08_Consent_Management.md](./08_Consent_Management.md) — Consent lifecycle, widgets, multi-language support
- [09_Database_Schema.md](./09_Database_Schema.md) — Notice and translation table schemas
- [10_API_Reference.md](./10_API_Reference.md) — Notice management API endpoints
