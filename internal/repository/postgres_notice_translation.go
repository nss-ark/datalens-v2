package repository

import (
	"context"

	"github.com/complyark/datalens/internal/domain/consent"
	"github.com/complyark/datalens/pkg/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConsentNoticeTranslationRepository struct {
	db *pgxpool.Pool
}

func NewPostgresConsentNoticeTranslationRepository(db *pgxpool.Pool) *PostgresConsentNoticeTranslationRepository {
	return &PostgresConsentNoticeTranslationRepository{db: db}
}

func (r *PostgresConsentNoticeTranslationRepository) SaveTranslation(ctx context.Context, t *consent.ConsentNoticeTranslation) error {
	query := `INSERT INTO consent_notice_translations (
		id, notice_id, notice_version, language_code, translated_text, translation_source, is_rtl, translated_at, reviewed_by, reviewed_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Exec(ctx, query,
		t.ID, t.NoticeID, t.NoticeVersion, t.LanguageCode, t.TranslatedText, t.TranslationSource, t.IsRTL, t.TranslatedAt, t.ReviewedBy, t.ReviewedAt,
	)
	return err
}

func (r *PostgresConsentNoticeTranslationRepository) GetByNoticeAndVersion(ctx context.Context, noticeID types.ID, version int) ([]consent.ConsentNoticeTranslation, error) {
	query := `SELECT 
		id, notice_id, notice_version, language_code, translated_text, translation_source, is_rtl, translated_at, reviewed_by, reviewed_at
	FROM consent_notice_translations 
	WHERE notice_id = $1 AND notice_version = $2
	ORDER BY language_code ASC`

	rows, err := r.db.Query(ctx, query, noticeID, version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translations []consent.ConsentNoticeTranslation
	for rows.Next() {
		var t consent.ConsentNoticeTranslation
		if err := rows.Scan(
			&t.ID, &t.NoticeID, &t.NoticeVersion, &t.LanguageCode, &t.TranslatedText, &t.TranslationSource, &t.IsRTL, &t.TranslatedAt, &t.ReviewedBy, &t.ReviewedAt,
		); err != nil {
			return nil, err
		}
		translations = append(translations, t)
	}
	return translations, nil
}

func (r *PostgresConsentNoticeTranslationRepository) GetByNoticeAndLang(ctx context.Context, noticeID types.ID, version int, lang string) (*consent.ConsentNoticeTranslation, error) {
	query := `SELECT 
		id, notice_id, notice_version, language_code, translated_text, translation_source, is_rtl, translated_at, reviewed_by, reviewed_at
	FROM consent_notice_translations 
	WHERE notice_id = $1 AND notice_version = $2 AND language_code = $3`

	var t consent.ConsentNoticeTranslation
	err := r.db.QueryRow(ctx, query, noticeID, version, lang).Scan(
		&t.ID, &t.NoticeID, &t.NoticeVersion, &t.LanguageCode, &t.TranslatedText, &t.TranslationSource, &t.IsRTL, &t.TranslatedAt, &t.ReviewedBy, &t.ReviewedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil // Return nil if explicitly not found, handling logic in service
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PostgresConsentNoticeTranslationRepository) Upsert(ctx context.Context, t *consent.ConsentNoticeTranslation) error {
	query := `INSERT INTO consent_notice_translations (
		id, notice_id, notice_version, language_code, translated_text, translation_source, is_rtl, translated_at, reviewed_by, reviewed_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	ON CONFLICT (notice_id, notice_version, language_code) 
	DO UPDATE SET 
		translated_text = EXCLUDED.translated_text,
		translation_source = EXCLUDED.translation_source,
		is_rtl = EXCLUDED.is_rtl,
		reviewed_by = EXCLUDED.reviewed_by,
		reviewed_at = EXCLUDED.reviewed_at,
		translated_at = EXCLUDED.translated_at`

	_, err := r.db.Exec(ctx, query,
		t.ID, t.NoticeID, t.NoticeVersion, t.LanguageCode, t.TranslatedText, t.TranslationSource, t.IsRTL, t.TranslatedAt, t.ReviewedBy, t.ReviewedAt,
	)
	return err
}
