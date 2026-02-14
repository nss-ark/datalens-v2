export interface DetectedPattern {
    pattern_name: string;
    confidence_score: number; // 0-100 or 0.0-1.0, assumed 0-100 based on context
    description: string;
    location: string; // e.g., "Line 5", "Button Text"
    regulation_clause: string; // e.g., "Guidelines 2023, Clause 2(1)(e)"
    fix_suggestion: string;
}

export interface DarkPatternAnalysisResult {
    overall_score: number; // 0-100
    detected_patterns: DetectedPattern[];
    analyzed_at: string;
}

export interface AnalyzeContentRequest {
    content: string;
    type: 'TEXT' | 'CODE' | 'HTML';
}
