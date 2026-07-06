-- Section 7.4 : la migration « up » (golang-migrate : deux fichiers par migration)
CREATE TABLE IF NOT EXISTS authors (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);
