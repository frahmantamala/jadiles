-- =====================================================
-- Migration: 002_add_version_columns.sql
-- Description: Add version column for optimistic locking
-- =====================================================
-- +goose Up

-- Add version column to users table
ALTER TABLE users ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;

-- Add version column to parent_profiles table
ALTER TABLE parent_profiles ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;

-- Add version column to vendors table
ALTER TABLE vendors ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;

-- Add version column to children table
ALTER TABLE children ADD COLUMN version INTEGER DEFAULT 1 NOT NULL;

-- Create index on version columns for better query performance
CREATE INDEX idx_users_version ON users(id, version);
CREATE INDEX idx_parent_profiles_version ON parent_profiles(id, version);
CREATE INDEX idx_vendors_version ON vendors(id, version);
CREATE INDEX idx_children_version ON children(id, version);

-- +goose Down

-- Remove indexes
DROP INDEX IF EXISTS idx_users_version;
DROP INDEX IF EXISTS idx_parent_profiles_version;
DROP INDEX IF EXISTS idx_vendors_version;
DROP INDEX IF EXISTS idx_children_version;

-- Remove version columns
ALTER TABLE users DROP COLUMN IF EXISTS version;
ALTER TABLE parent_profiles DROP COLUMN IF EXISTS version;
ALTER TABLE vendors DROP COLUMN IF EXISTS version;
ALTER TABLE children DROP COLUMN IF EXISTS version;
