-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'reserved',
    type VARCHAR(20) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    sha256 VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_status CHECK (status IN ('reserved', 'finalized', 'failed')),
    CONSTRAINT chk_type CHECK (type IN ('image', 'video')),
    CONSTRAINT unq_filename_sha256 UNIQUE (filename, sha256)
);

CREATE INDEX idx_media_status ON media(status);
CREATE INDEX idx_media_type ON media(type);
CREATE INDEX idx_media_created_at ON media(created_at);

-- Trigger to auto-update updated_at
CREATE TRIGGER update_media_updated_at
    BEFORE UPDATE ON media
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Junction table for many-to-many relationship between media and tags
CREATE TABLE IF NOT EXISTS media_tags (
    media_id UUID NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (media_id, tag_id)
);

CREATE INDEX idx_media_tags_media_id ON media_tags(media_id);
CREATE INDEX idx_media_tags_tag_id ON media_tags(tag_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS media_tags;
DROP TABLE IF EXISTS media;
-- +goose StatementEnd
