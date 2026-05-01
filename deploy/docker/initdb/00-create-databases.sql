-- Create one database per service. Each service applies its own schema
-- migrations from db/schema/ on startup (or via Atlas in production).

CREATE DATABASE eo_catalog;
CREATE DATABASE eo_pipeline;
CREATE DATABASE eo_analytics;

CREATE DATABASE sat_mission;
CREATE DATABASE sat_fsw;
CREATE DATABASE sat_telemetry;
CREATE DATABASE sat_command;
CREATE DATABASE sat_simulation;

CREATE DATABASE gs_mc;
CREATE DATABASE gs_rf;
CREATE DATABASE gs_scheduler;
CREATE DATABASE gs_ingest;

CREATE DATABASE gi_fusion;
CREATE DATABASE gi_analytics;
CREATE DATABASE gi_tiles;
CREATE DATABASE gi_reports;
CREATE DATABASE gi_predict;
