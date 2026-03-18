CREATE TABLE IF NOT EXISTS "public"."oauth_app"
(
    "app_uuid"     uuid PRIMARY KEY,
    "app_name"     varchar(80)                     NOT NULL,
    "payload"      jsonb                           DEFAULT NULL,
    "status"       integer                         NOT NULL DEFAULT 0,
    "created_at"   timestamp with time zone        DEFAULT CURRENT_TIMESTAMP,
    "updated_at"   timestamp with time zone        DEFAULT CURRENT_TIMESTAMP
);
