CREATE TABLE click_events (
    id BIGSERIAL PRIMARY KEY,
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ip_hash TEXT NOT NULL,
    country TEXT, region TEXT, city TEXT, timezone TEXT,
    browser TEXT, browser_version TEXT, os TEXT, device TEXT,
    is_bot BOOLEAN NOT NULL DEFAULT FALSE,
    referer TEXT, language TEXT, user_agent TEXT
);
CREATE INDEX idx_clicks_link_id_time ON click_events(link_id, clicked_at DESC);
CREATE INDEX idx_clicks_link_country ON click_events(link_id, country);

---- create above / drop below ----

DROP TABLE click_events;
