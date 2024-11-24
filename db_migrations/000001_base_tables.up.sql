CREATE TABLE IF NOT EXISTS job(
	id           INTEGER      PRIMARY KEY AUTOINCREMENT NOT NULL,
	buildreason  VARCHAR(256)                           NULL,
	origin       VARCHAR(64)                            NULL,
	queuetime    DATETIME                               NULL,
	pipelinename VARCHAR(64)                            NULL,
	status       VARCHAR(16)                            NULL,
	starttime    DATETIME                               NULL
);

CREATE TABLE IF NOT EXISTS buildresultline(
	jobid INT(11)     NOT NULL,
	line  TEXT        NOT NULL,
	pipe  VARCHAR(16) NOT NULL,
	time  DATETIME    NOT NULL,
    PRIMARY KEY (JobID, Time)
);
