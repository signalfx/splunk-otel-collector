CREATE USER 'otelu'@'localhost' IDENTIFIED BY 'otelp';
GRANT ALL PRIVILEGES ON *.* TO 'otelu'@'localhost' WITH GRANT OPTION;

CREATE DATABASE otel;

USE otel;

CREATE TABLE dummytable (myfield VARCHAR(20));

INSERT INTO mytable VALUES ('bar'), ('foo');
