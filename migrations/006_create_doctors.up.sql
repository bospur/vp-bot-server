-- Таблица врачей
CREATE TABLE doctors (
    id          SERIAL PRIMARY KEY,
    clinic_id   INT          NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    full_name   VARCHAR(255) NOT NULL,
    specialty   VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT         NOT NULL DEFAULT '',
    contacts    TEXT         NOT NULL DEFAULT '',
    photo_url   VARCHAR(500) NOT NULL DEFAULT '',
    status      VARCHAR(20)  NOT NULL DEFAULT 'draft',
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Еженедельное расписание врача (повторяющиеся слоты)
-- day_of_week: 0=Вс, 1=Пн, 2=Вт, 3=Ср, 4=Чт, 5=Пт, 6=Сб
CREATE TABLE doctor_schedules (
    id          SERIAL PRIMARY KEY,
    doctor_id   INT         NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    day_of_week SMALLINT    NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    time_from   TIME        NOT NULL,
    time_to     TIME        NOT NULL,
    UNIQUE (doctor_id, day_of_week, time_from)
);

-- Исключения: замены или выходные на конкретную дату
CREATE TABLE doctor_schedule_exceptions (
    id          SERIAL PRIMARY KEY,
    doctor_id   INT         NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    date        DATE        NOT NULL,
    is_day_off  BOOLEAN     NOT NULL DEFAULT false,
    time_from   TIME,
    time_to     TIME,
    UNIQUE (doctor_id, date)
);

-- Настройки клиники
CREATE TABLE clinic_settings (
    clinic_id              INT     PRIMARY KEY REFERENCES clinics(id) ON DELETE CASCADE,
    schedule_display_weeks SMALLINT NOT NULL DEFAULT 2 CHECK (schedule_display_weeks BETWEEN 1 AND 5)
);

-- Инициализируем настройки для существующих клиник
INSERT INTO clinic_settings (clinic_id)
SELECT id FROM clinics
ON CONFLICT DO NOTHING;
