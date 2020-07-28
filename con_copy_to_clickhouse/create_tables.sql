-- create syslogshold table
create table MON_DB.syslogshold
(
DB LowCardinality(String) CODEC(ZSTD(1)),
dbid UInt16 CODEC(Gorilla, ZSTD(1)),
reserved UInt16 CODEC(Gorilla, ZSTD(1)),
spid UInt16 CODEC(Gorilla, ZSTD(1)),
page UInt32 CODEC(Gorilla, ZSTD(1)),
xactid LowCardinality(String) CODEC(ZSTD(1)),
masterxactid LowCardinality(String) CODEC(ZSTD(1)),
starttime DateTime CODEC(Delta(4), ZSTD(1)),
name LowCardinality(String) CODEC(ZSTD(1)),
xloid UInt16 CODEC(Gorilla, ZSTD(1)),
log_datetime DateTime CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{layer}-{shard}/MON_DB.syslogshold', '{replica}')
PARTITION BY toYYYYMM(log_datetime)
ORDER BY (DB, dbid, name, spid, starttime)
SETTINGS index_granularity = 8192;


create table MON_DB.syslogshold
(
DB LowCardinality(String) CODEC(ZSTD(1)),
dbid UInt16 CODEC(Gorilla, ZSTD(1)),
reserved UInt16 CODEC(Gorilla, ZSTD(1)),
spid UInt16 CODEC(Gorilla, ZSTD(1)),
page UInt32 CODEC(Gorilla, ZSTD(1)),
xactid LowCardinality(String) CODEC(ZSTD(1)),
masterxactid LowCardinality(String) CODEC(ZSTD(1)),
starttime DateTime CODEC(Delta(4), ZSTD(1)),
name LowCardinality(String) CODEC(ZSTD(1)),
xloid UInt16 CODEC(Gorilla, ZSTD(1)),
log_datetime DateTime CODEC(Delta(4), ZSTD(1))
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(log_datetime)
ORDER BY (DB, dbid, name, spid, starttime)
SETTINGS index_granularity = 8192;


-- create sysprocesses table
create table MON_DB.sysprocesses
(
server LowCardinality(String) CODEC(ZSTD(1)),
loginame LowCardinality(String) CODEC(ZSTD(1)),
DB LowCardinality(String) CODEC(ZSTD(1)),
spid UInt16 CODEC(Gorilla, ZSTD(1)),
loggedindatetime DateTime CODEC(Delta(4), ZSTD(1)),
hostname LowCardinality(String) CODEC(ZSTD(1)),
ipaddr IPv4 CODEC(T64, ZSTD(1)),
hostprocess String,
log_datetime DateTime CODEC(Delta(4), ZSTD(1))
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{layer}-{shard}/MON_DB.sysprocesses', '{replica}')
PARTITION BY toYYYYMM(log_datetime)
ORDER BY (server, DB, loginame, spid, loggedindatetime)
SETTINGS index_granularity = 8192;

create table MON_DB.sysprocesses
(
server LowCardinality(String) CODEC(ZSTD(1)),
loginame LowCardinality(String) CODEC(ZSTD(1)),
DB LowCardinality(String) CODEC(ZSTD(1)),
spid UInt16 CODEC(Gorilla, ZSTD(1)),
loggedindatetime DateTime CODEC(Delta(4), ZSTD(1)),
hostname LowCardinality(String) CODEC(ZSTD(1)),
ipaddr IPv4 CODEC(T64, ZSTD(1)),
hostprocess String,
log_datetime DateTime CODEC(Delta(4), ZSTD(1))
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(log_datetime)
ORDER BY (server, DB, loginame, spid, loggedindatetime)
SETTINGS index_granularity = 8192;

