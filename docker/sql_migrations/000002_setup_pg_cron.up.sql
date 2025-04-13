CREATE EXTENSION IF NOT EXISTS pg_cron;

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

-- Function to schedule jobs using pg_cron
CREATE OR REPLACE FUNCTION schedule_job(job_id UUID, job_schedule TEXT) RETURNS VOID AS $$
DECLARE
    job_type TEXT;
BEGIN
    SELECT type INTO job_type FROM jobs WHERE id = job_id;
    
    IF job_type = 'scan' THEN
        PERFORM cron.schedule(job_id::TEXT, job_schedule, $$SELECT run_scan_job('$$ || job_id || $$')$$);
    ELSIF job_type = 'report' THEN
        PERFORM cron.schedule(job_id::TEXT, job_schedule, $$SELECT run_report_job('$$ || job_id || $$')$$);
    ELSE
        RAISE EXCEPTION 'Unknown job type: %', job_type;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically schedule jobs when inserted
CREATE OR REPLACE FUNCTION auto_schedule_job() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.schedule IS NOT NULL THEN
        PERFORM schedule_job(NEW.id, NEW.schedule);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER schedule_job_trigger
AFTER INSERT ON jobs
FOR EACH ROW
EXECUTE FUNCTION auto_schedule_job();