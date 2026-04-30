# eo-ml-worker

Kafka consumer that pulls Earth-observation inference jobs from
`eo.inference.jobs.v1`, runs the appropriate `eo_ml` pipeline, writes
the result to S3, and publishes a completion event to
`eo.inference.results.v1`.

## Job message

```json
{
  "job_id": "ULID",
  "tenant_id": "tenant-1",
  "model_name": "yolov8-port",
  "kind": "detect",
  "input_uri": "s3://bucket/scenes/x.npy",
  "output_uri": "s3://bucket/results/x.json"
}
```

## Result message

```json
{
  "job_id": "ULID",
  "status": "ok" | "failed",
  "output_uri": "s3://bucket/results/x.json",
  "error": null | "string"
}
```
