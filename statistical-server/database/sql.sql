CREATE TABLE IF NOT EXISTS register (
    xid String,
    user_id String,
    role_id String,
    device_id String,
    server_id Int64,
    ip String,
    rule_version String,
    created_at  DateTime
) engine=MergeTree() ORDER BY xid;


CREATE TABLE IF NOT EXISTS login (
    xid String,
    user_id String,
    role_id String,
    device_id String,
    server_id Int64,
    ip String,
    rule_version String,
    created_at  DateTime
) engine=MergeTree() ORDER BY xid;