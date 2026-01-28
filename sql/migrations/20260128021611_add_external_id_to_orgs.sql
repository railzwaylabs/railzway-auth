ALTER TABLE orgs ADD COLUMN external_id VARCHAR(255);
CREATE UNIQUE INDEX idx_orgs_external_id ON orgs(external_id);
