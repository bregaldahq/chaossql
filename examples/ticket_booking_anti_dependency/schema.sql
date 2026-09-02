CREATE TABLE seats (
    id INT PRIMARY KEY,
    row_num INT NOT NULL,
    seat_num INT NOT NULL,
    reserved_by TEXT DEFAULT NULL
);

CREATE TABLE booking_audit (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    seat_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
