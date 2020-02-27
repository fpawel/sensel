PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS measurement
(
    measurement_id INTEGER PRIMARY KEY NOT NULL,
    created_at     TIMESTAMP           NOT NULL DEFAULT (datetime('now', 'localtime')) UNIQUE,
    product_type   TEXT                NOT NULL DEFAULT '',
    name           TEXT                NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS measurement_pgs
(
    measurement_id INTEGER NOT NULL,
    gas            INTEGER NOT NULL,
    value          REAL    NOT NULL CHECK ( value >= 0 ),
    PRIMARY KEY (measurement_id, gas),
    CONSTRAINT measurement_concentration_foreign_key
        FOREIGN KEY (measurement_id) REFERENCES measurement (measurement_id)
            ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sample
(
    sample_id      INTEGER PRIMARY KEY NOT NULL,
    measurement_id INTEGER             NOT NULL,
    created_at     TIMESTAMP           NOT NULL DEFAULT (datetime('now', 'localtime')) UNIQUE,
    name           TEXT                NOT NULL CHECK ( name != '' ),
    gas            INTEGER             NOT NULL CHECK ( gas > 0 ),
    consumption    REAL                NOT NULL CHECK ( consumption >= 0 ),
    temperature    REAL                NOT NULL,
    current        REAL                NOT NULL CHECK ( current >= 0 ),
    CONSTRAINT sample_measurement_foreign_key
        FOREIGN KEY (measurement_id) REFERENCES measurement (measurement_id)
            ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS production
(
    place     INTEGER NOT NULL CHECK ( place BETWEEN 0 AND 15),
    sample_id INTEGER NOT NULL,
    break     BOOLEAN NOT NULL CHECK ( break IN (0, 1) ),
    value     REAL    NOT NULL,

    PRIMARY KEY (sample_id, place),
    CONSTRAINT product_value_sample_foreign_key
        FOREIGN KEY (sample_id) REFERENCES sample (sample_id)
            ON DELETE CASCADE

);

CREATE TABLE IF NOT EXISTS sample_log
(
    created_at TIMESTAMP NOT NULL DEFAULT (datetime('now', 'localtime')) UNIQUE,
    sample_id  INTEGER   NOT NULL,
    ok         BOOLEAN   NOT NULL CHECK ( ok IN (0, 1) ),
    message    TEXT      NOT NULL CHECK ( message != '' ),
    CONSTRAINT sample_log_foreign_key
        FOREIGN KEY (sample_id) REFERENCES sample (sample_id)
            ON DELETE CASCADE
);