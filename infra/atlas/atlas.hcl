// infra/atlas/atlas.hcl
//
// TASK-P0-DB-001 — platform-wide cluster migrations driven by Atlas in
// **versioned mode** (manual handwritten SQL files; Atlas applies them in
// order with checksum verification).
//
// Scope of this Atlas project:
//   • Cluster-wide concerns ONLY — extensions (TimescaleDB, PostGIS,
//     pg_trgm), retention policies, hypertable promotions, cross-DB
//     conventions.
//   • Per-service schema migrations live under
//     services/<svc>/db/schema/*.sql and are owned by that service.
//
// The migration runner in services/packages/db/migrate embeds these files
// via embed.FS and applies them at service boot before serving any RPCs.

env "local" {
  // Default for `docker compose up` workflow.
  // Set CHETANA_DB_URL_LOCAL to override; otherwise the docker-compose
  // Postgres credentials apply.
  url = getenv("CHETANA_DB_URL_LOCAL") != "" ? getenv("CHETANA_DB_URL_LOCAL") : "postgres://p9e:p9e@localhost:5432/postgres?sslmode=disable&search_path=public"
  dev = "docker://postgis/16/dev?search_path=public"

  migration {
    dir    = "file://../../services/packages/db/migrate/migrations"
    format = atlas
  }
}

env "test" {
  // Used by infra/atlas/test/migrate_test.go (Testcontainers).
  url = getenv("CHETANA_DB_URL_TEST")
  dev = "docker://postgis/16/dev?search_path=public"

  migration {
    dir    = "file://../../services/packages/db/migrate/migrations"
    format = atlas
  }
}

env "prod" {
  // GovCloud RDS / self-managed Postgres. CHETANA_DB_URL_PROD set by the
  // Helm pre-deploy hook (TASK-P0-INFRA-001) from a Secrets Manager source.
  url = getenv("CHETANA_DB_URL_PROD")
  dev = "docker://postgis/16/dev?search_path=public"

  migration {
    dir              = "file://../../services/packages/db/migrate/migrations"
    format           = atlas
    revisions_schema = "public"
  }

  // Lint configuration: any new migration must pass these checks.
  // Catches dangerous ops like dropping NOT NULL on populated columns.
  lint {
    destructive {
      error = true
    }
    data_depend {
      error = true
    }
    incompatible {
      error = true
    }
  }
}
