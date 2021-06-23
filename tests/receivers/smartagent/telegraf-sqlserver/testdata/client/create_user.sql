USE master;
GO
CREATE LOGIN [signalfxagent] WITH PASSWORD = N'Password!';
GO
GRANT VIEW SERVER STATE TO [signalfxagent];
GO
GRANT VIEW ANY DEFINITION TO [signalfxagent];
GO
SELECT login.name FROM master.sys.sql_logins as login WHERE login.name = 'signalfxagent';
