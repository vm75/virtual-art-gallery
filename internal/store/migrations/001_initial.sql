CREATE TABLE artworks (
    id INTEGER PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    date TEXT NOT NULL,
    surface TEXT NOT NULL,
    medium TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    alt_text TEXT NOT NULL DEFAULT '',
    visible INTEGER NOT NULL DEFAULT 0 CHECK (visible IN (0, 1)),
    source_name TEXT NOT NULL DEFAULT '',
    image_width INTEGER NOT NULL DEFAULT 0,
    image_height INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tags (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE artwork_tags (
    artwork_id INTEGER NOT NULL REFERENCES artworks(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (artwork_id, tag_id)
);

CREATE TABLE admin (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    username TEXT NOT NULL UNIQUE,
    password_hash BLOB NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    token_hash BLOB PRIMARY KEY,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE museum_rule_sets (
    id INTEGER PRIMARY KEY,
    version INTEGER NOT NULL,
    state TEXT NOT NULL CHECK (state IN ('draft', 'published')),
    seed INTEGER NOT NULL,
    rules_json TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
