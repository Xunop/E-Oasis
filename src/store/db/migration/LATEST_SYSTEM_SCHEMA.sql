-- migration_history
CREATE TABLE migration_history (
  version TEXT NOT NULL PRIMARY KEY,
  created_ts BIGINT NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- system_setting
CREATE TABLE system_setting (
  name TEXT NOT NULL,
  value TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  UNIQUE(name)
);

-- user
CREATE TABLE user (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_ts BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
  updated_ts BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
  last_login_ts BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
  username TEXT NOT NULL UNIQUE,
  row_status TEXT NOT NULL CHECK (row_status IN ('NORMAL', 'ARCHIVED')) DEFAULT 'NORMAL',
  role TEXT NOT NULL CHECK (role IN ('HOST', 'ADMIN', 'USER')) DEFAULT 'USER',
  email TEXT NOT NULL DEFAULT '',
  recive_book_email TEXT NOT NULL DEFAULT '',
  nickname TEXT NOT NULL DEFAULT '',
  password_hash TEXT NOT NULL,
  avatar_url TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_user_username ON user (username);

-- user_book_link
CREATE TABLE book_user_link (
  id INTEGER NOT NULL,
  user_id INTEGER,
  book_id INTEGER,
  PRIMARY KEY (id),
  FOREIGN KEY(user_id) REFERENCES user (id),
  UNIQUE(book_id, user_id)
);

-- user_setting
CREATE TABLE user_setting (
  user_id INTEGER NOT NULL,
  key TEXT NOT NULL,
  value TEXT NOT NULL,
  UNIQUE(user_id, key)
);

-- shelf
CREATE TABLE shelf (
	id INTEGER NOT NULL,
	uuid TEXT NOT NULL,
	name TEXT NOT NULL,
	is_public SMALLINT DEFAULT 0,
	user_id INTEGER,
	created BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
	last_modified BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
	display_order INTEGER NOT NULL COLLATE NOCASE,
	order_reverse SMALLINT DEFAULT 0,
	PRIMARY KEY (id),
	FOREIGN KEY(user_id) REFERENCES user (id) ON DELETE CASCADE
);

CREATE INDEX idx_shelf_user_id ON shelf (user_id);

-- book_shelf_link
CREATE TABLE book_shelf_link (
	id INTEGER NOT NULL,
	book_id INTEGER,
	position INTEGER NOT NULL DEFAULT 1,
	shelf_id INTEGER,
	date_added BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
	PRIMARY KEY (id),
	FOREIGN KEY(shelf_id) REFERENCES shelf (id) ON DELETE CASCADE
);

CREATE INDEX idx_book_shelf_link_book_id ON book_shelf_link (book_id);
CREATE INDEX idx_book_shelf_link_shelf_id ON book_shelf_link (shelf_id);

-- bookmark
CREATE TABLE bookmark (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    book_id INTEGER NOT NULL,
    position INTEGER NOT NULL,
    tips TEXT,
    FOREIGN KEY(user_id) REFERENCES user (id) ON DELETE CASCADE,
    UNIQUE(position, book_id, user_id)
);

CREATE INDEX idx_bookmark_user_id ON bookmark (user_id);
CREATE INDEX idx_bookmark_book_id ON bookmark (book_id);

-- duration_info
CREATE TABLE duration_info (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    book_id INTEGER NOT NULL,
    start_time BIGINT NOT NULL,
    read_duration INTEGER,
    percentage INTEGER,
    FOREIGN KEY(user_id) REFERENCES user (id) ON DELETE CASCADE
);

-- reading_status
CREATE TABLE reading_status (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    book_id INTEGER NOT NULL,
    last_read_time BIGINT,
    duration INTEGER NOT NULL DEFAULT 0,
    cur_page INTEGER NOT NULL DEFAULT 0,
    percentage INTEGER NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 0,
    page INTEGER NOT NULL DEFAULT 0,
    UNIQUE (user_id, book_id),
    FOREIGN KEY(user_id) REFERENCES user (id) ON DELETE CASCADE
);
-- tag
CREATE TABLE tag (
  name TEXT NOT NULL,
  creator_id INTEGER NOT NULL,
  UNIQUE(name, creator_id)
);

-- job
CREATE TABLE job (
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL,
  `path` TEXT NOT NULL,
  type TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  FOREIGN KEY (user_id) REFERENCES user(id),
  CONSTRAINT path_length CHECK (LENGTH(path) <= 255)
);

-- book_hash_link
CREATE TABLE book_hash_link (
  book_id INTEGER NOT NULL,
  hash TEXT NOT NULL,
  PRIMARY KEY (book_id, hash)
);
