CREATE TABLE IF NOT EXISTS job(
	id           INT      		NOT NULL PRIMARY KEY AUTO_INCREMENT ,
	buildreason  VARCHAR(256)   NULL,
	origin       VARCHAR(64)    NULL,
	queuetime    DATETIME       NULL,
	pipelinename VARCHAR(64)    NULL,
	status       VARCHAR(16)    NULL,
	starttime    DATETIME       NULL
);
