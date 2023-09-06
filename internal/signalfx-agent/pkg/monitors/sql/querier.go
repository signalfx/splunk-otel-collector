package sql

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/davecgh/go-spew/spew"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

type querier struct {
	query                     *Query
	valueColumnNamesToMetrics map[string]*Metric
	metricToIndex             map[*Metric]int
	dimensionColumnSets       []map[string]bool
	dimensions                [][]*types.Dimension
	compiledExprs             []*vm.Program
	rowSliceCached            []interface{}
	logger                    logrus.FieldLogger
	logQueries                bool
}

func newQuerier(query *Query, logQueries bool, logger logrus.FieldLogger) (*querier, error) {
	valueColumnNamesToMetrics := map[string]*Metric{}
	metricToIndex := map[*Metric]int{}

	for i, m := range query.Metrics {
		valueColumnNamesToMetrics[strings.ToLower(m.ValueColumn)] = &query.Metrics[i]
		metricToIndex[&query.Metrics[i]] = i
	}

	dimensionColumnSets := make([]map[string]bool, len(query.Metrics))
	for i := range dimensionColumnSets {
		dimensionColumnSets[i] = map[string]bool{}
	}

	dimensions := make([][]*types.Dimension, len(query.Metrics))

	for i, m := range query.Metrics {
		for _, dim := range m.DimensionColumns {
			dimensionColumnSets[i][strings.ToLower(dim)] = true
		}

		dimensions[i] = make([]*types.Dimension, 0)
		for dim, propColumns := range m.DimensionPropertyColumns {
			dimColumn := strings.ToLower(dim)
			if dimensionColumnSets[i][dimColumn] {
				props := make(map[string]string)
				for _, p := range propColumns {
					props[strings.ToLower(p)] = ""
				}
				dimensions[i] = append(dimensions[i], &types.Dimension{
					Name:       dimColumn,
					Properties: props,
				})
			}
		}
	}

	compiledExprs := make([]*vm.Program, len(query.DatapointExpressions))
	for i, dpExpr := range query.DatapointExpressions {
		ruleProg, err := expr.Compile(dpExpr)
		if err != nil {
			return nil, fmt.Errorf("failed to compile expression: %v", err)
		}

		compiledExprs[i] = ruleProg
	}

	return &querier{
		query:                     query,
		valueColumnNamesToMetrics: valueColumnNamesToMetrics,
		metricToIndex:             metricToIndex,
		dimensionColumnSets:       dimensionColumnSets,
		dimensions:                dimensions,
		compiledExprs:             compiledExprs,
		logger:                    logger.WithField("statement", query.Query),
		logQueries:                logQueries,
	}, nil
}

func (q *querier) doQuery(ctx context.Context, database *sql.DB, output types.Output) error {
	rows, err := database.QueryContext(ctx, q.query.Query, q.query.Params...)
	if err != nil {
		return fmt.Errorf("error executing statement %s: %v", q.query.Query, err)
	}
	for rows.Next() {
		dps, dims, err := q.convertCurrentRowToDatapointAndDimensions(rows)
		if err != nil {
			rows.Close()
			return err
		}

		output.SendDatapoints(dps...)

		for i := range dims {
			for _, dim := range dims[i] {
				output.SendDimensionUpdate(dim)
			}
		}
	}
	return rows.Close()
}

func (q *querier) convertCurrentRowToDatapointAndDimensions(rows *sql.Rows) ([]*datapoint.Datapoint, [][]*types.Dimension, error) {
	rowSlice, err := q.getRowSlice(rows)
	if err != nil {
		return nil, nil, err
	}
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	if err := rows.Scan(rowSlice...); err != nil {
		return nil, nil, err
	}
	if q.logQueries {
		q.logger.Info("Got results %s", spew.Sdump(rowSlice))
	}

	dps := make([]*datapoint.Datapoint, 0, len(q.query.Metrics)+len(q.query.DatapointExpressions))
	var dims [][]*types.Dimension

	if len(q.query.Metrics) > 0 {
		var err error
		var structuredDPs []*datapoint.Datapoint
		structuredDPs, dims, err = q.convertCurrentRowStructured(rowSlice, columnNames)
		if err != nil {
			q.logger.WithError(err).Warn("Failed to convert row to datapoints and dimensions")
		}

		dps = append(dps, structuredDPs...)
	}

	if len(q.query.DatapointExpressions) > 0 {
		var err error
		exprDPs := q.convertCurrentRowExpressions(rowSlice, columnNames)
		if err != nil {
			q.logger.WithError(err).Warn("Failed to convert row to datapoints")
		}
		dps = append(dps, exprDPs...)
	}

	return dps, dims, nil
}

func (q *querier) convertCurrentRowExpressions(rowSlice []interface{}, columnNames []string) []*datapoint.Datapoint {
	dps := make([]*datapoint.Datapoint, 0, len(q.compiledExprs))

	for _, compiled := range q.compiledExprs {
		env := newExprEnv(rowSlice, columnNames)

		if q.logQueries {
			q.logger.WithFields(logrus.Fields{
				"expression": compiled.Source.Content(),
				"context":    spew.Sdump(env),
			}).Info("Evaluating expression")
		}

		dp, err := vm.Run(compiled, env)
		if err != nil {
			q.logger.WithError(err).Errorf("Failed to evaluate expression '%s'", compiled.Source.Content())
			continue
		}

		// It is fine if the expression evaluates to nil (e.g. the row didn't
		// have some expected data so no metric is appropriate)
		if dp == nil {
			continue
		}

		if v, ok := dp.(*datapoint.Datapoint); ok {
			dps = append(dps, v)
		} else {
			q.logger.WithField("expression", compiled.Source.Content()).WithField("result", dp).Warn("Result of expression is not a datapoint")
			continue
		}
	}

	return dps
}

func (q *querier) convertCurrentRowStructured(rowSlice []interface{}, columnNames []string) ([]*datapoint.Datapoint, [][]*types.Dimension, error) {
	// Clone all properties before updating them
	for i := range q.dimensions {
		for j := range q.dimensions[i] {
			props := make(map[string]string)
			for propName := range q.dimensions[i][j].Properties {
				props[propName] = ""
			}
			q.dimensions[i][j] = &types.Dimension{
				Name:       q.dimensions[i][j].Name,
				Properties: props,
			}
		}
	}

	dps := make([]*datapoint.Datapoint, 0, len(q.query.Metrics))
	for i := range q.query.Metrics {
		dps = append(dps, q.query.Metrics[i].NewDatapoint())
	}

	for i := range rowSlice {
		switch v := rowSlice[i].(type) {
		case *sql.NullFloat64:
			if !v.Valid {
				return nil, nil, fmt.Errorf("column %d is null", i)
			}

			metric, ok := q.valueColumnNamesToMetrics[strings.ToLower(columnNames[i])]
			if !ok || metric == nil {
				// This is a logical error in the code, not user input error
				panic("valueColumn was not properly mapped to metric")
			}

			dp := dps[q.metricToIndex[metric]]
			dp.Value = datapoint.NewFloatValue(v.Float64)

		case *sql.NullString:
			dimVal := v.String
			if !v.Valid {
				// Make sure the value gets properly blanked out since we are
				// reusing rowSlice between rows/queries.
				dimVal = ""
			}
			for j := range q.query.Metrics {
				for _, dim := range q.dimensions[j] {
					if strings.EqualFold(dim.Name, columnNames[i]) {
						dim.Name = columnNames[i]
						dim.Value = dimVal
					}
					for k := range dim.Properties {
						if strings.EqualFold(columnNames[i], k) {
							dim.Properties[k] = utils.TruncateDimensionValue(dimVal)
						}
					}
				}
				if !q.dimensionColumnSets[j][strings.ToLower(columnNames[i])] {
					continue
				}

				dps[j].Dimensions[columnNames[i]] = dimVal
			}
		}
	}

	var n int
	for i := range dps {
		if dps[i].Value == nil {
			q.logger.Warnf("Metric %s's value column '%s' did not correspond to a value\nrowSlice: %s", q.query.Metrics[i].MetricName, q.query.Metrics[i].ValueColumn, spew.Sdump(rowSlice))
			continue
		}
		dps[n] = dps[i]
		n++
	}

	return dps[:n], q.dimensions, nil
}

func (q *querier) getRowSlice(rows *sql.Rows) ([]interface{}, error) {
	if len(q.query.DatapointExpressions) > 0 {
		return q.getRowSliceForExpression(rows)
	}
	return q.getRowSliceForStructured(rows)
}

func (q *querier) getRowSliceForStructured(rows *sql.Rows) ([]interface{}, error) {
	if q.rowSliceCached != nil {
		return q.rowSliceCached, nil
	}

	cts, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	dimColsSeen := map[string]bool{}
	propColsSeen := map[string]bool{}
	rowSlice := make([]interface{}, len(cts))
OUTER:
	for i, ct := range cts {
		for _, metric := range q.query.Metrics {

			if strings.EqualFold(ct.Name(), metric.ValueColumn) {
				// Values are always numeric
				rowSlice[i] = &sql.NullFloat64{}
				// Can't also be a dimension column or value in another metric
				continue OUTER
			}

			for _, propertyColumns := range metric.DimensionPropertyColumns {
				for _, colName := range propertyColumns {
					if strings.EqualFold(ct.Name(), colName) {
						propColsSeen[colName] = true
					}
				}
			}

			for _, colName := range metric.DimensionColumns {
				if strings.EqualFold(ct.Name(), colName) {
					dimColsSeen[colName] = true
					rowSlice[i] = &sql.NullString{}
					// Cannot also be a value column if dimension
					continue OUTER
				}
			}

		}
		// This column is unused in generating metrics so just make it a string
		rowSlice[i] = &sql.NullString{}
	}

	for _, metric := range q.query.Metrics {
		for _, dimCol := range metric.DimensionColumns {
			if !dimColsSeen[dimCol] {
				return nil, fmt.Errorf("dimension column '%s' does not exist", dimCol)
			}
		}

		for _, propertyColumns := range metric.DimensionPropertyColumns {
			for _, propCol := range propertyColumns {
				if !propColsSeen[propCol] {
					return nil, fmt.Errorf("property column '%s' does not exist", propCol)
				}
			}
		}
	}

	q.rowSliceCached = rowSlice
	return rowSlice, nil
}

// This is a more flexible row slice generator that looks at the db's type
// reporting to generate a row slice instead of the user's config.
func (q *querier) getRowSliceForExpression(rows *sql.Rows) ([]interface{}, error) {
	if q.rowSliceCached != nil {
		return q.rowSliceCached, nil
	}

	cts, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	if err := validateStructuredMetrics(q.query, cts); err != nil {
		return nil, err
	}

	rowSlice := make([]interface{}, len(cts))
	for i, ct := range cts {
		scanType := ct.ScanType()
		if scanType.Kind() == reflect.Ptr {
			scanType = scanType.Elem()
		}
		switch {
		case scanType.ConvertibleTo(reflect.TypeOf(time.Time{})):
			rowSlice[i] = &sql.NullTime{}
		case scanType.ConvertibleTo(reflect.TypeOf(float64(0))):
			rowSlice[i] = &sql.NullFloat64{}
		case scanType.Kind() == reflect.Bool:
			rowSlice[i] = &sql.NullBool{}
		case scanType.Kind() == reflect.String:
			rowSlice[i] = &sql.NullString{}
		default:
			intf := interface{}(nil)
			rowSlice[i] = &intf
		}
	}

	q.rowSliceCached = rowSlice
	return rowSlice, nil
}

func validateStructuredMetrics(query *Query, columnTypes []*sql.ColumnType) error {
	dimColsSeen := map[string]bool{}
	propColsSeen := map[string]bool{}

OUTER:
	for _, ct := range columnTypes {
		for _, metric := range query.Metrics {
			if strings.EqualFold(ct.Name(), metric.ValueColumn) {
				// Can't also be a dimension column or value in another metric
				continue OUTER
			}

			for _, propertyColumns := range metric.DimensionPropertyColumns {
				for _, colName := range propertyColumns {
					if strings.EqualFold(ct.Name(), colName) {
						propColsSeen[colName] = true
					}
				}
			}

			for _, colName := range metric.DimensionColumns {
				if strings.EqualFold(ct.Name(), colName) {
					dimColsSeen[colName] = true
					// Cannot also be a value column if dimension
					continue OUTER
				}
			}

		}
	}

	for _, metric := range query.Metrics {
		for _, dimCol := range metric.DimensionColumns {
			if !dimColsSeen[dimCol] {
				return fmt.Errorf("dimension column '%s' does not exist", dimCol)
			}
		}

		for _, propertyColumns := range metric.DimensionPropertyColumns {
			for _, propCol := range propertyColumns {
				if !propColsSeen[propCol] {
					return fmt.Errorf("property column '%s' does not exist", propCol)
				}
			}
		}
	}

	return nil
}
