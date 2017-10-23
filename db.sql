CREATE TABLE `user_relation` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `user1` VARCHAR(64) NOT NULL,
    `user2` VARCHAR(64) NOT NULL,
    `encounters` INTEGER
)
