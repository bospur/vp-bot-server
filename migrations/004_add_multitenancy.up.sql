-- ── Клиники ──────────────────────────────────────────────────────────────────
CREATE TABLE clinics (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Вставляем дефолтную клинику для существующих данных
INSERT INTO clinics (name, slug) VALUES ('Default', 'default');

-- ── Пользователи ─────────────────────────────────────────────────────────────
CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    clinic_id     INT NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    login         VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'editor', -- admin, editor, viewer
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Добавляем clinic_id к существующим таблицам ───────────────────────────────

-- animals
ALTER TABLE animals ADD COLUMN clinic_id INT REFERENCES clinics(id) ON DELETE CASCADE;
UPDATE animals SET clinic_id = (SELECT id FROM clinics WHERE slug = 'default');
ALTER TABLE animals ALTER COLUMN clinic_id SET NOT NULL;

-- categories (наследуют клинику через animal, но для удобства дублируем)
ALTER TABLE categories ADD COLUMN clinic_id INT REFERENCES clinics(id) ON DELETE CASCADE;
UPDATE categories SET clinic_id = (SELECT id FROM clinics WHERE slug = 'default');
ALTER TABLE categories ALTER COLUMN clinic_id SET NOT NULL;

-- articles
ALTER TABLE articles ADD COLUMN clinic_id INT REFERENCES clinics(id) ON DELETE CASCADE;
UPDATE articles SET clinic_id = (SELECT id FROM clinics WHERE slug = 'default');
ALTER TABLE articles ALTER COLUMN clinic_id SET NOT NULL;
