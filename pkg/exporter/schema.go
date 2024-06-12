package exporter

import (
	"context"
)

type SchemaCreator interface {
	InitSchema(ctx context.Context) error
}
