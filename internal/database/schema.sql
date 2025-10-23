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

CREATE TABLE items (
  id TEXT PRIMARY KEY DEFAULT (gen_random_uuid()),
  realm TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT '',
  sub_category TEXT NOT NULL DEFAULT '',
  icon TEXT NOT NULL DEFAULT '',
  icon_tier_text TEXT NOT NULL DEFAULT '',
  name TEXT NOT NULL DEFAULT '',
  base_type TEXT NOT NULL DEFAULT '',
  rarity TEXT NOT NULL DEFAULT '',
  w INTEGER NOT NULL DEFAULT 0,
  h INTEGER NOT NULL DEFAULT 0,
  ilvl INTEGER NOT NULL DEFAULT 0,
  sockets_count INTEGER NOT NULL DEFAULT 0,
  properties BLOB NOT NULL DEFAULT X'',
  requirements BLOB NOT NULL DEFAULT X'',
  enchant_mods BLOB NOT NULL DEFAULT X'',
  rune_mods BLOB NOT NULL DEFAULT X'',
  implicit_mods BLOB NOT NULL DEFAULT X'',
  explicit_mods BLOB NOT NULL DEFAULT X'',
  fractured_mods BLOB NOT NULL DEFAULT X'',
  desecrated_mods BLOB NOT NULL DEFAULT X'',
  flavour_text TEXT NOT NULL DEFAULT '',
  descr_text TEXT NOT NULL DEFAULT '',
  sec_descr_text TEXT NOT NULL DEFAULT '',
  support BOOLEAN NOT NULL DEFAULT 0,
  duplicated BOOLEAN NOT NULL DEFAULT 0,
  corrupted BOOLEAN NOT NULL DEFAULT 0,
  sanctified BOOLEAN NOT NULL DEFAULT 0,
  desecrated BOOLEAN NOT NULL DEFAULT 0,
  buy BLOB NOT NULL DEFAULT X'',
  sell BLOB NOT NULL DEFAULT X'',
  user_id TEXT,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE prices (
  id INTEGER PRIMARY KEY,
  want_id TEXT,
  want_count DECIMAL(10,4),
  have_count DECIMAL(10,4),
  have_id TEXT,
  league TEXT,
  stock INTEGER,
  volume_traded INTEGER,
  timestamp TEXT,
  FOREIGN KEY (want_id) REFERENCES items(id) ON DELETE CASCADE
);

CREATE TABLE queries (
  id INTEGER PRIMARY KEY,
  item_id TEXT,
  realm TEXT,
  league TEXT,
  search_query TEXT,
  update_interval INTEGER,
  next_run INTEGER,
  status TEXT CHECK(status IN ('queued', 'in_progress')) DEFAULT 'queued',
  started_at INTEGER,
  run_once BOOLEAN,
  FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);
