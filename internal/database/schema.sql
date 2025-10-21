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
  realm TEXT,
  w INTEGER,
  h INTEGER,
  icon TEXT,
  name TEXT,
  base_type TEXT,
  category TEXT,
  rarity TEXT,
  support BOOLEAN,
  desecrated BOOLEAN,
  properties TEXT,
  requirements TEXT,
  enchant_mods TEXT,
  rune_mods TEXT,
  implicit_mods TEXT,
  explicit_mods TEXT,
  fractured_mods TEXT,
  desecrated_mods TEXT,
  flavour_text TEXT,
  descr_text TEXT,
  sec_descr_text TEXT,
  icon_tier_text TEXT,
  gem_sockets INTEGER,
  user_id TEXT,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE prices (
  id INTEGER PRIMARY KEY,
  item_id TEXT,
  league TEXT,
  value NUMERIC(10,4),
  currency TEXT,
  stock INTEGER,
  volume_traded INTEGER,
  timestamp TEXT,
  FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
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
