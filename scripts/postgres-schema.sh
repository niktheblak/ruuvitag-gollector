./docker-run.sh \
  postgres-schema \
  --postgres.conn "postgresql://ruuvitag:changeme@postgres:5432/ruuvitag?sslmode=disable" \
  --postgres.table measurements
