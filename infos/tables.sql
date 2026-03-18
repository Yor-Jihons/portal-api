-- 1. カテゴリテーブル
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_name TEXT NOT NULL UNIQUE
);

-- 2. 学習履歴テーブル
CREATE TABLE study_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL,
    content TEXT,
    date DATE NOT NULL DEFAULT (CURRENT_DATE),
    time INTEGER NOT NULL DEFAULT 1
);

-- 3. 中間テーブル (多対多の紐付け)
CREATE TABLE study_log_categories (
    study_log_id INTEGER REFERENCES study_logs(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (study_log_id, category_id)
);
