-- Sample test data for PostgreSQL tests

-- Insert sample services
INSERT INTO services (id, name, type, config, created_at, updated_at)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'test_postgres', 'postgresql', '{"host": "localhost", "port": 5432, "user": "test_user", "database": "test_db"}', NOW(), NOW()),
    ('22222222-2222-2222-2222-222222222222', 'test_mysql', 'mysql', '{"host": "localhost", "port": 3306, "user": "test_user", "database": "test_db"}', NOW(), NOW()),
    ('33333333-3333-3333-3333-333333333333', 'test_sqlite', 'sqlite', '{"path": "/tmp/test.db"}', NOW(), NOW());

-- Insert sample jobs
INSERT INTO jobs (id, service_id, type, status, schedule, created_at, updated_at, last_run)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', 'scan', 'pending', '0 0 * * *', NOW(), NOW(), NULL),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222222', 'scan', 'pending', '0 12 * * *', NOW(), NOW(), NULL),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', '33333333-3333-3333-3333-333333333333', 'report', 'complete', '0 0 * * 0', NOW(), NOW(), NOW() - INTERVAL '1 day');

-- Insert sample job results
INSERT INTO job_results (id, job_id, start_time, end_time, status, details)
VALUES
    ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'cccccccc-cccc-cccc-cccc-cccccccccccc', NOW() - INTERVAL '2 day', NOW() - INTERVAL '1 day', 'complete', 'Report generated successfully');

-- Create a test schema with test tables for schema inspection testing
CREATE SCHEMA test_schema;

-- Create test tables in test_schema
CREATE TABLE test_schema.users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE test_schema.posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES test_schema.users(id),
    title VARCHAR(100) NOT NULL,
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample data into test tables
INSERT INTO test_schema.users (username, email)
VALUES
    ('user1', 'user1@example.com'),
    ('user2', 'user2@example.com'),
    ('user3', 'user3@example.com');

INSERT INTO test_schema.posts (user_id, title, content)
VALUES
    (1, 'First Post', 'This is the first post content'),
    (1, 'Second Post', 'This is the second post content'),
    (2, 'Hello World', 'Hello world post content'),
    (3, 'Introduction', 'Introducing myself to the community');

-- Create a table with diverse data types to test type handling
CREATE TABLE test_schema.data_types (
    id SERIAL PRIMARY KEY,
    int_val INTEGER,
    bigint_val BIGINT,
    text_val TEXT,
    bool_val BOOLEAN,
    date_val DATE,
    timestamp_val TIMESTAMP,
    timestamptz_val TIMESTAMPTZ,
    numeric_val NUMERIC(10,2),
    json_val JSON,
    jsonb_val JSONB,
    uuid_val UUID,
    array_val TEXT[]
);

-- Insert sample values for different data types
INSERT INTO test_schema.data_types (
    int_val, bigint_val, text_val, bool_val, date_val, timestamp_val,
    timestamptz_val, numeric_val, json_val, jsonb_val, uuid_val, array_val
)
VALUES (
    42,
    9223372036854775807,
    'Sample text value',
    TRUE,
    '2025-04-12',
    '2025-04-12 12:00:00',
    '2025-04-12 12:00:00+00',
    123.45,
    '{"key": "value", "nested": {"inner": "value"}}',
    '{"indexed": true, "data": [1, 2, 3]}',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    ARRAY['one', 'two', 'three']
);