# 11. Frontend Components

## Overview

The DataLens Control Centre Frontend is built with React 18, TypeScript, and Vite. It provides a comprehensive UI for compliance management.

---

## Application Structure

```
frontend/src/
├── pages/                    # Page components (56+)
│   ├── Dashboard.tsx
│   ├── Agents.tsx
│   ├── AgentDetails.tsx
│   ├── PIIInventory.tsx
│   ├── PIIReviewQueue.tsx
│   ├── DataSources.tsx
│   ├── DataSubjects.tsx
│   ├── DataLineage.tsx
│   ├── DSRRequests.tsx
│   ├── DSRDetails.tsx
│   ├── ConsentRecords.tsx
│   ├── ConsentAnalytics.tsx
│   ├── Grievances.tsx
│   ├── NominationClaims.tsx
│   ├── RoPA.tsx
│   ├── Reports.tsx
│   ├── Purposes.tsx
│   ├── DepartmentsManagement.tsx
│   ├── ThirdParties.tsx
│   ├── UserManagement.tsx
│   ├── AuditLogs.tsx
│   ├── Settings.tsx
│   ├── Login.tsx
│   ├── DataPrincipalPortal/      # Public-facing portal
│   │   ├── PortalDashboard.tsx
│   │   ├── SubmitDSR.tsx
│   │   ├── TrackDSR.tsx
│   │   ├── ConsentPreferences.tsx
│   │   └── ...
│   └── SuperAdmin/               # Platform admin
│       ├── ClientManagement.tsx
│       ├── SystemHealth.tsx
│       └── ...
├── components/               # Reusable components (16+)
│   ├── Layout/
│   ├── DataTable/
│   ├── Charts/
│   ├── Forms/
│   └── common/
├── services/                 # API clients (25+)
├── hooks/                    # Custom hooks
├── utils/                    # Utilities
├── types/                    # TypeScript types
└── App.tsx
```

---

## Page Categories

### Overview & Dashboard

| Page | Purpose | Key Features |
|------|---------|--------------|
| `Dashboard.tsx` | Main landing page | KPIs, recent activity, alerts |

### Agent Management

| Page | Purpose | Key Features |
|------|---------|--------------|
| `Agents.tsx` | List all agents | Status indicators, quick actions |
| `AgentDetails.tsx` | Single agent view | Config, data sources, scans |
| `DataSources.tsx` | Manage data sources | Add/edit/test connections |
| `ScanHistory.tsx` | Scan run history | Results, errors, duration |

### PII Management

| Page | Purpose | Key Features |
|------|---------|--------------|
| `PIIInventory.tsx` | Verified PII inventory | Filter, search, export |
| `PIIReviewQueue.tsx` | Pending verifications | Verify/reject workflow |
| `PIIDetails.tsx` | Single PII field | History, purposes, recipients |
| `DataLineage.tsx` | Data flow visualization | Interactive graph |

### Data Subjects

| Page | Purpose | Key Features |
|------|---------|--------------|
| `DataSubjects.tsx` | List all subjects | Search, filter by type |
| `DataSubjectDetails.tsx` | Single subject view | PII locations, consent, DSRs |

### DSR Management

| Page | Purpose | Key Features |
|------|---------|--------------|
| `DSRRequests.tsx` | List all DSRs | Status filters, SLA tracking |
| `DSRDetails.tsx` | Single DSR view | Timeline, affected data, actions |
| `DSRCreate.tsx` | Create new DSR | Form wizard |
| `DSRTemplates.tsx` | Response templates | Email/letter templates |

### Consent Management

| Page | Purpose | Key Features |
|------|---------|--------------|
| `ConsentRecords.tsx` | Consent history | Filter by purpose, status |
| `ConsentAnalytics.tsx` | Consent metrics | Charts, trends |
| `ConsentNotices.tsx` | Manage notices | Create/edit notices |

### Compliance

| Page | Purpose | Key Features |
|------|---------|--------------|
| `RoPA.tsx` | Records of Processing | Generate, export |
| `Reports.tsx` | Report dashboard | Generate, download |
| `AuditLogs.tsx` | Activity audit trail | Search, filter, export |

### Settings

| Page | Purpose | Key Features |
|------|---------|--------------|
| `Purposes.tsx` | Processing purposes | CRUD operations |
| `DepartmentsManagement.tsx` | Internal departments | CRUD operations |
| `ThirdParties.tsx` | External vendors | CRUD, DPA tracking |
| `RetentionPolicies.tsx` | Data retention | Configure policies |
| `UserManagement.tsx` | Team members | Invite, roles, permissions |
| `Settings.tsx` | General settings | Tenant configuration |

---

## Core Components

### Layout Components

```tsx
// Layout/Sidebar.tsx
// Main navigation sidebar
<Sidebar>
  <Logo />
  <NavGroup title="Overview">
    <NavItem to="/dashboard" icon={<DashboardIcon />}>Dashboard</NavItem>
  </NavGroup>
  <NavGroup title="Discovery">
    <NavItem to="/agents">Agents</NavItem>
    <NavItem to="/pii/inventory">PII Inventory</NavItem>
    <NavItem to="/pii/review">Review Queue</NavItem>
  </NavGroup>
  {/* More groups... */}
</Sidebar>

// Layout/Header.tsx
// Top navigation bar
<Header>
  <SearchBar />
  <NotificationBell count={5} />
  <UserMenu user={currentUser} />
</Header>
```

### Data Table

```tsx
// components/DataTable/DataTable.tsx
// Reusable table with sorting, filtering, pagination

<DataTable
  columns={[
    { key: 'name', header: 'Name', sortable: true },
    { key: 'status', header: 'Status', render: StatusBadge },
    { key: 'created_at', header: 'Created', format: 'date' },
    { key: 'actions', header: '', render: ActionsMenu }
  ]}
  data={items}
  loading={isLoading}
  pagination={{
    page: 1,
    pageSize: 20,
    total: 100,
    onPageChange: handlePageChange
  }}
  onSort={handleSort}
  onRowClick={handleRowClick}
/>
```

### Charts

```tsx
// components/Charts/
// Recharts-based visualizations

// PieChart for PII categories
<PIICategoryChart data={[
  { category: 'EMAIL_ADDRESS', count: 150 },
  { category: 'PHONE_NUMBER', count: 120 },
  // ...
]} />

// BarChart for DSR by status
<DSRStatusChart data={[
  { status: 'PENDING', count: 5 },
  { status: 'IN_PROGRESS', count: 12 },
  // ...
]} />

// LineChart for consent trends
<ConsentTrendChart
  data={monthlyData}
  purposes={['marketing', 'analytics']}
/>
```

### Forms

```tsx
// components/Forms/
// Reusable form components

<FormField
  label="Email"
  name="email"
  type="email"
  required
  error={errors.email}
/>

<SelectField
  label="Status"
  name="status"
  options={[
    { value: 'ACTIVE', label: 'Active' },
    { value: 'INACTIVE', label: 'Inactive' }
  ]}
/>

<DatePicker
  label="Due Date"
  name="due_date"
  minDate={today}
/>
```

### Status Badges

```tsx
// components/common/StatusBadge.tsx

<StatusBadge status="ACTIVE" />     // Green
<StatusBadge status="PENDING" />    // Yellow
<StatusBadge status="FAILED" />     // Red
<StatusBadge status="COMPLETED" />  // Blue
```

---

## Key Page Patterns

### List Page Pattern

```tsx
const AgentsPage = () => {
  const { data, isLoading, error } = useAgents();
  
  return (
    <PageLayout title="Agents">
      <PageHeader>
        <Button onClick={handleCreate}>Add Agent</Button>
      </PageHeader>
      
      <FilterBar>
        <SearchInput />
        <StatusFilter />
      </FilterBar>
      
      <DataTable
        columns={columns}
        data={data}
        loading={isLoading}
      />
    </PageLayout>
  );
};
```

### Detail Page Pattern

```tsx
const AgentDetailsPage = () => {
  const { id } = useParams();
  const { data: agent, isLoading } = useAgent(id);
  
  return (
    <PageLayout title={agent?.name}>
      <Breadcrumbs items={[
        { label: 'Agents', href: '/agents' },
        { label: agent?.name }
      ]} />
      
      <Tabs>
        <Tab label="Overview">
          <AgentOverview agent={agent} />
        </Tab>
        <Tab label="Data Sources">
          <DataSourcesList agentId={id} />
        </Tab>
        <Tab label="Scan History">
          <ScanHistoryList agentId={id} />
        </Tab>
      </Tabs>
    </PageLayout>
  );
};
```

### Form Page Pattern

```tsx
const CreateDSRPage = () => {
  const { mutate: createDSR, isLoading } = useCreateDSR();
  
  const handleSubmit = (values) => {
    createDSR(values, {
      onSuccess: () => navigate('/dsr')
    });
  };
  
  return (
    <PageLayout title="Create DSR Request">
      <Form onSubmit={handleSubmit}>
        <FormSection title="Request Type">
          <RadioGroup name="type" options={dsrTypes} />
        </FormSection>
        
        <FormSection title="Data Subject">
          <FormField name="email" label="Email" required />
          <FormField name="name" label="Name" />
        </FormSection>
        
        <FormActions>
          <Button variant="outline" onClick={() => navigate(-1)}>
            Cancel
          </Button>
          <Button type="submit" loading={isLoading}>
            Create Request
          </Button>
        </FormActions>
      </Form>
    </PageLayout>
  );
};
```

---

## API Services

```tsx
// services/api.ts
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
});

api.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// services/agents.ts
export const agentsApi = {
  list: () => api.get('/agents'),
  get: (id: string) => api.get(`/agents/${id}`),
  create: (data: CreateAgentDto) => api.post('/agents', data),
  update: (id: string, data: UpdateAgentDto) => api.put(`/agents/${id}`, data),
  delete: (id: string) => api.delete(`/agents/${id}`),
};

// hooks/useAgents.ts
export const useAgents = () => {
  return useQuery(['agents'], agentsApi.list);
};
```

---

## Data Principal Portal

The public-facing portal allows data subjects to:

| Feature | Page |
|---------|------|
| Submit DSR | `SubmitDSR.tsx` |
| Track DSR | `TrackDSR.tsx` |
| Manage consent | `ConsentPreferences.tsx` |
| Verify identity | `VerifyIdentity.tsx` |
| View data | `ViewMyData.tsx` |

---

## Super Admin Portal

Platform administration features:

| Feature | Page |
|---------|------|
| Manage tenants | `ClientManagement.tsx` |
| System health | `SystemHealth.tsx` |
| Global settings | `GlobalSettings.tsx` |
| Usage analytics | `UsageAnalytics.tsx` |
