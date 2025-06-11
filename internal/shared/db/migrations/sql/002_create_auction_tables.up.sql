-- Crear table for auction lots
CREATE TABLE IF NOT EXISTS auction_lots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    initial_price DECIMAL(18, 2) NOT NULL,
    current_price DECIMAL(18, 2) NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    state VARCHAR(50) NOT NULL, -- e.g., 'pending', 'active', 'finished', 'cancelled'
    last_bid_time TIMESTAMP WITH TIME ZONE, -- Nullable, for time extension logic
    time_extension INTERVAL NOT NULL, -- time extension
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Crear un índice en el estado para consultas rápidas de lotes activos/pendientes
-- Index for querys on active/pending state lots
CREATE INDEX idx_auction_lots_state ON auction_lots (state);

-- table for bids
CREATE TABLE IF NOT EXISTS bids (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lot_id UUID NOT NULL,
    user_id UUID NOT NULL, -- user id that make the bid
    amount DECIMAL(18, 2) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- foreign key for auction_lots table
    CONSTRAINT fk_bids_lot_id
        FOREIGN KEY (lot_id)
        REFERENCES auction_lots (id)
        ON DELETE CASCADE, -- if a lot is deleted, their bids are deleted too

-- user table foreign key
    CONSTRAINT fk_bids_user_id
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE -- id a user is deleted, their bids are deleted too
);

-- Indexes for better query performance for querys on bids table
CREATE INDEX idx_bids_lot_id ON bids (lot_id);
CREATE INDEX idx_bids_user_id ON bids (user_id);
CREATE INDEX idx_bids_lot_id_timestamp ON bids (lot_id, timestamp DESC); -- useful to retrieve most recent bids from a lot

-- Trigger to update 'updated_at' in auction_lots
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_auction_lots_updated_at
BEFORE UPDATE ON auction_lots
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();