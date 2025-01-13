-- +goose Up
-- +goose StatementBegin
CREATE TABLE songs (
    "group" TEXT NOT NULL,
    song TEXT NOT NULL,
    "text" TEXT,
    release_date TEXT,
    link TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE songs;
-- +goose StatementEnd
