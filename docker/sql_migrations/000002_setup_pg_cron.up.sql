-- Create basic functions for job management (no pg_cron dependency)

-- Function to run scan jobs
CREATE OR REPLACE FUNCTION run_scan_job(job_id UUID) RETURNS VOID AS $$
BEGIN
    -- Update job status to running
    UPDATE jobs SET status = 'running', last_run = CURRENT_TIMESTAMP WHERE id = job_id;
    
    -- Insert a new job result
    INSERT INTO job_results (id, job_id, start_time, end_time, status, details)
    VALUES (gen_random_uuid(), job_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'complete', 'Scan completed');
    
    -- Update job status to complete
    UPDATE jobs SET status = 'complete', updated_at = CURRENT_TIMESTAMP WHERE id = job_id;
END;
$$ LANGUAGE plpgsql;

-- Function to run report jobs
CREATE OR REPLACE FUNCTION run_report_job(job_id UUID) RETURNS VOID AS $$
BEGIN
    -- Update job status to running
    UPDATE jobs SET status = 'running', last_run = CURRENT_TIMESTAMP WHERE id = job_id;
    
    -- Insert a new job result
    INSERT INTO job_results (id, job_id, start_time, end_time, status, details)
    VALUES (gen_random_uuid(), job_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'complete', 'Report generated');
    
    -- Update job status to complete
    UPDATE jobs SET status = 'complete', updated_at = CURRENT_TIMESTAMP WHERE id = job_id;
END;
$$ LANGUAGE plpgsql;