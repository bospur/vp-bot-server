CREATE TABLE animals (
    id        SERIAL PRIMARY KEY,
    name      VARCHAR(100) NOT NULL,  -- название животного (кошка, собака...)
    slug      VARCHAR(100) NOT NULL UNIQUE,  -- url-friendly имя (cat, dog...)
    icon      VARCHAR(10),            -- эмодзи или иконка
    sort_order INT NOT NULL DEFAULT 0, -- порядок отображения в списке
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
