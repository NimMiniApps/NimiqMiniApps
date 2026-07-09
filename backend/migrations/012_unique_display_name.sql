CREATE UNIQUE INDEX IF NOT EXISTS users_display_name_lower_unique
    ON users (LOWER(display_name))
    WHERE display_name IS NOT NULL;
