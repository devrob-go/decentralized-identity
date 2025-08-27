-- Database initialization for Decentralized Identity System
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Auth Service tables (existing)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR NOT NULL,
    email VARCHAR UNIQUE NOT NULL,
    password VARCHAR NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- DID Manager tables
CREATE TABLE IF NOT EXISTS dids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    did VARCHAR(255) NOT NULL UNIQUE,
    user_hash VARCHAR(64) NOT NULL UNIQUE,
    public_key TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    blockchain_tx VARCHAR(66),
    -- Ethereum transaction hash
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create blockchain_jobs table
CREATE TABLE IF NOT EXISTS blockchain_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_type VARCHAR(50) NOT NULL,
    did_id UUID NOT NULL REFERENCES dids(id) ON DELETE CASCADE,
    user_hash VARCHAR(64) NOT NULL,
    did VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_dids_user_id ON dids(user_id);

CREATE INDEX IF NOT EXISTS idx_dids_status ON dids(status);

CREATE INDEX IF NOT EXISTS idx_dids_user_hash ON dids(user_hash);

CREATE INDEX IF NOT EXISTS idx_blockchain_jobs_status ON blockchain_jobs(status);

CREATE INDEX IF NOT EXISTS idx_blockchain_jobs_did_id ON blockchain_jobs(did_id);

CREATE INDEX IF NOT EXISTS idx_blockchain_jobs_created_at ON blockchain_jobs(created_at);

-- Create status check constraints
ALTER TABLE
    dids
ADD
    CONSTRAINT chk_dids_status CHECK (
        status IN (
            'pending',
            'active',
            'revoked',
            'expired',
            'failed'
        )
    );

ALTER TABLE
    blockchain_jobs
ADD
    CONSTRAINT chk_blockchain_jobs_status CHECK (
        status IN (
            'pending',
            'processing',
            'completed',
            'failed',
            'retrying'
        )
    );

-- Create function to update updated_at timestamp
CREATE
OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $ $ BEGIN NEW.updated_at = NOW();

RETURN NEW;

END;

$ $ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_dids_updated_at BEFORE
UPDATE
    ON dids FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_blockchain_jobs_updated_at BEFORE
UPDATE
    ON blockchain_jobs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create view for DID status overview
CREATE
OR REPLACE VIEW did_status_overview AS
SELECT
    d.id,
    d.user_id,
    d.did,
    d.status as did_status,
    d.blockchain_tx,
    d.created_at as did_created_at,
    COUNT(bj.id) as pending_jobs,
    MAX(bj.updated_at) as last_job_update
FROM
    dids d
    LEFT JOIN blockchain_jobs bj ON d.id = bj.did_id
    AND bj.status IN ('pending', 'processing')
GROUP BY
    d.id,
    d.user_id,
    d.did,
    d.status,
    d.blockchain_tx,
    d.created_at;