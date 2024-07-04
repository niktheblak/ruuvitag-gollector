//go:build !postgres

package postgres

func New(ctx context.Context, cfg Config) (exporter.Exporter, error) {
	return exporter.NoOp{ReportedName: "Postgres"}, nil
}
