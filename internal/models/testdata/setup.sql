
CREATE TABLE snippets
(
    id serial primary key,
    title character varying(100) NOT NULL,
    content text NOT NULL,
    created timestamp with time zone NOT NULL,
    expires timestamp with time zone NOT NULL
);

CREATE INDEX idx_snippets_created ON snippets (created);

CREATE TABLE users
(
    id serial primary key,
    name character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    hashed_password character(60) NOT NULL,
    created timestamp with time zone,
    CONSTRAINT users_uc_email UNIQUE (email)
);

INSERT INTO users (name, email, hashed_password, created) VALUES (
    'Alice Jones',
    'alice@example.com',
    '$2a$12$NuTjWXm3KKntReFwyBVHyuf/to.HEwTy.eS206TNfkGfr6HzGJSWG',
    '2022-01-01 10:00:00'
);

