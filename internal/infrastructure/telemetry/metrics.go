package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Metrics contém as métricas de negócio da aplicação
type Metrics struct {
	// Counters
	EntitiesCreated metric.Int64Counter
	EntitiesUpdated metric.Int64Counter
	EntitiesDeleted metric.Int64Counter

	// Histograms
	OperationDuration metric.Float64Histogram
}

// NewMetrics creates business metrics instruments using the provided meter.
func NewMetrics(meter metric.Meter) (*Metrics, error) {
	entitiesCreated, err := meter.Int64Counter(
		"entities_created_total",
		metric.WithDescription("Total number of entities created"),
		metric.WithUnit("{entity}"),
	)
	if err != nil {
		return nil, err
	}

	entitiesUpdated, err := meter.Int64Counter(
		"entities_updated_total",
		metric.WithDescription("Total number of entities updated"),
		metric.WithUnit("{entity}"),
	)
	if err != nil {
		return nil, err
	}

	entitiesDeleted, err := meter.Int64Counter(
		"entities_deleted_total",
		metric.WithDescription("Total number of entities deleted (soft delete)"),
		metric.WithUnit("{entity}"),
	)
	if err != nil {
		return nil, err
	}

	operationDuration, err := meter.Float64Histogram(
		"entities_operation_duration_seconds",
		metric.WithDescription("Duration of entities operations in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		EntitiesCreated:   entitiesCreated,
		EntitiesUpdated:   entitiesUpdated,
		EntitiesDeleted:   entitiesDeleted,
		OperationDuration: operationDuration,
	}, nil
}

// RecordCreate registra uma criação de pessoa
func (m *Metrics) RecordCreate(ctx context.Context) {
	if m != nil && m.EntitiesCreated != nil {
		m.EntitiesCreated.Add(ctx, 1)
	}
}

// RecordUpdate registra uma atualização de pessoa
func (m *Metrics) RecordUpdate(ctx context.Context) {
	if m != nil && m.EntitiesUpdated != nil {
		m.EntitiesUpdated.Add(ctx, 1)
	}
}

// RecordDelete registra uma deleção de pessoa
func (m *Metrics) RecordDelete(ctx context.Context) {
	if m != nil && m.EntitiesDeleted != nil {
		m.EntitiesDeleted.Add(ctx, 1)
	}
}

// RecordDuration registra a duração de uma operação
func (m *Metrics) RecordDuration(ctx context.Context, durationSeconds float64, operation string) {
	if m != nil && m.OperationDuration != nil {
		m.OperationDuration.Record(ctx, durationSeconds,
			metric.WithAttributes(attribute.String("operation", operation)),
		)
	}
}
