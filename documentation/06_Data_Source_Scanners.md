# 06. Data Source Scanners

## Overview

The DataLens Agent includes specialized scanners for different data source types. Each scanner connects to the target system, retrieves metadata/samples, and uses the PII Detection Engine to identify personal data.

---

## Supported Data Sources

| Data Source | Scanner | Status |
|-------------|---------|--------|
| PostgreSQL | `database_scanner.go` | ✅ Full Support |
| MySQL | `database_scanner.go` | ✅ Full Support |
| MongoDB | `mongodb_scanner.go` | ✅ Full Support |
| SQL Server | `database_scanner.go` | 🔄 Planned |
| File System | `file_scanner.go` | ✅ Full Support |
| Amazon S3 | `s3_scanner.go` | ✅ Full Support |
| Salesforce | `salesforce_scanner.go` | ✅ Full Support |
| Email (IMAP) | `imap_scanner.go` | ✅ Full Support |

---

## Database Scanner

**File**: [`services/database_scanner.go`](file:///e:/Comply%20Ark/Technical/Data%20Lens%20Application/DataLensApplication/DataLensAgent/backend/services/database_scanner.go) (904 lines)

### Architecture

```
┌────────────────────────────────────────────────────────────────────────────┐
│                        DATABASE SCANNER SERVICE                             │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  DatabaseScannerService                                                     │
│  ├── piiDetector       *PIIDetectionService                                │
│  ├── encryptionService *EncryptionService                                  │
│  ├── dataSourceRepo    *repository.DataSourceRepository                    │
│  ├── piiCandidateRepo  *repository.PIICandidateRepository                  │
│  └── scanRunRepo       *repository.ScanRunRepository                       │
│                                                                             │
│  Methods:                                                                   │
│  ├── ScanDataSource(dataSourceID, clientID) → ScanResult                   │
│  ├── scanPostgreSQL(dataSource, connDetails, scanRunID, clientID)          │
│  ├── scanMySQL(dataSource, connDetails, scanRunID, clientID)               │
│  ├── getPostgreSQLTables(db) → []TableInfo                                 │
│  ├── getPostgreSQLColumns(db, schema, table) → []ColumnInfo                │
│  ├── samplePostgreSQLColumn(db, schema, table, column) → samples           │
│  ├── getMySQLTables(db, database) → []TableInfo                            │
│  ├── getMySQLColumns(db, database, table) → []ColumnInfo                   │
│  ├── sampleMySQLColumn(db, table, column) → samples                        │
│  ├── TestConnection(dsType, connDetailsJSON) → error                       │
│  ├── GetTables(dataSourceID) → []TableInfo                                 │
│  ├── GetTableColumns(dataSourceID, schemaName, tableName) → []ColumnInfo   │
│  └── DiscoverDatabases(dsType, serverReq) → []DatabaseInfo                 │
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
```

### Scanning Process

```
1. Load data source configuration    ─► 2. Decrypt connection details
         │                                        │
         ▼                                        ▼
3. Connect to database              ─► 4. Get list of tables
         │                                        │
         ▼                                        ▼
5. For each table:                     6. Get text columns
   │                                              │
   ▼                                              ▼
7. Sample 100 rows per column       ─► 8. Detect PII in samples
         │                                        │
         ▼                                        ▼
9. Save PII candidates              ─► 10. Update scan run status
```

### PostgreSQL Specifics

**Table Discovery Query:**
```sql
SELECT 
    schemaname AS schema, 
    tablename AS table_name
FROM pg_tables 
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY schemaname, tablename
```

**Column Discovery Query:**
```sql
SELECT 
    column_name, 
    data_type, 
    is_nullable
FROM information_schema.columns 
WHERE table_schema = $1 AND table_name = $2
```

**Sampling Query:**
```sql
SELECT DISTINCT column_name 
FROM schema.table 
WHERE column_name IS NOT NULL 
LIMIT 100
```

### MySQL Specifics

**Table Discovery Query:**
```sql
SHOW TABLES FROM database_name
```

**Column Discovery Query:**
```sql
SELECT 
    COLUMN_NAME, 
    DATA_TYPE, 
    IS_NULLABLE
FROM information_schema.COLUMNS 
WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
```

### Scan Result Structure

```go
type ScanResult struct {
    DataSourceID      int
    ScanRunID         string
    Status            string    // running, completed, failed
    TablesScanned     int
    ColumnsScanned    int
    PIICandidatesFound int
    ErrorMessage      string
    StartedAt         time.Time
    CompletedAt       time.Time
    Duration          string
}
```

---

## File Scanner

**File**: [`services/file_scanner.go`](file:///e:/Comply%20Ark/Technical/Data%20Lens%20Application/DataLensApplication/DataLensAgent/backend/services/file_scanner.go) (536 lines)

### Supported File Types

| Category | Extensions | Method |
|----------|------------|--------|
| Text Files | .txt, .log, .csv, .json, .xml | Direct text reading |
| Documents | .pdf, .docx, .doc | Text extraction |
| Spreadsheets | .xlsx, .xls | Cell-by-cell reading |
| Images | .jpg, .png, .tiff, .bmp | OCR text extraction |
| Configuration | .yaml, .yml, .ini, .cfg | Direct text reading |

### Architecture

```go
type FileScanner struct {
    piiDetector       *PIIDetectionService
    encryptionService *EncryptionService
    dataSourceRepo    *repository.DataSourceRepository
    piiCandidateRepo  *repository.PIICandidateRepository
    scanRunRepo       *repository.ScanRunRepository
    ocrService        *OCRService
    supportedExtensions map[string]bool
}
```

### Scanning Process

```
1. Get file system configuration   ─► 2. Walk directory tree
         │                                        │
         ▼                                        ▼
3. For each file:                     4. Check if extension supported
   │                                              │
   ▼                                              ▼
5. Process by file type:              6. Send text to PII detector
   - Text: Read directly
   - Image: OCR extract
   - PDF: Text extract
         │                                        │
         ▼                                        ▼
7. Save PII with file path/line    ─► 8. Update scan status
```

### Detection Result Structure

```go
type PIIFileDetection struct {
    FilePath        string   // Full path to file
    FileName        string   // Base name
    FileType        string   // Extension
    PIICategory     string   // EMAIL_ADDRESS, etc.
    ConfidenceScore float64
    SampleData      string   // Masked sample
    LineNumber      int      // For text files
    PageNumber      int      // For PDFs
    SheetName       string   // For Excel
}
```

### File Type Detection

```go
func (s *FileScanner) isLikelyTextFile(path string) bool {
    textExtensions := map[string]bool{
        ".txt": true, ".log": true, ".csv": true,
        ".json": true, ".xml": true, ".yaml": true,
        ".yml": true, ".ini": true, ".cfg": true,
        ".md": true, ".html": true, ".htm": true,
    }
    ext := strings.ToLower(filepath.Ext(path))
    return textExtensions[ext]
}
```

---

## S3 Scanner

**File**: `services/s3_scanner.go`

### Features

- Scans S3 buckets for files containing PII
- Supports same file types as File Scanner
- Uses AWS SDK for Go
- Handles pagination for large buckets

### Connection Details

```json
{
  "bucket": "my-data-bucket",
  "region": "ap-south-1",
  "access_key_id": "AKIA...",
  "secret_access_key": "***",
  "prefix": "data/uploads/",
  "max_objects": 1000
}
```

---

## Salesforce Scanner

**File**: `services/salesforce_scanner.go`

### Features

- Connects via Salesforce REST API
- Scans standard and custom objects
- Respects field-level security

### Scanned Objects

| Object | Fields Scanned |
|--------|----------------|
| Contact | Email, Phone, Name, Address |
| Lead | Email, Phone, Name, Company |
| Account | Phone, Website |
| Custom Objects | All text fields |

### Connection Details

```json
{
  "instance_url": "https://na123.salesforce.com",
  "client_id": "3MVG9...",
  "client_secret": "***",
  "username": "admin@company.com",
  "password": "***",
  "security_token": "***"
}
```

---

## IMAP Scanner

**File**: `services/imap_scanner.go`

### Features

- Connects to IMAP email servers
- Scans email subjects and bodies
- Supports SSL/TLS

### Connection Details

```json
{
  "host": "imap.gmail.com",
  "port": 993,
  "username": "mailbox@company.com",
  "password": "***",
  "use_tls": true,
  "folders": ["INBOX", "Sent"],
  "max_messages": 500
}
```

---

## MongoDB Scanner

**File**: `services/mongodb_scanner.go`

### Features

- Scans MongoDB collections
- Analyzes document field structure
- Samples random documents

### Connection Details

```json
{
  "connection_string": "mongodb://localhost:27017",
  "database": "production",
  "auth_database": "admin",
  "username": "reader",
  "password": "***"
}
```

---

## Common Features

### Connection Testing

All scanners support connection testing before saving:

```go
func (s *DatabaseScannerService) TestConnection(dsType string, connDetailsJSON json.RawMessage) error {
    switch dsType {
    case "postgresql":
        return s.testPostgreSQLConnection(connDetails)
    case "mysql":
        return s.testMySQLConnection(connDetails)
    default:
        return fmt.Errorf("unsupported data source type: %s", dsType)
    }
}
```

### Credential Encryption

Connection details are encrypted at rest using AES-256:

```go
func (s *EncryptionService) EncryptConnectionDetails(details interface{}) (string, error)
func (s *EncryptionService) DecryptConnectionDetails(encrypted string) (interface{}, error)
```

### Scan Scheduling

Scans can be scheduled via cron expressions:

```yaml
scanning:
  schedules:
    - data_source_id: 1
      cron: "0 2 * * *"  # Daily at 2 AM
    - data_source_id: 2
      cron: "0 */6 * * *"  # Every 6 hours
```

---

## Error Handling

### Common Errors

| Error | Cause | Resolution |
|-------|-------|------------|
| `connection refused` | Database not reachable | Check network/firewall |
| `authentication failed` | Wrong credentials | Update connection details |
| `permission denied` | Insufficient privileges | Grant SELECT permission |
| `timeout` | Large tables/slow network | Increase timeout settings |
| `unsupported type` | Unknown data source | Check supported types |
