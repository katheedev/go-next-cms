package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-next-cms/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ DB *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repository { return &Repository{DB: db} }

func (r *Repository) Countries(ctx context.Context) ([]models.Country, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, code, name, default_language FROM countries ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Country{}
	for rows.Next() {
		var c models.Country
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.DefaultLanguage); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) CountryByCode(ctx context.Context, code string) (*models.Country, error) {
	var c models.Country
	err := r.DB.QueryRow(ctx, `SELECT id, code, name, default_language FROM countries WHERE code=$1`, strings.ToUpper(code)).Scan(&c.ID, &c.Code, &c.Name, &c.DefaultLanguage)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) CitiesByCountry(ctx context.Context, countryID int64) ([]models.City, error) {
	rows, err := r.DB.Query(ctx, `SELECT id,country_id,name,slug FROM cities WHERE country_id=$1 ORDER BY name`, countryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.City
	for rows.Next() {
		var c models.City
		if err := rows.Scan(&c.ID, &c.CountryID, &c.Name, &c.Slug); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) Categories(ctx context.Context) ([]models.Category, error) {
	rows, err := r.DB.Query(ctx, `SELECT id,name,slug FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) Merchants(ctx context.Context) ([]models.Merchant, error) {
	rows, err := r.DB.Query(ctx, `SELECT id,name,slug,logo_url,contact,verified FROM merchants ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Merchant
	for rows.Next() {
		var m models.Merchant
		if err := rows.Scan(&m.ID, &m.Name, &m.Slug, &m.LogoURL, &m.Contact, &m.Verified); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) DealTypes(ctx context.Context) ([]models.DealType, error) {
	rows, err := r.DB.Query(ctx, `SELECT id,code,name FROM deal_types ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.DealType
	for rows.Next() {
		var d models.DealType
		if err := rows.Scan(&d.ID, &d.Code, &d.Name); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *Repository) ListDeals(ctx context.Context, f models.DealFilter) ([]models.Deal, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	where := []string{"d.status='published'", "d.end_at > NOW()", "co.code=$1"}
	args := []any{strings.ToUpper(f.CountryCode)}
	idx := 2
	if f.CitySlug != "" {
		where = append(where, fmt.Sprintf("ci.slug=$%d", idx))
		args = append(args, f.CitySlug)
		idx++
	}
	if f.CategorySlug != "" {
		where = append(where, fmt.Sprintf("ca.slug=$%d", idx))
		args = append(args, f.CategorySlug)
		idx++
	}
	if f.DealTypeID > 0 {
		where = append(where, fmt.Sprintf("d.deal_type_id=$%d", idx))
		args = append(args, f.DealTypeID)
		idx++
	}
	if f.MerchantID > 0 {
		where = append(where, fmt.Sprintf("d.merchant_id=$%d", idx))
		args = append(args, f.MerchantID)
		idx++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(LOWER(d.title) LIKE $%d OR LOWER(d.description) LIKE $%d)", idx, idx))
		args = append(args, "%"+strings.ToLower(f.Search)+"%")
		idx++
	}
	if f.EndingSoon {
		where = append(where, "d.end_at <= NOW() + INTERVAL '7 days'")
	}

	q := fmt.Sprintf(`SELECT d.id,d.title,d.slug,d.description,d.country_id,co.code, d.city_id,ci.name,d.category_id,ca.name,ca.slug,d.merchant_id,m.name,d.deal_type_id,dt.name,d.start_at,d.end_at,d.featured,d.image_url,d.status,d.created_by_user_id,d.rejection_reason,d.created_at,d.updated_at
	FROM deals d
	JOIN countries co ON co.id=d.country_id
	JOIN cities ci ON ci.id=d.city_id
	JOIN categories ca ON ca.id=d.category_id
	LEFT JOIN merchants m ON m.id=d.merchant_id
	JOIN deal_types dt ON dt.id=d.deal_type_id
	WHERE %s
	ORDER BY d.featured DESC, d.end_at ASC
	LIMIT %d OFFSET %d`, strings.Join(where, " AND "), f.PageSize, (f.Page-1)*f.PageSize)

	rows, err := r.DB.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDeals(rows)
}

func (r *Repository) FeaturedDeals(ctx context.Context, countryCode string, limit int) ([]models.Deal, error) {
	rows, err := r.DB.Query(ctx, `SELECT d.id,d.title,d.slug,d.description,d.country_id,co.code,d.city_id,ci.name,d.category_id,ca.name,ca.slug,d.merchant_id,m.name,d.deal_type_id,dt.name,d.start_at,d.end_at,d.featured,d.image_url,d.status,d.created_by_user_id,d.rejection_reason,d.created_at,d.updated_at
	FROM deals d
	JOIN countries co ON co.id=d.country_id
	JOIN cities ci ON ci.id=d.city_id
	JOIN categories ca ON ca.id=d.category_id
	LEFT JOIN merchants m ON m.id=d.merchant_id
	JOIN deal_types dt ON dt.id=d.deal_type_id
	WHERE d.status='published' AND d.end_at>NOW() AND d.featured=true AND co.code=$1 ORDER BY d.created_at DESC LIMIT $2`, strings.ToUpper(countryCode), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDeals(rows)
}

func (r *Repository) EndingSoonDeals(ctx context.Context, countryCode string, limit int) ([]models.Deal, error) {
	rows, err := r.DB.Query(ctx, `SELECT d.id,d.title,d.slug,d.description,d.country_id,co.code,d.city_id,ci.name,d.category_id,ca.name,ca.slug,d.merchant_id,m.name,d.deal_type_id,dt.name,d.start_at,d.end_at,d.featured,d.image_url,d.status,d.created_by_user_id,d.rejection_reason,d.created_at,d.updated_at
	FROM deals d
	JOIN countries co ON co.id=d.country_id
	JOIN cities ci ON ci.id=d.city_id
	JOIN categories ca ON ca.id=d.category_id
	LEFT JOIN merchants m ON m.id=d.merchant_id
	JOIN deal_types dt ON dt.id=d.deal_type_id
	WHERE d.status='published' AND d.end_at>NOW() AND d.end_at<=NOW()+INTERVAL '7 days' AND co.code=$1 ORDER BY d.end_at ASC LIMIT $2`, strings.ToUpper(countryCode), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDeals(rows)
}

func (r *Repository) DealBySlug(ctx context.Context, countryCode, slug string) (*models.Deal, error) {
	rows, err := r.DB.Query(ctx, `SELECT d.id,d.title,d.slug,d.description,d.country_id,co.code,d.city_id,ci.name,d.category_id,ca.name,ca.slug,d.merchant_id,m.name,d.deal_type_id,dt.name,d.start_at,d.end_at,d.featured,d.image_url,d.status,d.created_by_user_id,d.rejection_reason,d.created_at,d.updated_at
	FROM deals d
	JOIN countries co ON co.id=d.country_id
	JOIN cities ci ON ci.id=d.city_id
	JOIN categories ca ON ca.id=d.category_id
	LEFT JOIN merchants m ON m.id=d.merchant_id
	JOIN deal_types dt ON dt.id=d.deal_type_id
	WHERE d.slug=$1 AND co.code=$2 LIMIT 1`, slug, strings.ToUpper(countryCode))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ds, err := scanDeals(rows)
	if err != nil || len(ds) == 0 {
		if err == nil {
			err = pgx.ErrNoRows
		}
		return nil, err
	}
	return &ds[0], nil
}

func scanDeals(rows pgx.Rows) ([]models.Deal, error) {
	var out []models.Deal
	for rows.Next() {
		var d models.Deal
		if err := rows.Scan(&d.ID, &d.Title, &d.Slug, &d.Description, &d.CountryID, &d.CountryCode, &d.CityID, &d.CityName, &d.CategoryID, &d.CategoryName, &d.CategorySlug, &d.MerchantID, &d.MerchantName, &d.DealTypeID, &d.DealTypeName, &d.StartAt, &d.EndAt, &d.Featured, &d.ImageURL, &d.Status, &d.CreatedByUserID, &d.RejectionReason, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *Repository) DealTranslation(ctx context.Context, dealID int64, lang string) (*models.DealTranslation, error) {
	var t models.DealTranslation
	err := r.DB.QueryRow(ctx, `SELECT deal_id,lang,title,description FROM deal_translations WHERE deal_id=$1 AND lang=$2`, dealID, lang).Scan(&t.DealID, &t.Lang, &t.Title, &t.Description)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) CreateUser(ctx context.Context, u *models.User) error {
	return r.DB.QueryRow(ctx, `INSERT INTO users (email,password_hash,name,role) VALUES ($1,$2,$3,$4) RETURNING id,created_at`, u.Email, u.PasswordHash, u.Name, u.Role).Scan(&u.ID, &u.CreatedAt)
}

func (r *Repository) UserByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.DB.QueryRow(ctx, `SELECT id,email,password_hash,name,role,created_at FROM users WHERE email=$1`, strings.ToLower(email)).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UserByID(ctx context.Context, id int64) (*models.User, error) {
	var u models.User
	err := r.DB.QueryRow(ctx, `SELECT id,email,password_hash,name,role,created_at FROM users WHERE id=$1`, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateDeal(ctx context.Context, d *models.Deal, translations []models.DealTranslation) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `INSERT INTO deals (title,slug,description,country_id,city_id,category_id,merchant_id,deal_type_id,start_at,end_at,featured,image_url,status,created_by_user_id)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id,created_at,updated_at`, d.Title, d.Slug, d.Description, d.CountryID, d.CityID, d.CategoryID, d.MerchantID, d.DealTypeID, d.StartAt, d.EndAt, d.Featured, d.ImageURL, d.Status, d.CreatedByUserID).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return err
	}
	for _, t := range translations {
		_, err = tx.Exec(ctx, `INSERT INTO deal_translations (deal_id,lang,title,description) VALUES ($1,$2,$3,$4)`, d.ID, t.Lang, t.Title, t.Description)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) SubmissionDeals(ctx context.Context, userID int64) ([]models.Deal, error) {
	rows, err := r.DB.Query(ctx, `SELECT d.id,d.title,d.slug,d.description,d.country_id,co.code,d.city_id,ci.name,d.category_id,ca.name,ca.slug,d.merchant_id,m.name,d.deal_type_id,dt.name,d.start_at,d.end_at,d.featured,d.image_url,d.status,d.created_by_user_id,d.rejection_reason,d.created_at,d.updated_at
	FROM deals d
	JOIN countries co ON co.id=d.country_id
	JOIN cities ci ON ci.id=d.city_id
	JOIN categories ca ON ca.id=d.category_id
	LEFT JOIN merchants m ON m.id=d.merchant_id
	JOIN deal_types dt ON dt.id=d.deal_type_id
	WHERE d.created_by_user_id=$1 ORDER BY d.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDeals(rows)
}

func (r *Repository) UpdateSubmission(ctx context.Context, d *models.Deal) error {
	_, err := r.DB.Exec(ctx, `UPDATE deals SET title=$1,description=$2,city_id=$3,category_id=$4,merchant_id=$5,deal_type_id=$6,start_at=$7,end_at=$8,image_url=$9,status='pending',updated_at=NOW() WHERE id=$10`,
		d.Title, d.Description, d.CityID, d.CategoryID, d.MerchantID, d.DealTypeID, d.StartAt, d.EndAt, d.ImageURL, d.ID)
	return err
}

func (r *Repository) DealByID(ctx context.Context, id int64) (*models.Deal, error) {
	rows, err := r.DB.Query(ctx, `SELECT d.id,d.title,d.slug,d.description,d.country_id,co.code,d.city_id,ci.name,d.category_id,ca.name,ca.slug,d.merchant_id,m.name,d.deal_type_id,dt.name,d.start_at,d.end_at,d.featured,d.image_url,d.status,d.created_by_user_id,d.rejection_reason,d.created_at,d.updated_at
	FROM deals d
	JOIN countries co ON co.id=d.country_id
	JOIN cities ci ON ci.id=d.city_id
	JOIN categories ca ON ca.id=d.category_id
	LEFT JOIN merchants m ON m.id=d.merchant_id
	JOIN deal_types dt ON dt.id=d.deal_type_id
	WHERE d.id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ds, err := scanDeals(rows)
	if err != nil || len(ds) == 0 {
		if err == nil {
			err = pgx.ErrNoRows
		}
		return nil, err
	}
	return &ds[0], nil
}

func (r *Repository) PendingDeals(ctx context.Context) ([]models.Deal, error) {
	rows, err := r.DB.Query(ctx, `SELECT d.id,d.title,d.slug,d.description,d.country_id,co.code,d.city_id,ci.name,d.category_id,ca.name,ca.slug,d.merchant_id,m.name,d.deal_type_id,dt.name,d.start_at,d.end_at,d.featured,d.image_url,d.status,d.created_by_user_id,d.rejection_reason,d.created_at,d.updated_at
	FROM deals d
	JOIN countries co ON co.id=d.country_id
	JOIN cities ci ON ci.id=d.city_id
	JOIN categories ca ON ca.id=d.category_id
	LEFT JOIN merchants m ON m.id=d.merchant_id
	JOIN deal_types dt ON dt.id=d.deal_type_id
	WHERE d.status='pending' ORDER BY d.created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDeals(rows)
}

func (r *Repository) ModerateDeal(ctx context.Context, id int64, status models.DealStatus, reason *string) error {
	_, err := r.DB.Exec(ctx, `UPDATE deals SET status=$1,rejection_reason=$2,updated_at=NOW() WHERE id=$3`, status, reason, id)
	return err
}

func (r *Repository) DashboardCounts(ctx context.Context) (int, int, error) {
	var pending, published int
	if err := r.DB.QueryRow(ctx, `SELECT COUNT(*) FROM deals WHERE status='pending'`).Scan(&pending); err != nil {
		return 0, 0, err
	}
	if err := r.DB.QueryRow(ctx, `SELECT COUNT(*) FROM deals WHERE status='published'`).Scan(&published); err != nil {
		return 0, 0, err
	}
	return pending, published, nil
}

func (r *Repository) Users(ctx context.Context) ([]models.User, error) {
	rows, err := r.DB.Query(ctx, `SELECT id,email,password_hash,name,role,created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateUserRole(ctx context.Context, id int64, role models.UserRole) error {
	_, err := r.DB.Exec(ctx, `UPDATE users SET role=$1 WHERE id=$2`, role, id)
	return err
}

func (r *Repository) SetAdminConfig(ctx context.Context, key string, value string) error {
	_, err := r.DB.Exec(ctx, `INSERT INTO admin_config (key,value) VALUES ($1,$2::jsonb) ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value`, key, value)
	return err
}

func (r *Repository) AdminConfigs(ctx context.Context) ([]models.AdminConfig, error) {
	rows, err := r.DB.Query(ctx, `SELECT key, value::text FROM admin_config ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.AdminConfig
	for rows.Next() {
		var c models.AdminConfig
		if err := rows.Scan(&c.Key, &c.Value); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) SlugExists(ctx context.Context, countryID int64, slug string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM deals WHERE country_id=$1 AND slug=$2)`, countryID, slug).Scan(&exists)
	return exists, err
}

func (r *Repository) UpsertTranslation(ctx context.Context, dealID int64, lang, title, description string) error {
	_, err := r.DB.Exec(ctx, `INSERT INTO deal_translations (deal_id,lang,title,description) VALUES ($1,$2,$3,$4)
	ON CONFLICT (deal_id,lang) DO UPDATE SET title=EXCLUDED.title,description=EXCLUDED.description`, dealID, lang, title, description)
	return err
}

func (r *Repository) CreateCountry(ctx context.Context, c *models.Country) error {
	return r.DB.QueryRow(ctx, `INSERT INTO countries (code,name,default_language) VALUES ($1,$2,$3) RETURNING id`, c.Code, c.Name, c.DefaultLanguage).Scan(&c.ID)
}

func (r *Repository) CreateCity(ctx context.Context, c *models.City) error {
	return r.DB.QueryRow(ctx, `INSERT INTO cities (country_id,name,slug) VALUES ($1,$2,$3) RETURNING id`, c.CountryID, c.Name, c.Slug).Scan(&c.ID)
}

func (r *Repository) CreateCategory(ctx context.Context, c *models.Category) error {
	return r.DB.QueryRow(ctx, `INSERT INTO categories (name,slug) VALUES ($1,$2) RETURNING id`, c.Name, c.Slug).Scan(&c.ID)
}

func (r *Repository) CreateMerchant(ctx context.Context, m *models.Merchant) error {
	return r.DB.QueryRow(ctx, `INSERT INTO merchants (name,slug,logo_url,contact,verified) VALUES ($1,$2,$3,$4,$5) RETURNING id`, m.Name, m.Slug, m.LogoURL, m.Contact, m.Verified).Scan(&m.ID)
}

func (r *Repository) CreateDealType(ctx context.Context, dt *models.DealType) error {
	return r.DB.QueryRow(ctx, `INSERT INTO deal_types (code,name) VALUES ($1,$2) RETURNING id`, dt.Code, dt.Name).Scan(&dt.ID)
}

func (r *Repository) UpdateDealStatusDirect(ctx context.Context, id int64, status models.DealStatus) error {
	_, err := r.DB.Exec(ctx, `UPDATE deals SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	return err
}

func (r *Repository) TouchExpiredDeals(ctx context.Context) error {
	_, err := r.DB.Exec(ctx, `UPDATE deals SET status='expired' WHERE status='published' AND end_at < NOW()`)
	return err
}

func DateNowPlus(days int) time.Time { return time.Now().AddDate(0, 0, days) }
