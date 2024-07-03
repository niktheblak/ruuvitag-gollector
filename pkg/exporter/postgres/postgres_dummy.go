//go:build !postgres

package postgres

func New(ctx context.Context, psqlInfo, table, timeColumn string, logger *slog.Logger) (exporter.Exporter, error) {
	return exporter.NoOp{ReportedName: "Postgres"}, nil
}
