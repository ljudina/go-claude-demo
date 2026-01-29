CREATE TABLE nav_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    icon TEXT,
    parent_id INTEGER,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES nav_items(id) ON DELETE CASCADE
);

CREATE TABLE nav_item_roles (
    nav_item_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (nav_item_id, role_id),
    FOREIGN KEY (nav_item_id) REFERENCES nav_items(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);
