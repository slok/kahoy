package storage

import (
	"context"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
)

type noopReportRepository struct {
	logger log.Logger
}

// NewNoopReportRepository returns a new NOOP report repository
func NewNoopReportRepository(logger log.Logger) ReportRepository {
	return noopReportRepository{
		logger: logger.WithValues(log.Kv{"app-svc": "storage.noopReportRepository"}),
	}
}

func (n noopReportRepository) StoreReport(ctx context.Context, report model.Report) error {
	n.logger.Debugf("ignoring report store by NOOP report repository")
	return nil
}
