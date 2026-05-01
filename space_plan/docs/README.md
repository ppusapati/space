# Ground Station Module

> Satellite Tracking, Pass Scheduling, Telemetry & Command Operations

## Overview

The Ground Station module provides comprehensive capabilities for satellite operations including TLE management, pass prediction and scheduling, real-time telemetry processing, command and control, anomaly detection, and alerting.

## Services

| Service | RPCs | Description |
|---------|------|-------------|
| [SatelliteService](./api/satellite-service.md) | 8 | Satellite catalog, TLE management |
| [GroundStationService](./api/groundstation-service.md) | 7 | Ground station configuration |
| [PassService](./api/pass-service.md) | 9 | Pass prediction and scheduling |
| [TelemetryService](./api/telemetry-service.md) | 6 | Real-time and historical telemetry |
| [CommandService](./api/command-service.md) | 8 | Command & control with 2-person auth |
| [AnomalyService](./api/anomaly-service.md) | 6 | Anomaly detection and tracking |
| [AlertService](./api/alert-service.md) | 8 | Alert management and routing |
| **Total** | **52** | |

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       GROUND STATION MODULE                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                    REAL-TIME OPERATIONS LAYER                         │ │
│  │                                                                       │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │ │
│  │  │  Telemetry  │  │   Command   │  │   Alert     │  │  Pass       │  │ │
│  │  │  Streaming  │  │   Queue     │  │  Manager    │  │  Executor   │  │ │
│  │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  │ │
│  │         │                │                │                │         │ │
│  │         └────────────────┴────────────────┴────────────────┘         │ │
│  │                                  │                                    │ │
│  │                    ┌─────────────┴─────────────┐                     │ │
│  │                    │     NATS JetStream        │                     │ │
│  │                    │    (Event Bus)            │                     │ │
│  │                    └───────────────────────────┘                     │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                    APPLICATION SERVICES LAYER                         │ │
│  │                                                                       │ │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐        │ │
│  │  │ Satellite  │ │   Pass     │ │ Telemetry  │ │  Command   │        │ │
│  │  │  Service   │ │  Service   │ │  Service   │ │  Service   │        │ │
│  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘        │ │
│  │                                                                       │ │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐                       │ │
│  │  │Ground Stn  │ │  Anomaly   │ │   Alert    │                       │ │
│  │  │  Service   │ │  Service   │ │  Service   │                       │ │
│  │  └────────────┘ └────────────┘ └────────────┘                       │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                    DATA & INTEGRATION LAYER                           │ │
│  │                                                                       │ │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐        │ │
│  │  │ PostgreSQL │ │TimescaleDB │ │Space-Track │ │    SDR     │        │ │
│  │  │   (Core)   │ │(Telemetry) │ │   Client   │ │ Controller │        │ │
│  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘        │ │
│  │                                                                       │ │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐                       │ │
│  │  │  AWS GS    │ │  Antenna   │ │    SGP4    │                       │ │
│  │  │   Client   │ │ Controller │ │ Propagator │                       │ │
│  │  └────────────┘ └────────────┘ └────────────┘                       │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Data Model

See [Data Model](./data-model.md) for complete schema.

### Key Tables

| Table | Description |
|-------|-------------|
| `satellites` | Satellite catalog with frequencies |
| `tle_history` | Two-Line Element set history |
| `ground_stations` | Ground station locations and capabilities |
| `pass_schedules` | Scheduled satellite passes |
| `pass_executions` | Pass execution records |
| `commands` | Command submissions and approvals |
| `command_definitions` | Command catalog with parameters |
| `telemetry_raw` | Raw telemetry frames (TimescaleDB) |
| `anomalies` | Detected anomalies |
| `alerts` | Active and historical alerts |

## Components

| Component | Description |
|-----------|-------------|
| [Pass Aggregate](./components/pass-aggregate.md) | Pass scheduling and execution |
| [Command Workflow](./components/command-workflow.md) | Two-person authorization |
| [Telemetry Streaming](./components/telemetry-streaming.md) | Real-time TM processing |

## Integrations

| Integration | Purpose |
|-------------|---------|
| [Space-Track](./integrations/space-track.md) | TLE and CDM data |
| [AWS Ground Station](./integrations/aws-groundstation.md) | Contact scheduling |
| [SDR Hardware](./integrations/sdr-hardware.md) | RF signal reception |
| [Antenna Controller](./integrations/antenna-controller.md) | Antenna pointing |

## Key Flows

### Pass Execution Flow

```
1. Scheduler triggers 30 min before AOS
   └── PassService.PrepareExecution()
   
2. Pre-pass preparation
   ├── Verify TLE freshness
   ├── Calculate trajectory
   ├── Configure SDR frequency
   └── Start antenna tracking

3. AOS (Acquisition of Signal)
   ├── Record actual AOS time
   ├── Start telemetry capture
   └── Enable command queue

4. Pass execution
   ├── Stream telemetry to NATS
   ├── Process commands from queue
   └── Monitor signal quality

5. LOS (Loss of Signal)
   ├── Record actual LOS time
   ├── Stop telemetry capture
   └── Park antenna

6. Post-pass
   ├── Generate execution report
   ├── Run limit checks
   └── Detect anomalies
```

### Command Authorization Flow

```
1. Operator submits command
   └── CommandService.Submit()
   
2. System evaluates hazard level
   ├── SAFE → Approved automatically
   └── CAUTION/CRITICAL → Awaiting approval

3. Approver reviews command
   ├── Different user required
   ├── Must have command.approve permission
   └── CommandService.Approve()

4. Command queued for next pass
   └── status = QUEUED

5. During pass, command sent
   ├── status = SENT
   └── Wait for acknowledgment

6. Spacecraft acknowledges
   └── status = ACKED
```

## Telemetry Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    SDR      │───►│   Decoder   │───►│    NATS     │───►│  Consumers  │
│  Hardware   │    │  (CCSDS)    │    │  JetStream  │    │             │
└─────────────┘    └─────────────┘    └──────┬──────┘    └─────────────┘
                                             │
                        ┌────────────────────┼────────────────────┐
                        │                    │                    │
                        ▼                    ▼                    ▼
                 ┌─────────────┐      ┌─────────────┐      ┌─────────────┐
                 │  Archiver   │      │   Limit     │      │  WebSocket  │
                 │ (Timescale) │      │  Checker    │      │   Gateway   │
                 └─────────────┘      └──────┬──────┘      └─────────────┘
                                             │
                                             ▼
                                      ┌─────────────┐
                                      │   Alerts    │
                                      └─────────────┘
```

## Configuration

```yaml
groundstation:
  tle:
    refresh_interval: 6h
    max_age_hours: 72
    source: space-track  # or celestrak
    
  pass:
    prediction_days: 7
    min_elevation_deg: 5
    schedule_buffer_min: 5
    
  telemetry:
    frame_rate_limit: 1000  # frames/sec
    retention_raw_days: 7
    retention_1min_days: 90
    retention_1hour_years: 5
    
  command:
    require_approval_hazard: critical
    approval_timeout_hours: 24
    command_expiry_hours: 72
    
  alerts:
    escalation_minutes: [5, 15, 60]
    critical_channels: [email, sms, pagerduty]
```

## Metrics

```prometheus
# Pass metrics
p9e_passes_scheduled_total{satellite_id, ground_station_id}
p9e_passes_executed_total{satellite_id, status}
p9e_pass_duration_seconds{satellite_id}

# Telemetry metrics
p9e_telemetry_frames_total{satellite_id}
p9e_telemetry_latency_seconds{satellite_id}
p9e_telemetry_errors_total{satellite_id, error_type}

# Command metrics
p9e_commands_total{satellite_id, status, hazard_level}
p9e_command_latency_seconds{satellite_id}

# Alert metrics
p9e_alerts_active{satellite_id, severity}
p9e_alerts_total{satellite_id, severity, type}
```

## Related Documents

- [System Overview](../00-architecture/system-overview.md)
- [Data Architecture](../00-architecture/data-architecture.md)
- [Satellite Module](../03-satellite/README.md)
