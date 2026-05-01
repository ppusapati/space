-- gs-rf schema. RF link budgets + measurements per pass.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS link_budgets (
    id                    uuid PRIMARY KEY,
    tenant_id             uuid NOT NULL,
    pass_id               uuid NOT NULL,
    station_id            uuid NOT NULL,
    antenna_id            uuid NOT NULL,
    satellite_id          uuid NOT NULL,
    carrier_freq_hz       bigint NOT NULL CHECK (carrier_freq_hz >= 0),
    tx_power_dbm          double precision NOT NULL DEFAULT 0,
    tx_gain_dbi           double precision NOT NULL DEFAULT 0,
    rx_gain_dbi           double precision NOT NULL DEFAULT 0,
    rx_noise_temp_k       double precision NOT NULL DEFAULT 0,
    bandwidth_hz          double precision NOT NULL DEFAULT 0,
    slant_range_km        double precision NOT NULL DEFAULT 0,
    free_space_loss_db    double precision NOT NULL DEFAULT 0,
    atmospheric_loss_db   double precision NOT NULL DEFAULT 0,
    polarization_loss_db  double precision NOT NULL DEFAULT 0,
    pointing_loss_db      double precision NOT NULL DEFAULT 0,
    predicted_eb_n0_db    double precision NOT NULL DEFAULT 0,
    predicted_snr_db      double precision NOT NULL DEFAULT 0,
    link_margin_db        double precision NOT NULL DEFAULT 0,
    notes                 text NOT NULL DEFAULT '',
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    created_by            text NOT NULL DEFAULT 'system',
    updated_by            text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS link_budgets_tenant_idx     ON link_budgets (tenant_id);
CREATE INDEX IF NOT EXISTS link_budgets_pass_idx       ON link_budgets (pass_id);
CREATE INDEX IF NOT EXISTS link_budgets_station_idx    ON link_budgets (station_id);
CREATE INDEX IF NOT EXISTS link_budgets_satellite_idx  ON link_budgets (satellite_id);
CREATE INDEX IF NOT EXISTS link_budgets_created_at_idx ON link_budgets (created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS link_measurements (
    id                uuid PRIMARY KEY,
    tenant_id         uuid NOT NULL,
    pass_id           uuid NOT NULL,
    station_id        uuid NOT NULL,
    antenna_id        uuid NOT NULL,
    sampled_at        timestamptz NOT NULL,
    rssi_dbm          double precision NOT NULL DEFAULT 0,
    snr_db            double precision NOT NULL DEFAULT 0,
    ber               double precision NOT NULL DEFAULT 0 CHECK (ber >= 0 AND ber <= 1),
    fer               double precision NOT NULL DEFAULT 0 CHECK (fer >= 0 AND fer <= 1),
    frequency_hz      bigint NOT NULL DEFAULT 0,
    doppler_shift_hz  double precision NOT NULL DEFAULT 0,
    created_at        timestamptz NOT NULL DEFAULT now(),
    created_by        text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS link_measurements_tenant_idx    ON link_measurements (tenant_id);
CREATE INDEX IF NOT EXISTS link_measurements_pass_idx      ON link_measurements (pass_id);
CREATE INDEX IF NOT EXISTS link_measurements_sampled_idx   ON link_measurements (sampled_at DESC, id DESC);
