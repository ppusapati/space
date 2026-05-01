# gi-ml-worker

Kafka consumer that pulls geospatial-intelligence prediction jobs from
`gi.predict.jobs.v1`, runs the matching `gi_ml` predictor, writes the
result to S3, and publishes a completion event to
`gi.predict.results.v1`.

## Job kinds

* `flood_risk` — composite flood-risk index over four input rasters.
* `drought_spi` — Standardised Precipitation Index over a rolling
  window.
* `forecast_ndvi` — least-squares NDVI trend forecast.
* `urban_markov` — Markov-chain land-cover projection.
