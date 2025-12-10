-- Merchants
CREATE TABLE IF NOT EXISTS merchants (
    npc_id UUID PRIMARY KEY,
    shop_name VARCHAR(100),
    specialization VARCHAR(50),
    wealth INT NOT NULL DEFAULT 0,
    price_modifier FLOAT NOT NULL DEFAULT 1.0,
    reputation FLOAT NOT NULL DEFAULT 50.0,
    data JSONB, -- Stores sales history, business hours, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Market Data
CREATE TABLE IF NOT EXISTS market_data (
    location_id UUID NOT NULL,
    item_id UUID NOT NULL,
    local_supply INT NOT NULL DEFAULT 0,
    local_demand INT NOT NULL DEFAULT 0,
    average_price FLOAT NOT NULL DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (location_id, item_id)
);

CREATE TABLE IF NOT EXISTS price_history (
    location_id UUID NOT NULL,
    item_id UUID NOT NULL,
    price FLOAT NOT NULL,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_price_history_lookup ON price_history(location_id, item_id, recorded_at);

-- Trade Routes
CREATE TABLE IF NOT EXISTS trade_routes (
    route_id UUID PRIMARY KEY,
    merchant_id UUID REFERENCES merchants(npc_id),
    origin_id UUID NOT NULL,
    destination_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL,
    data JSONB, -- Stores cargo, profit estimates
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_trade_routes_merchant ON trade_routes(merchant_id);

-- Barter Offers
CREATE TABLE IF NOT EXISTS barter_offers (
    offer_id UUID PRIMARY KEY,
    offered_by UUID NOT NULL,
    offered_to UUID NOT NULL,
    status VARCHAR(20) NOT NULL,
    data JSONB, -- Stores items offered/requested
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_barter_offers_participants ON barter_offers(offered_by, offered_to);
