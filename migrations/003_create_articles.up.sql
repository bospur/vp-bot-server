CREATE TABLE articles (
    id         SERIAL PRIMARY KEY,
    title      VARCHAR(255) NOT NULL,  -- заголовок статьи
    content    TEXT NOT NULL,          -- содержимое (шаги первой помощи)
    slug       VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Связующая таблица many-to-many между статьями и категориями
CREATE TABLE article_categories (
    article_id  INT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, category_id)  -- пара уникальна
);
