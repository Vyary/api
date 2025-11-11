CREATE TABLE users (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  access_token TEXT,
  expires_in INTEGER,
  token_type TEXT,
  scope TEXT,
  sub TEXT,
  role TEXT DEFAULT 'user'
);

CREATE TABLE refresh_tokens (user_id TEXT, token_id TEXT, expires_at TEXT);

CREATE TABLE stats (id TEXT PRIMARY KEY, text TEXT, type TEXT);

CREATE TABLE items (
  id TEXT PRIMARY KEY DEFAULT (gen_random_uuid ()),
  realm TEXT,
  category TEXT DEFAULT '',
  sub_category TEXT DEFAULT '',
  icon TEXT DEFAULT '',
  icon_tier_text TEXT DEFAULT '',
  name TEXT DEFAULT '',
  base_type TEXT DEFAULT '',
  rarity TEXT DEFAULT '',
  w INTEGER DEFAULT 0,
  h INTEGER DEFAULT 0,
  ilvl INTEGER DEFAULT 0,
  socketed_items BLOB,
  properties BLOB,
  requirements BLOB,
  enchant_mods BLOB,
  rune_mods BLOB,
  implicit_mods BLOB,
  explicit_mods BLOB,
  fractured_mods BLOB,
  desecrated_mods BLOB,
  flavour_text TEXT DEFAULT '',
  descr_text TEXT DEFAULT '',
  sec_descr_text TEXT DEFAULT '',
  support BOOLEAN DEFAULT 0,
  duplicated BOOLEAN DEFAULT 0,
  corrupted BOOLEAN DEFAULT 0,
  sanctified BOOLEAN DEFAULT 0,
  desecrated BOOLEAN DEFAULT 0,
  user_id TEXT,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE prices (
  id INTEGER PRIMARY KEY,
  item_id TEXT,
  price REAL,
  currency_id TEXT,
  volume REAL,
  stock REAL,
  league TEXT,
  timestamp INTEGER
);

CREATE TABLE queries (
  id INTEGER PRIMARY KEY,
  item_id TEXT,
  realm TEXT,
  league TEXT,
  search_query TEXT,
  update_interval INTEGER,
  next_run INTEGER,
  status TEXT CHECK (status IN ('queued', 'in_progress')) DEFAULT 'queued',
  started_at INTEGER,
  run_once BOOLEAN,
  FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE
);
