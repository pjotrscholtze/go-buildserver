CREATE TABLE IF NOT EXISTS buildresultline(
	jobid INT	      NOT NULL,
	line  TEXT        NOT NULL,
	pipe  VARCHAR(16) NOT NULL,
	time  DATETIME    NOT NULL,
    PRIMARY KEY (JobID, Time)
);
