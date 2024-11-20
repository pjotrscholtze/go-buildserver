CREATE TABLE IF NOT EXISTS Job(
	id           INTEGER      PRIMARY KEY AUTOINCREMENT NOT NULL,
	buildreason  VARCHAR(256)                           NULL,
	origin       VARCHAR(64)                            NULL,
	queuetime    DATETIME                               NULL,
	pipelinename VARCHAR(64)                            NULL,
	status       VARCHAR(16)                            NULL,
	starttime    DATETIME                               NULL
);

CREATE TABLE IF NOT EXISTS BuildResultLine(
	JobID INT(11)     NOT NULL,
	Line  TEXT        NOT NULL,
	Pipe  VARCHAR(16) NOT NULL,
	Time  DATETIME    NOT NULL,
    PRIMARY KEY (JobID, Time)
);
