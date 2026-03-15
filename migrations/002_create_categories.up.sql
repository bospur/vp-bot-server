CREATE TABLE categories (
    id         SERIAL PRIMARY KEY,
    animal_id  INT NOT NULL REFERENCES animals(id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,  -- название категории (Отравление, Травма...)
    slug       VARCHAR(100) NOT NULL,  -- url-friendly имя
    icon       VARCHAR(10),            -- эмодзи или иконка
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (animal_id, slug)           -- slug уникален в рамках одного животного
);
