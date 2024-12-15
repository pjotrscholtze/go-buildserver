CREATE TABLE IF NOT EXISTS buildresultline(
	id           INTEGER      PRIMARY KEY AUTOINCREMENT NOT NULL,
	jobid INT	      NOT NULL,
	line  TEXT        NOT NULL,
	pipe  VARCHAR(16) NOT NULL,
	time  DATETIME    NOT NULL
);
