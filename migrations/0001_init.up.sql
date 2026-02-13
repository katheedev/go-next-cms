CREATE TABLE countries (
  id BIGSERIAL PRIMARY KEY,
  code VARCHAR(2) UNIQUE NOT NULL,
  name TEXT NOT NULL,
  default_language VARCHAR(2) NOT NULL
);

CREATE TABLE cities (
  id BIGSERIAL PRIMARY KEY,
  country_id BIGINT NOT NULL REFERENCES countries(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  slug TEXT NOT NULL,
  UNIQUE(country_id, slug)
);

CREATE TABLE categories (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  slug TEXT UNIQUE NOT NULL
);

CREATE TABLE merchants (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  logo_url TEXT,
  contact TEXT,
  verified BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE deal_types (
  id BIGSERIAL PRIMARY KEY,
  code TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL
);

CREATE TYPE deal_status AS ENUM ('draft','pending','approved','published','rejected','expired');
CREATE TYPE user_role AS ENUM ('submitter','admin');

CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  name TEXT NOT NULL,
  role user_role NOT NULL DEFAULT 'submitter',
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE deals (
  id BIGSERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  slug TEXT NOT NULL,
  description TEXT NOT NULL,
  country_id BIGINT NOT NULL REFERENCES countries(id),
  city_id BIGINT NOT NULL REFERENCES cities(id),
  category_id BIGINT NOT NULL REFERENCES categories(id),
  merchant_id BIGINT REFERENCES merchants(id),
  deal_type_id BIGINT NOT NULL REFERENCES deal_types(id),
  start_at TIMESTAMP NOT NULL,
  end_at TIMESTAMP NOT NULL,
  featured BOOLEAN NOT NULL DEFAULT false,
  image_url TEXT,
  status deal_status NOT NULL DEFAULT 'draft',
  rejection_reason TEXT,
  created_by_user_id BIGINT NOT NULL REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE(country_id, slug)
);

CREATE TABLE deal_translations (
  deal_id BIGINT NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
  lang VARCHAR(2) NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  PRIMARY KEY (deal_id, lang)
);

CREATE TABLE admin_config (
  key TEXT PRIMARY KEY,
  value JSONB NOT NULL
);

CREATE INDEX idx_deals_country_status ON deals(country_id, status, end_at);
