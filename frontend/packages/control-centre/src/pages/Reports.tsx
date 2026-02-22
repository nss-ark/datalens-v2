import { useQuery } from '@tanstack/react-query';
import {
    FileText, ShieldCheck, Building2, Globe, Download,
    AlertTriangle, CheckSquare, ClipboardList, Briefcase, Target
} from 'lucide-react';
import { Button } from '@datalens/shared';
import { reportService } from '../services/reportService';
import type { ComplianceSnapshot, ExportEntity } from '../services/reportService';

// ── Helpers ──────────────────────────────────────────────────────────────

function scoreColor(score: number): string {
    if (score >= 80) return '#10b981';
    if (score >= 60) return '#f59e0b';
    return '#ef4444';
}

function scoreLabel(score: number): string {
    if (score >= 80) return 'Excellent';
    if (score >= 60) return 'Fair';
    return 'Needs Attention';
}

// ── Circular Gauge ───────────────────────────────────────────────────────

function CircularGauge({ score, size = 180 }: { score: number; size?: number }) {
    const strokeWidth = 14;
    const radius = (size - strokeWidth) / 2;
    const circumference = 2 * Math.PI * radius;
    const offset = circumference - (score / 100) * circumference;
    const color = scoreColor(score);

    return (
        <div style={{ position: 'relative', width: size, height: size }}>
            <svg width={size} height={size} style={{ transform: 'rotate(-90deg)' }}>
                <circle
                    cx={size / 2} cy={size / 2} r={radius}
                    fill="none" stroke="#e5e7eb" strokeWidth={strokeWidth}
                />
                <circle
                    cx={size / 2} cy={size / 2} r={radius}
                    fill="none" stroke={color} strokeWidth={strokeWidth}
                    strokeDasharray={circumference}
                    strokeDashoffset={offset}
                    strokeLinecap="round"
                    style={{ transition: 'stroke-dashoffset 1s ease-in-out' }}
                />
            </svg>
            <div style={{
                position: 'absolute', inset: 0,
                display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
            }}>
                <span style={{ fontSize: '2.75rem', fontWeight: 800, color, lineHeight: 1 }}>{score}</span>
                <span style={{ fontSize: '0.875rem', color: '#6b7280', fontWeight: 500 }}>/100</span>
                <span style={{
                    marginTop: '4px', fontSize: '0.75rem', fontWeight: 600, color,
                    backgroundColor: `${color}18`, padding: '2px 10px', borderRadius: '9999px',
                }}>{scoreLabel(score)}</span>
            </div>
        </div>
    );
}

// ── Progress Bar ─────────────────────────────────────────────────────────

function ProgressBar({ value, max = 100 }: { value: number; max?: number }) {
    const pct = Math.min(100, (value / max) * 100);
    const color = scoreColor(value);
    return (
        <div style={{
            height: '6px', borderRadius: '3px', backgroundColor: '#e5e7eb',
            overflow: 'hidden', width: '100%',
        }}>
            <div style={{
                height: '100%', borderRadius: '3px', backgroundColor: color,
                width: `${pct}%`, transition: 'width 1s ease-in-out',
            }} />
        </div>
    );
}

// ── Pillar Card ──────────────────────────────────────────────────────────

interface PillarCardProps {
    title: string;
    score: number;
    Icon: React.ElementType;
    metric: string;
    metricValue: string | number;
}

function PillarCard({ title, score, Icon, metric, metricValue }: PillarCardProps) {
    const color = scoreColor(score);
    return (
        <div style={{
            background: '#fff', borderRadius: '16px', padding: '24px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04)',
            border: '1px solid #f3f4f6', display: 'flex', flexDirection: 'column', gap: '12px',
            transition: 'box-shadow 0.2s', cursor: 'default',
        }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                    <div style={{
                        width: 36, height: 36, borderRadius: '10px',
                        backgroundColor: `${color}18`, display: 'flex', alignItems: 'center', justifyContent: 'center',
                    }}>
                        <Icon size={18} style={{ color }} />
                    </div>
                    <span style={{ fontWeight: 600, fontSize: '0.9375rem', color: '#1f2937' }}>{title}</span>
                </div>
                <span style={{ fontSize: '1.5rem', fontWeight: 800, color }}>{score}</span>
            </div>
            <ProgressBar value={score} />
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8125rem' }}>
                <span style={{ color: '#6b7280' }}>{metric}</span>
                <span style={{ fontWeight: 600, color: '#374151' }}>{metricValue}</span>
            </div>
        </div>
    );
}

// ── Recommendation Card ──────────────────────────────────────────────────

function RecommendationCard({ text }: { text: string }) {
    // determine icon and color by keyword
    let Icon = AlertTriangle;
    let borderColor = '#f59e0b';
    let bgColor = '#fffbeb';
    const lower = text.toLowerCase();
    if (lower.includes('department')) { Icon = Building2; borderColor = '#3b82f6'; bgColor = '#eff6ff'; }
    else if (lower.includes('dsr') || lower.includes('request')) { Icon = FileText; borderColor = '#ef4444'; bgColor = '#fef2f2'; }
    else if (lower.includes('third-part') || lower.includes('dpa')) { Icon = Globe; borderColor = '#8b5cf6'; bgColor = '#f5f3ff'; }

    return (
        <div style={{
            display: 'flex', alignItems: 'center', gap: '14px',
            padding: '14px 18px', borderRadius: '12px',
            borderLeft: `4px solid ${borderColor}`, backgroundColor: bgColor,
        }}>
            <Icon size={20} style={{ color: borderColor, flexShrink: 0 }} />
            <span style={{ fontSize: '0.875rem', color: '#374151', lineHeight: 1.5 }}>{text}</span>
        </div>
    );
}

// ── Export Card ───────────────────────────────────────────────────────────

interface ExportCardProps {
    title: string;
    entity: ExportEntity;
    count?: number;
    Icon: React.ElementType;
    iconColor: string;
}

function ExportCard({ title, entity, count, Icon, iconColor }: ExportCardProps) {
    return (
        <div style={{
            background: '#fff', borderRadius: '14px', padding: '20px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.05)',
            border: '1px solid #f3f4f6', display: 'flex', flexDirection: 'column', gap: '12px',
        }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                <div style={{
                    width: 36, height: 36, borderRadius: '10px',
                    backgroundColor: `${iconColor}14`, display: 'flex', alignItems: 'center', justifyContent: 'center',
                }}>
                    <Icon size={18} style={{ color: iconColor }} />
                </div>
                <div>
                    <div style={{ fontWeight: 600, fontSize: '0.875rem', color: '#1f2937' }}>{title}</div>
                    {count !== undefined && (
                        <div style={{ fontSize: '0.75rem', color: '#6b7280' }}>{count} records</div>
                    )}
                </div>
            </div>
            <div style={{ display: 'flex', gap: '8px' }}>
                <Button
                    variant="outline" size="sm"
                    icon={<Download size={14} />}
                    onClick={() => reportService.exportEntity(entity, 'csv')}
                >
                    CSV
                </Button>
                <Button
                    variant="outline" size="sm"
                    icon={<Download size={14} />}
                    onClick={() => reportService.exportEntity(entity, 'json')}
                >
                    JSON
                </Button>
            </div>
        </div>
    );
}

// ── Export entities config ────────────────────────────────────────────────

function getExportCards(snapshot?: ComplianceSnapshot) {
    return [
        { title: 'DSR Requests', entity: 'dsr' as ExportEntity, Icon: FileText, iconColor: '#3b82f6', count: snapshot?.pillars.dsr_compliance.total_requests },
        { title: 'Breaches', entity: 'breaches' as ExportEntity, Icon: ShieldCheck, iconColor: '#ef4444', count: snapshot?.pillars.breach_management.total_incidents },
        { title: 'Consent Records', entity: 'consent-records' as ExportEntity, Icon: CheckSquare, iconColor: '#10b981', count: snapshot?.pillars.consent_management.total_consents },
        { title: 'Audit Logs', entity: 'audit-logs' as ExportEntity, Icon: ClipboardList, iconColor: '#6366f1', count: undefined },
        { title: 'Departments', entity: 'departments' as ExportEntity, Icon: Building2, iconColor: '#0ea5e9', count: snapshot?.pillars.data_governance.departments_total },
        { title: 'Third Parties', entity: 'third-parties' as ExportEntity, Icon: Globe, iconColor: '#8b5cf6', count: snapshot?.pillars.data_governance.third_parties_total },
        { title: 'Purposes', entity: 'purposes' as ExportEntity, Icon: Briefcase, iconColor: '#f59e0b', count: snapshot?.pillars.data_governance.purposes_mapped },
    ];
}

// ── Page ─────────────────────────────────────────────────────────────────

const Reports = () => {
    const { data: snapshot, isLoading, isError } = useQuery({
        queryKey: ['compliance-snapshot'],
        queryFn: () => reportService.getComplianceSnapshot(),
    });

    // Loading state
    if (isLoading) {
        return (
            <div style={{ padding: '32px', display: 'flex', flexDirection: 'column', gap: '24px' }}>
                <div style={{ height: 32, width: 240, borderRadius: 8, backgroundColor: '#f3f4f6' }} />
                <div style={{ display: 'flex', justifyContent: 'center' }}>
                    <div style={{ width: 180, height: 180, borderRadius: '50%', backgroundColor: '#f3f4f6' }} />
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16 }}>
                    {[1, 2, 3, 4].map(i => (
                        <div key={i} style={{ height: 120, borderRadius: 16, backgroundColor: '#f3f4f6' }} />
                    ))}
                </div>
            </div>
        );
    }

    // Error state
    if (isError) {
        return (
            <div style={{ padding: '32px', textAlign: 'center' }}>
                <AlertTriangle size={48} style={{ color: '#ef4444', marginBottom: 16 }} />
                <h2 style={{ color: '#1f2937', marginBottom: 8 }}>Failed to load compliance data</h2>
                <p style={{ color: '#6b7280' }}>Please try refreshing the page or contact support.</p>
            </div>
        );
    }

    const overall = snapshot?.overall_score ?? 0;
    const pillars = snapshot?.pillars;
    const recommendations = snapshot?.recommendations ?? [];
    const exportCards = getExportCards(snapshot);

    return (
        <div style={{ padding: '24px 32px', maxWidth: 1200, margin: '0 auto' }}>
            {/* Page Header */}
            <div style={{ marginBottom: '32px' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '4px' }}>
                    <Target size={24} style={{ color: '#3b82f6' }} />
                    <h1 style={{ fontSize: '1.5rem', fontWeight: 700, color: '#111827' }}>
                        Compliance Reports
                    </h1>
                </div>
                <p style={{ color: '#6b7280', fontSize: '0.875rem' }}>
                    DPDPA compliance scorecard, recommendations, and data exports
                </p>
            </div>

            {/* ── Section A: Compliance Scorecard ────────────────────────── */}
            <div style={{
                background: 'linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%)',
                borderRadius: '20px', padding: '36px',
                border: '1px solid #e2e8f0', marginBottom: '28px',
            }}>
                <h2 style={{
                    fontSize: '0.75rem', fontWeight: 700, letterSpacing: '0.08em',
                    textTransform: 'uppercase' as const, color: '#64748b', marginBottom: '24px',
                }}>
                    Compliance Scorecard
                </h2>

                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '28px' }}>
                    <CircularGauge score={overall} />

                    {snapshot?.generated_at && (
                        <span style={{ fontSize: '0.75rem', color: '#94a3b8' }}>
                            Last generated: {new Date(snapshot.generated_at).toLocaleString()}
                        </span>
                    )}
                </div>

                {/* Pillar Cards */}
                <div style={{
                    display: 'grid', gap: '16px', marginTop: '28px',
                    gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
                }}>
                    {pillars && (
                        <>
                            <PillarCard
                                title="Consent" score={pillars.consent_management.score}
                                Icon={CheckSquare}
                                metric="Active Consents" metricValue={pillars.consent_management.active_consents}
                            />
                            <PillarCard
                                title="DSR Compliance" score={pillars.dsr_compliance.score}
                                Icon={FileText}
                                metric="On-Time Rate"
                                metricValue={pillars.dsr_compliance.total_requests > 0
                                    ? `${Math.round((pillars.dsr_compliance.completed_on_time / pillars.dsr_compliance.total_requests) * 100)}%`
                                    : 'N/A'}
                            />
                            <PillarCard
                                title="Breach Mgmt" score={pillars.breach_management.score}
                                Icon={ShieldCheck}
                                metric="Total Incidents" metricValue={pillars.breach_management.total_incidents}
                            />
                            <PillarCard
                                title="Governance" score={pillars.data_governance.score}
                                Icon={Building2}
                                metric="Dept. Coverage"
                                metricValue={`${pillars.data_governance.departments_with_owners}/${pillars.data_governance.departments_total}`}
                            />
                        </>
                    )}
                </div>
            </div>

            {/* ── Section B: Recommendations ─────────────────────────────── */}
            {recommendations.length > 0 && (
                <div style={{ marginBottom: '28px' }}>
                    <h2 style={{
                        fontSize: '0.75rem', fontWeight: 700, letterSpacing: '0.08em',
                        textTransform: 'uppercase' as const, color: '#64748b', marginBottom: '14px',
                    }}>
                        Recommendations ({recommendations.length})
                    </h2>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                        {recommendations.map((rec, i) => (
                            <RecommendationCard key={i} text={rec} />
                        ))}
                    </div>
                </div>
            )}

            {/* ── Section C: Data Exports ─────────────────────────────────── */}
            <div>
                <h2 style={{
                    fontSize: '0.75rem', fontWeight: 700, letterSpacing: '0.08em',
                    textTransform: 'uppercase' as const, color: '#64748b', marginBottom: '14px',
                }}>
                    Data Exports
                </h2>
                <div style={{
                    display: 'grid', gap: '16px',
                    gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))',
                }}>
                    {exportCards.map(card => (
                        <ExportCard
                            key={card.entity}
                            title={card.title}
                            entity={card.entity}
                            Icon={card.Icon}
                            iconColor={card.iconColor}
                            count={card.count}
                        />
                    ))}
                </div>
            </div>
        </div>
    );
};

export default Reports;
