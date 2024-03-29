CREATE TABLE test_case_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  host TEXT DEFAULT '' ,
  method TEXT NOT NULL DEFAULT '',
  `path` TEXT NOT NULL DEFAULT '' ,
  request_id TEXT NOT NULL DEFAULT '',
  request_header TEXT NOT NULL DEFAULT '',
  request_query TEXT NOT NULL DEFAULT '',
  request_body TEXT NOT NULL DEFAULT '',
  response_header TEXT NOT NULL DEFAULT '',
  response_body TEXT NOT NULL DEFAULT '',
  test_case_id TEXT NOT NULL DEFAULT '',
  `description` TEXT NOT NULL DEFAULT '',
  test_case_input TEXT NOT NULL DEFAULT '',
  test_case_output TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);