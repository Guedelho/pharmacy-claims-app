DROP INDEX IF EXISTS idx_reversals_timestamp;
DROP INDEX IF EXISTS idx_reversals_claim_id;
DROP INDEX IF EXISTS idx_claims_timestamp;
DROP INDEX IF EXISTS idx_claims_ndc;
DROP INDEX IF EXISTS idx_claims_npi;

DROP TABLE IF EXISTS reversals;
DROP TABLE IF EXISTS claims;
DROP TABLE IF EXISTS pharmacies;
