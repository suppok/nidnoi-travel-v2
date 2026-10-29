-- Staff users of the backoffice. "user" is reserved in Postgres, hence app_user.
CREATE TABLE app_user (
    id                   UUID        PRIMARY KEY,
    email                TEXT        NOT NULL,
    name                 TEXT        NOT NULL,
    password_hash        TEXT        NOT NULL,
    -- Fixed role set; every user has at least one role. Keep in sync with pkg/constants/role.
    roles                TEXT[]      NOT NULL,
    must_change_password BOOLEAN     NOT NULL DEFAULT TRUE,
    last_login_at        TIMESTAMPTZ NULL,
    status               BOOLEAN     NOT NULL DEFAULT TRUE,
    create_date          TIMESTAMPTZ NOT NULL,
    update_date          TIMESTAMPTZ NOT NULL,
    create_by            UUID        NOT NULL,
    update_by            UUID        NOT NULL,
    CONSTRAINT app_user_email_lower CHECK (email = lower(email)),
    CONSTRAINT app_user_roles_not_empty CHECK (cardinality(roles) >= 1),
    CONSTRAINT app_user_roles_known CHECK (roles <@ ARRAY['super_admin', 'admin', 'sales', 'finance']::TEXT[])
);

CREATE UNIQUE INDEX app_user_email_key ON app_user (email);

-- At most one super admin (the bootstrap user).
CREATE UNIQUE INDEX app_user_one_super_admin ON app_user ((TRUE)) WHERE 'super_admin' = ANY (roles);
