# Phase 4 — Batch 4F Task Specifications

**Sprint**: Phase 4 — OCR Adapters (Sarvam + Tesseract) + Portal Polish  
**Estimated Duration**: 1 day  
**Pre-requisites**: Batch 4E complete

---

## Execution Order

**Parallel Group 1**:
- Task 4F-1 (Backend) + Task 4F-2 (Frontend) — can run in PARALLEL

---

## Task 4F-1: Backend — OCR Adapter Pattern + Sarvam Vision Integration

**Agent**: Backend  
**Priority**: P0  
**Effort**: Medium (3-4h)

### Objective

Refactor the existing `parsing_service.go` to use an extensible OCR adapter pattern. Add Sarvam Vision API as a second OCR backend alongside existing Tesseract CLI. The system should try Tesseract first (local, free) and fall back to Sarvam Vision if Tesseract fails or is unavailable.

### Current State
- `internal/service/ai/parsing_service.go` (226 lines) has:
  - Native PDF, DOCX, XLSX text extraction
  - Tesseract CLI integration for images (`exec.Command("tesseract", filePath, "stdout")`)
  - `ParsingService` interface with `Parse(ctx, filePath, mimeType)` and `IsOCRAvailable()`
  - OCR availability detected by `exec.LookPath("tesseract")`

### Requirements

#### 1. OCR Adapter Interface: `internal/service/ai/ocr_adapter.go` [NEW]

```go
package ai

import "context"

// OCRAdapter defines a pluggable interface for OCR providers.
type OCRAdapter interface {
    // Name returns the adapter name (e.g., "tesseract", "sarvam").
    Name() string
    // IsAvailable returns true if the adapter is configured and ready.
    IsAvailable() bool
    // ExtractText performs OCR on the file and returns extracted text.
    ExtractText(ctx context.Context, filePath string, language string) (string, error)
    // SupportedFormats returns file extensions this adapter can process.
    SupportedFormats() []string
}
```

#### 2. Tesseract Adapter: `internal/service/ai/ocr_tesseract.go` [NEW]

Extract existing Tesseract logic from `parsing_service.go` into this adapter:
```go
type TesseractAdapter struct {
    available bool
    logger    *slog.Logger
}
```
- `IsAvailable()`: checks `exec.LookPath("tesseract")`
- `ExtractText()`: runs `tesseract <file> stdout -l <language>` (default language: "eng")
- `SupportedFormats()`: `[".png", ".jpg", ".jpeg", ".tiff", ".bmp", ".gif"]`

#### 3. Sarvam Vision Adapter: `internal/service/ai/ocr_sarvam.go` [NEW]

```go
type SarvamAdapter struct {
    apiKey    string
    baseURL   string  // default: "https://api.sarvam.ai"
    available bool
    logger    *slog.Logger
}

func NewSarvamAdapter(logger *slog.Logger) *SarvamAdapter {
    apiKey := os.Getenv("SARVAM_API_KEY")
    return &SarvamAdapter{
        apiKey:    apiKey,
        baseURL:   getEnvOrDefault("SARVAM_BASE_URL", "https://api.sarvam.ai"),
        available: apiKey != "",
        logger:    logger,
    }
}
```

`ExtractText()` implementation:
1. Read file bytes
2. POST to Sarvam Vision API endpoint with multipart/form-data:
   - Endpoint: `POST {baseURL}/api/document/ocr` (check actual docs)
   - Headers: `Authorization: Bearer {apiKey}`
   - Body: file upload
3. Parse JSON response, extract text content
4. Return extracted text

`SupportedFormats()`: `[".png", ".jpg", ".jpeg", ".pdf", ".tiff"]`

> **Note**: The exact Sarvam API endpoint and request/response format should be checked from `docs.sarvam.ai`. If the agent can't access the docs, stub the HTTP call with clear TODOs and return format expectations.

#### 4. Refactor: `internal/service/ai/parsing_service.go` [MODIFY]

- Replace `ocrAvailable bool` with `adapters []OCRAdapter`
- In `NewParsingService()`: Initialize both adapters, add available ones to the ordered list (Tesseract first, Sarvam second)
- `parseImage()`: Try adapters in order. If first fails, try second. Log which adapter was used.
- `IsOCRAvailable()`: Return true if any adapter is available
- Add `GetAvailableAdapters() []string` method to ParsingService interface

```go
func (s *parsingServiceImpl) parseImage(ctx context.Context, filePath string) (string, error) {
    for _, adapter := range s.adapters {
        if !adapter.IsAvailable() { continue }
        text, err := adapter.ExtractText(ctx, filePath, "eng")
        if err != nil {
            s.logger.Warn("OCR adapter failed, trying next", "adapter", adapter.Name(), "error", err)
            continue
        }
        if text != "" {
            s.logger.Info("OCR succeeded", "adapter", adapter.Name(), "file", filePath)
            return text, nil
        }
    }
    return "[OCR Unavailable - No adapters could extract text]", nil
}
```

#### 5. Config: Add env vars to `.env.example`

```env
# OCR — Sarvam Vision (optional, falls back to Tesseract)
SARVAM_API_KEY=
SARVAM_BASE_URL=https://api.sarvam.ai
```

### Acceptance Criteria
- [ ] `OCRAdapter` interface defined
- [ ] Tesseract adapter extracts existing logic (no behavior change)
- [ ] Sarvam adapter calls Vision API (or stubbed with TODOs)
- [ ] Adapters tried in priority order (Tesseract → Sarvam)
- [ ] Graceful fallback if adapter fails
- [ ] `go build ./...` passes

---

## Task 4F-2: Frontend — Portal Polish + Minor CC Fixes

**Agent**: Frontend  
**Priority**: P1  
**Effort**: Small (2-3h)

### Objective

Polish pass on the Data Principal Portal. The Dashboard is already well-designed (644-line component with premium inline styles). Focus on:
1. Ensuring all Portal pages are consistent in styling
2. NominationModal improvements
3. Any remaining placeholder or rough edges

### Current State
- Portal has 40 files, Dashboard is 644 lines with premium styling
- `NominationModal.tsx` exists but may need UX refinements
- `BreachNotifications.tsx` was redesigned in an earlier batch
- `Profile` components already polished

### Requirements

#### 1. Portal Page Consistency Check
Review ALL portal pages and ensure consistent styling:
- `Dashboard.tsx` — reference (already polished)
- `Login.tsx` — ensure matches design system
- `Profile.tsx` — already polished, verify
- `Requests.tsx` + `RequestNew.tsx` — ensure table/form consistency
- `History.tsx` — ensure table matches Dashboard card style
- `ConsentManage.tsx` — verify consent cards styled correctly
- `Grievance/MyGrievances.tsx` + `SubmitGrievance.tsx` — check form styling

Fix any inconsistencies found: spacing, colors, font weights, card styles, button styles.

#### 2. NominationModal Enhancement
- Ensure nomination form has clear labeling (DPDPA S14 context: "Right to Nominate")
- Add a brief explainer paragraph about what nomination means
- Ensure the modal uses consistent styling with other modals (DSRRequestModal, DPRRequestModal)

#### 3. Control Centre — Quick Polish Pass
Run through CC pages built in recent batches and verify visual consistency:
- RoPA page, Departments page, ThirdParties page — quick spot-check
- Fix any obvious spacing/alignment issues

### Acceptance Criteria
- [ ] Portal pages visually consistent with Dashboard styling
- [ ] NominationModal has clear labeling and DPDPA explainer
- [ ] No obvious styling regressions
- [ ] `npm run build -w @datalens/portal` passes
- [ ] `npm run build -w @datalens/control-centre` passes

---

## Summary

| Task | Agent | Priority | Effort |
|------|-------|----------|--------|
| 4F-1: OCR Adapter Pattern + Sarvam | Backend | P0 | Medium (3-4h) |
| 4F-2: Portal + CC Polish | Frontend | P1 | Small (2-3h) |

**Parallelism**: Both tasks run in parallel (no dependencies).
