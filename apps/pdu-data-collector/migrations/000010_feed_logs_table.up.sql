INSERT INTO `logs` (`type`, `key`, `value`)
VALUES
    -- ConnectInfluxDB
    ('ConnectInfluxDB', 'Name', 'ConnectInfluxDB'),
    ('ConnectInfluxDB', 'Code', 'INFLUXDB01'),
    ('ConnectInfluxDB', 'Category', 'InfluxDB'),
    ('ConnectInfluxDB', 'Description', 'Connects to the InfluxDB database'),
    ('ConnectInfluxDB', 'Level', 'info'),

    -- OutputInfluxDB
    ('OutputInfluxDB', 'Name', 'OutputInfluxDB'),
    ('OutputInfluxDB', 'Code', 'INFLUXDB02'),
    ('OutputInfluxDB', 'Category', 'InfluxDB'),
    ('OutputInfluxDB', 'Description', 'Outputs data to the InfluxDB database'),
    ('OutputInfluxDB', 'Level', 'info'),

    -- AggregateInfluxDB
    ('AggregateInfluxDB', 'Name', 'AggregateInfluxDB'),
    ('AggregateInfluxDB', 'Code', 'INFLUXDB03'),
    ('AggregateInfluxDB', 'Category', 'InfluxDB'),
    ('AggregateInfluxDB', 'Description', 'Performs aggregation on InfluxDB data'),
    ('AggregateInfluxDB', 'Level', 'info'),

    -- BackupInfluxDB
    ('BackupInfluxDB', 'Name', 'BackupInfluxDB'),
    ('BackupInfluxDB', 'Code', 'INFLUXDB04'),
    ('BackupInfluxDB', 'Category', 'InfluxDB'),
    ('BackupInfluxDB', 'Description', 'Backs up data from the InfluxDB database'),
    ('BackupInfluxDB', 'Level', 'info'),

    -- ConnectSqlServer
    ('ConnectSqlServer', 'Name', 'ConnectSqlServer'),
    ('ConnectSqlServer', 'Code', 'SQLSERVER01'),
    ('ConnectSqlServer', 'Category', 'SQLServer'),
    ('ConnectSqlServer', 'Description', 'Connects to the SQL Server database'),
    ('ConnectSqlServer', 'Level', 'info'),

    -- OutputSqlServer
    ('OutputSqlServer', 'Name', 'OutputSqlServer'),
    ('OutputSqlServer', 'Code', 'SQLSERVER02'),
    ('OutputSqlServer', 'Category', 'SQLServer'),
    ('OutputSqlServer', 'Description', 'Outputs data to the SQL Server database'),
    ('OutputSqlServer', 'Level', 'info'),

    -- ConnectMysql
    ('ConnectMysql', 'Name', 'ConnectMysql'),
    ('ConnectMysql', 'Code', 'MYSQL01'),
    ('ConnectMysql', 'Category', 'Mysql'),
    ('ConnectMysql', 'Description', 'Connects to the MySQL database'),
    ('ConnectMysql', 'Level', 'info'),

    -- LoadMigration
    ('LoadMigration', 'Name', 'LoadMigration'),
    ('LoadMigration', 'Code', 'MIG01'),
    ('LoadMigration', 'Category', 'Migration'),
    ('LoadMigration', 'Description', 'Loads migration files for database schema updates'),
    ('LoadMigration', 'Level', 'info'),

    -- LoadCrontab
    ('LoadCrontab', 'Name', 'LoadCrontab'),
    ('LoadCrontab', 'Code', 'CRON01'),
    ('LoadCrontab', 'Category', 'Cron'),
    ('LoadCrontab', 'Description', 'Loads and schedules crontab jobs'),
    ('LoadCrontab', 'Level', 'info'),

    -- LoadEnvConfig
    ('LoadEnvConfig', 'Name', 'LoadEnvConfig'),
    ('LoadEnvConfig', 'Code', 'ENV01'),
    ('LoadEnvConfig', 'Category', 'Config'),
    ('LoadEnvConfig', 'Description', 'Loads environment configuration settings'),
    ('LoadEnvConfig', 'Level', 'info'),

    -- MatchTag
    ('MatchTag', 'Name', 'MatchTag'),
    ('MatchTag', 'Code', 'TAG01'),
    ('MatchTag', 'Category', 'Tag'),
    ('MatchTag', 'Description', 'Matches and applies tags to incoming data'),
    ('MatchTag', 'Level', 'info');