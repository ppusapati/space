# gi_ml

Geospatial-intelligence predictive analytics (GI-FR-030..033).

Provides:

* **Flood-risk index** — combines elevation, distance-from-river,
  rainfall, and saturated-soil terms into a 0..1 risk surface.
* **Drought index** — Standardised Precipitation Index (SPI) over a
  rolling window of monthly precipitation.
* **Crop NDVI trend forecaster** — least-squares linear-trend forecaster
  with confidence intervals.
* **Urban-growth Markov projector** — applies a per-pixel transition
  matrix to project a categorical land-cover map forward in time.
