-- import Table
CREATE TABLE import (
    id UUID PRIMARY KEY NOT NULL,
    resource VARCHAR(64) NOT NULL,
    file_name VARCHAR(120) NOT NULL,
    created_at timestamp NOT NULL
);
