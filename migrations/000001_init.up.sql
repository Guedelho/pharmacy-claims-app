CREATE TABLE IF NOT EXISTS pharmacies (
    id SERIAL PRIMARY KEY,
    npi VARCHAR(10) UNIQUE NOT NULL,
    chain VARCHAR(20) NOT NULL,
    CONSTRAINT valid_chain CHECK (chain IN ('health', 'saint', 'doctor'))
);

CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY,
    ndc VARCHAR(11) NOT NULL,
    npi VARCHAR(10) NOT NULL,
    quantity DECIMAL(10,2) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    timestamp TIMESTAMP NOT NULL,

    FOREIGN KEY (npi) REFERENCES pharmacies(npi)
);

CREATE TABLE IF NOT EXISTS reversals (
    id UUID PRIMARY KEY,
    claim_id UUID NOT NULL,
    timestamp TIMESTAMP NOT NULL,

    FOREIGN KEY (claim_id) REFERENCES claims(id)
);

CREATE INDEX IF NOT EXISTS idx_claims_npi ON claims(npi);
CREATE INDEX IF NOT EXISTS idx_claims_ndc ON claims(ndc);
CREATE INDEX IF NOT EXISTS idx_claims_timestamp ON claims(timestamp);
CREATE INDEX IF NOT EXISTS idx_reversals_claim_id ON reversals(claim_id);
CREATE INDEX IF NOT EXISTS idx_reversals_timestamp ON reversals(timestamp);
