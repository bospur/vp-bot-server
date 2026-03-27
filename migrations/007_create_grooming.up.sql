-- Коллекция пород/услуг грумера
-- day_of_week соглашение как в doctor_schedules: 0=Вс, 1=Пн, 2=Вт, 3=Ср, 4=Чт, 5=Пт, 6=Сб
CREATE TABLE grooming_breeds (
    id          SERIAL PRIMARY KEY,
    clinic_id   INT             NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    breed       VARCHAR(255)    NOT NULL,
    duration    INT             NOT NULL CHECK (duration > 0), -- минуты стрижки
    price       NUMERIC(10, 2),                                -- опционально
    description TEXT,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Шаблон рабочей недели грумера (один ряд = один рабочий день)
CREATE TABLE grooming_weekly_template (
    id          SERIAL PRIMARY KEY,
    clinic_id   INT         NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    day_of_week SMALLINT    NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    time_from   TIME        NOT NULL,
    time_to     TIME        NOT NULL,
    CONSTRAINT  chk_grooming_time CHECK (time_from < time_to),
    UNIQUE (clinic_id, day_of_week)
);

-- Записи животных на конкретную дату
CREATE TABLE grooming_appointments (
    id          SERIAL PRIMARY KEY,
    clinic_id   INT             NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    breed_id    INT             NOT NULL REFERENCES grooming_breeds(id) ON DELETE RESTRICT,
    date        DATE            NOT NULL,
    pet_name    VARCHAR(255)    NOT NULL,
    owner_phone VARCHAR(50)     NOT NULL DEFAULT '',
    start_time  TIME            NOT NULL,
    end_time    TIME            NOT NULL,
    CONSTRAINT  chk_appointment_time CHECK (start_time < end_time),
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Индексы для быстрого поиска по дате и проверки пересечений
CREATE INDEX idx_grooming_appointments_date ON grooming_appointments (clinic_id, date);
