package sqleng

import (
	"container/list"
	"context"
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana/pkg/setting"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/tsdb"

	"github.com/grafana/grafana/pkg/components/null"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"xorm.io/core"
	"xorm.io/xorm"
)

// MetaKeyExecutedQueryString is the key where the executed query should get stored
const MetaKeyExecutedQueryString = "executedQueryString"

// SqlMacroEngine interpolates macros into sql. It takes in the Query to have access to query context and
// timeRange to be able to generate queries that use from and to.
type SqlMacroEngine interface {
	Interpolate(query *tsdb.Query, timeRange *tsdb.TimeRange, sql string) (string, error)
}

// SqlQueryResultTransformer transforms a query result row to RowValues with proper types.
type SqlQueryResultTransformer interface {
	// TransformQueryResult transforms a query result row to RowValues with proper types.
	TransformQueryResult(columnTypes []*sql.ColumnType, rows *core.Rows) (tsdb.RowValues, error)
	// TransformQueryError transforms a query error.
	TransformQueryError(err error) error
}

type engineCacheType struct {
	cache    map[int64]*xorm.Engine
	versions map[int64]int
	sync.Mutex
}

var engineCache = engineCacheType{
	cache:    make(map[int64]*xorm.Engine),
	versions: make(map[int64]int),
}

var sqlIntervalCalculator = tsdb.NewIntervalCalculator(nil)

//nolint:gocritic
// NewXormEngine is an xorm.Engine factory, that can be stubbed by tests.
var NewXormEngine = func(driverName string, connectionString string) (*xorm.Engine, error) {
	return xorm.NewEngine(driverName, connectionString)
}

const timeEndColumnName = "timeend"

type sqlQueryEndpoint struct {
	macroEngine            SqlMacroEngine
	queryResultTransformer SqlQueryResultTransformer
	engine                 *xorm.Engine
	timeColumnNames        []string
	metricColumnTypes      []string
	log                    log.Logger
}

type SqlQueryEndpointConfiguration struct {
	DriverName        string
	Datasource        *models.DataSource
	ConnectionString  string
	TimeColumnNames   []string
	MetricColumnTypes []string
}

var NewSqlQueryEndpoint = func(config *SqlQueryEndpointConfiguration, queryResultTransformer SqlQueryResultTransformer, macroEngine SqlMacroEngine, log log.Logger) (tsdb.TsdbQueryEndpoint, error) {
	queryEndpoint := sqlQueryEndpoint{
		queryResultTransformer: queryResultTransformer,
		macroEngine:            macroEngine,
		timeColumnNames:        []string{"time"},
		log:                    log,
	}

	if len(config.TimeColumnNames) > 0 {
		queryEndpoint.timeColumnNames = config.TimeColumnNames
	}

	if len(config.MetricColumnTypes) > 0 {
		queryEndpoint.metricColumnTypes = config.MetricColumnTypes
	}

	engineCache.Lock()
	defer engineCache.Unlock()

	if engine, present := engineCache.cache[config.Datasource.Id]; present {
		if version := engineCache.versions[config.Datasource.Id]; version == config.Datasource.Version {
			queryEndpoint.engine = engine
			return &queryEndpoint, nil
		}
	}

	engine, err := NewXormEngine(config.DriverName, config.ConnectionString)
	if err != nil {
		return nil, err
	}

	maxOpenConns := config.Datasource.JsonData.Get("maxOpenConns").MustInt(0)
	engine.SetMaxOpenConns(maxOpenConns)
	maxIdleConns := config.Datasource.JsonData.Get("maxIdleConns").MustInt(2)
	engine.SetMaxIdleConns(maxIdleConns)
	connMaxLifetime := config.Datasource.JsonData.Get("connMaxLifetime").MustInt(14400)
	engine.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)

	engineCache.versions[config.Datasource.Id] = config.Datasource.Version
	engineCache.cache[config.Datasource.Id] = engine
	queryEndpoint.engine = engine

	return &queryEndpoint, nil
}

const rowLimit = 1000000

// Query is the main function for the SqlQueryEndpoint
func (e *sqlQueryEndpoint) Query(ctx context.Context, dsInfo *models.DataSource, tsdbQuery *tsdb.TsdbQuery) (*tsdb.Response, error) {
	result := &tsdb.Response{
		Results: make(map[string]*tsdb.QueryResult),
	}

	var wg sync.WaitGroup

	for _, query := range tsdbQuery.Queries {
		rawSQL := query.Model.Get("rawSql").MustString()
		if rawSQL == "" {
			continue
		}

		queryResult := &tsdb.QueryResult{Meta: simplejson.New(), RefId: query.RefId}
		result.Results[query.RefId] = queryResult

		// global substitutions
		rawSQL, err := Interpolate(query, tsdbQuery.TimeRange, rawSQL)
		if err != nil {
			queryResult.Error = err
			continue
		}

		// datasource specific substitutions
		rawSQL, err = e.macroEngine.Interpolate(query, tsdbQuery.TimeRange, rawSQL)
		if err != nil {
			queryResult.Error = err
			continue
		}

		queryResult.Meta.Set(MetaKeyExecutedQueryString, rawSQL)

		wg.Add(1)

		go func(rawSQL string, query *tsdb.Query, queryResult *tsdb.QueryResult) {
			defer wg.Done()
			session := e.engine.NewSession()
			defer session.Close()
			db := session.DB()

			rows, err := db.Query(rawSQL)
			if err != nil {
				queryResult.Error = e.queryResultTransformer.TransformQueryError(err)
				return
			}

			defer rows.Close()

			format := query.Model.Get("format").MustString("time_series")

			switch format {
			case "time_series":
				err := e.transformToTimeSeries(query, rows, queryResult, tsdbQuery)
				if err != nil {
					queryResult.Error = err
					return
				}
			case "table":
				err := e.transformToTable(query, rows, queryResult, tsdbQuery)
				if err != nil {
					queryResult.Error = err
					return
				}
			}
		}(rawSQL, query, queryResult)
	}
	wg.Wait()

	return result, nil
}

// global macros/substitutions for all sql datasources
var Interpolate = func(query *tsdb.Query, timeRange *tsdb.TimeRange, sql string) (string, error) {
	minInterval, err := tsdb.GetIntervalFrom(query.DataSource, query.Model, time.Second*60)
	if err != nil {
		return sql, nil
	}
	interval := sqlIntervalCalculator.Calculate(timeRange, minInterval)

	sql = strings.ReplaceAll(sql, "$__interval_ms", strconv.FormatInt(interval.Milliseconds(), 10))
	sql = strings.ReplaceAll(sql, "$__interval", interval.Text)
	sql = strings.ReplaceAll(sql, "$__unixEpochFrom()", fmt.Sprintf("%d", timeRange.GetFromAsSecondsEpoch()))
	sql = strings.ReplaceAll(sql, "$__unixEpochTo()", fmt.Sprintf("%d", timeRange.GetToAsSecondsEpoch()))

	return sql, nil
}

func (e *sqlQueryEndpoint) transformToTable(query *tsdb.Query, rows *core.Rows, result *tsdb.QueryResult, tsdbQuery *tsdb.TsdbQuery) error {
	columnNames, err := rows.Columns()
	columnCount := len(columnNames)

	if err != nil {
		return err
	}

	rowCount := 0
	timeIndex := -1
	timeEndIndex := -1

	table := &tsdb.Table{
		Columns: make([]tsdb.TableColumn, columnCount),
		Rows:    make([]tsdb.RowValues, 0),
	}

	for i, name := range columnNames {
		table.Columns[i].Text = name

		for _, tc := range e.timeColumnNames {
			if name == tc {
				timeIndex = i
				break
			}

			if timeIndex >= 0 && name == timeEndColumnName {
				timeEndIndex = i
				break
			}
		}
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}

	for ; rows.Next(); rowCount++ {
		if rowCount > rowLimit {
			return fmt.Errorf("query row limit exceeded, limit %d", rowLimit)
		}

		values, err := e.queryResultTransformer.TransformQueryResult(columnTypes, rows)
		if err != nil {
			return err
		}

		// converts column named time and timeend to unix timestamp in milliseconds
		// to make native mssql datetime types and epoch dates work in
		// annotation and table queries.
		ConvertSqlTimeColumnToEpochMs(values, timeIndex)
		ConvertSqlTimeColumnToEpochMs(values, timeEndIndex)
		table.Rows = append(table.Rows, values)
	}

	result.Tables = append(result.Tables, table)
	result.Meta.Set("rowCount", rowCount)
	return nil
}

func newProcessCfg(query *tsdb.Query, tsdbQuery *tsdb.TsdbQuery, rows *core.Rows) (*processCfg, error) {
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fillMissing := query.Model.Get("fill").MustBool(false)

	cfg := &processCfg{
		rowCount:           0,
		columnTypes:        columnTypes,
		columnNames:        columnNames,
		rows:               rows,
		timeIndex:          -1,
		metricIndex:        -1,
		metricPrefix:       false,
		fillMissing:        fillMissing,
		seriesByQueryOrder: list.New(),
		pointsBySeries:     make(map[string]*tsdb.TimeSeries),
		tsdbQuery:          tsdbQuery,
	}
	return cfg, nil
}

func (e *sqlQueryEndpoint) transformToTimeSeries(query *tsdb.Query, rows *core.Rows, result *tsdb.QueryResult,
	tsdbQuery *tsdb.TsdbQuery) error {
	cfg, err := newProcessCfg(query, tsdbQuery, rows)
	if err != nil {
		return err
	}

	// check columns of resultset: a column named time is mandatory
	// the first text column is treated as metric name unless a column named metric is present
	for i, col := range cfg.columnNames {
		for _, tc := range e.timeColumnNames {
			if col == tc {
				cfg.timeIndex = i
				continue
			}
		}
		switch col {
		case "metric":
			cfg.metricIndex = i
		default:
			if cfg.metricIndex == -1 {
				columnType := cfg.columnTypes[i].DatabaseTypeName()

				for _, mct := range e.metricColumnTypes {
					if columnType == mct {
						cfg.metricIndex = i
						continue
					}
				}
			}
		}
	}

	// use metric column as prefix with multiple value columns
	if cfg.metricIndex != -1 && len(cfg.columnNames) > 3 {
		cfg.metricPrefix = true
	}

	if cfg.timeIndex == -1 {
		return fmt.Errorf("Found no column named %s", strings.Join(e.timeColumnNames, " or "))
	}

	if cfg.fillMissing {
		cfg.fillInterval = query.Model.Get("fillInterval").MustFloat64() * 1000
		switch query.Model.Get("fillMode").MustString() {
		case "null":
		case "previous":
			cfg.fillPrevious = true
		case "value":
			cfg.fillValue.Float64 = query.Model.Get("fillValue").MustFloat64()
			cfg.fillValue.Valid = true
		}
	}

	for rows.Next() {
		if err := e.processRow(cfg); err != nil {
			return err
		}
	}

	for elem := cfg.seriesByQueryOrder.Front(); elem != nil; elem = elem.Next() {
		key := elem.Value.(string)
		result.Series = append(result.Series, cfg.pointsBySeries[key])
		if !cfg.fillMissing {
			continue
		}

		series := cfg.pointsBySeries[key]
		// fill in values from last fetched value till interval end
		intervalStart := series.Points[len(series.Points)-1][1].Float64
		intervalEnd := float64(tsdbQuery.TimeRange.MustGetTo().UnixNano() / 1e6)

		if cfg.fillPrevious {
			if len(series.Points) > 0 {
				cfg.fillValue = series.Points[len(series.Points)-1][0]
			} else {
				cfg.fillValue.Valid = false
			}
		}

		// align interval start
		intervalStart = math.Floor(intervalStart/cfg.fillInterval) * cfg.fillInterval
		for i := intervalStart + cfg.fillInterval; i < intervalEnd; i += cfg.fillInterval {
			series.Points = append(series.Points, tsdb.TimePoint{cfg.fillValue, null.FloatFrom(i)})
			cfg.rowCount++
		}
	}

	result.Meta.Set("rowCount", cfg.rowCount)
	return nil
}

type processCfg struct {
	rowCount           int
	columnTypes        []*sql.ColumnType
	columnNames        []string
	rows               *core.Rows
	timeIndex          int
	metricIndex        int
	metricPrefix       bool
	metricPrefixValue  string
	fillMissing        bool
	pointsBySeries     map[string]*tsdb.TimeSeries
	seriesByQueryOrder *list.List
	fillValue          null.Float
	tsdbQuery          *tsdb.TsdbQuery
	fillInterval       float64
	fillPrevious       bool
}

func (e *sqlQueryEndpoint) processRow(cfg *processCfg) error {
	var timestamp float64
	var value null.Float
	var metric string

	if cfg.rowCount > rowLimit {
		return fmt.Errorf("query row limit exceeded, limit %d", rowLimit)
	}

	values, err := e.queryResultTransformer.TransformQueryResult(cfg.columnTypes, cfg.rows)
	if err != nil {
		return err
	}

	// converts column named time to unix timestamp in milliseconds to make
	// native mysql datetime types and epoch dates work in
	// annotation and table queries.
	ConvertSqlTimeColumnToEpochMs(values, cfg.timeIndex)

	switch columnValue := values[cfg.timeIndex].(type) {
	case int64:
		timestamp = float64(columnValue)
	case float64:
		timestamp = columnValue
	default:
		return fmt.Errorf("invalid type for column time, must be of type timestamp or unix timestamp, got: %T %v",
			columnValue, columnValue)
	}

	if cfg.metricIndex >= 0 {
		if columnValue, ok := values[cfg.metricIndex].(string); ok {
			if cfg.metricPrefix {
				cfg.metricPrefixValue = columnValue
			} else {
				metric = columnValue
			}
		} else {
			return fmt.Errorf("column metric must be of type %s. metric column name: %s type: %s but datatype is %T",
				strings.Join(e.metricColumnTypes, ", "), cfg.columnNames[cfg.metricIndex],
				cfg.columnTypes[cfg.metricIndex].DatabaseTypeName(), values[cfg.metricIndex])
		}
	}

	for i, col := range cfg.columnNames {
		if i == cfg.timeIndex || i == cfg.metricIndex {
			continue
		}

		if value, err = ConvertSqlValueColumnToFloat(col, values[i]); err != nil {
			return err
		}

		if cfg.metricIndex == -1 {
			metric = col
		} else if cfg.metricPrefix {
			metric = cfg.metricPrefixValue + " " + col
		}

		series, exist := cfg.pointsBySeries[metric]
		if !exist {
			series = &tsdb.TimeSeries{Name: metric}
			cfg.pointsBySeries[metric] = series
			cfg.seriesByQueryOrder.PushBack(metric)
		}

		if cfg.fillMissing {
			var intervalStart float64
			if !exist {
				intervalStart = float64(cfg.tsdbQuery.TimeRange.MustGetFrom().UnixNano() / 1e6)
			} else {
				intervalStart = series.Points[len(series.Points)-1][1].Float64 + cfg.fillInterval
			}

			if cfg.fillPrevious {
				if len(series.Points) > 0 {
					cfg.fillValue = series.Points[len(series.Points)-1][0]
				} else {
					cfg.fillValue.Valid = false
				}
			}

			// align interval start
			intervalStart = math.Floor(intervalStart/cfg.fillInterval) * cfg.fillInterval

			for i := intervalStart; i < timestamp; i += cfg.fillInterval {
				series.Points = append(series.Points, tsdb.TimePoint{cfg.fillValue, null.FloatFrom(i)})
				cfg.rowCount++
			}
		}

		series.Points = append(series.Points, tsdb.TimePoint{value, null.FloatFrom(timestamp)})

		if setting.Env == setting.Dev {
			e.log.Debug("Rows", "metric", metric, "time", timestamp, "value", value)
		}
	}

	return nil
}

// ConvertSqlTimeColumnToEpochMs converts column named time to unix timestamp in milliseconds
// to make native datetime types and epoch dates work in annotation and table queries.
func ConvertSqlTimeColumnToEpochMs(values tsdb.RowValues, timeIndex int) {
	if timeIndex >= 0 {
		switch value := values[timeIndex].(type) {
		case time.Time:
			values[timeIndex] = float64(value.UnixNano()) / float64(time.Millisecond)
		case *time.Time:
			if value != nil {
				values[timeIndex] = float64(value.UnixNano()) / float64(time.Millisecond)
			}
		case int64:
			values[timeIndex] = int64(tsdb.EpochPrecisionToMs(float64(value)))
		case *int64:
			if value != nil {
				values[timeIndex] = int64(tsdb.EpochPrecisionToMs(float64(*value)))
			}
		case uint64:
			values[timeIndex] = int64(tsdb.EpochPrecisionToMs(float64(value)))
		case *uint64:
			if value != nil {
				values[timeIndex] = int64(tsdb.EpochPrecisionToMs(float64(*value)))
			}
		case int32:
			values[timeIndex] = int64(tsdb.EpochPrecisionToMs(float64(value)))
		case *int32:
			if value != nil {
				values[timeIndex] = int64(tsdb.EpochPrecisionToMs(float64(*value)))
			}
		case uint32:
			values[timeIndex] = int64(tsdb.EpochPrecisionToMs(float64(value)))
		case *uint32:
			if value != nil {
				values[timeIndex] = int64(tsdb.EpochPrecisionToMs(float64(*value)))
			}
		case float64:
			values[timeIndex] = tsdb.EpochPrecisionToMs(value)
		case *float64:
			if value != nil {
				values[timeIndex] = tsdb.EpochPrecisionToMs(*value)
			}
		case float32:
			values[timeIndex] = tsdb.EpochPrecisionToMs(float64(value))
		case *float32:
			if value != nil {
				values[timeIndex] = tsdb.EpochPrecisionToMs(float64(*value))
			}
		}
	}
}

// ConvertSqlValueColumnToFloat converts timeseries value column to float.
//nolint: gocyclo
func ConvertSqlValueColumnToFloat(columnName string, columnValue interface{}) (null.Float, error) {
	var value null.Float

	switch typedValue := columnValue.(type) {
	case int:
		value = null.FloatFrom(float64(typedValue))
	case *int:
		if typedValue == nil {
			value.Valid = false
		} else {
			value = null.FloatFrom(float64(*typedValue))
		}
	case int64:
		value = null.FloatFrom(float64(typedValue))
	case *int64:
		if typedValue == nil {
			value.Valid = false
		} else {
			value = null.FloatFrom(float64(*typedValue))
		}
	case int32:
		value = null.FloatFrom(float64(typedValue))
	case *int32:
		if typedValue == nil {
			value.Valid = false
		} else {
			value = null.FloatFrom(float64(*typedValue))
		}
	case int16:
		value = null.FloatFrom(float64(typedValue))
	case *int16:
		if typedValue == nil {
			value.Valid = false
		} else {
			value = null.FloatFrom(float64(*typedValue))
		}
	case int8:
		value = null.FloatFrom(float64(typedValue))
	case *int8:
		if typedValue == nil {
			value.Valid = false
		} else {
			value = null.FloatFrom(float64(*typedValue))
		}
	case uint:
		value = null.FloatFrom(float64(typedValue))
	case *uint:
		if typedValue == nil {
			value.Valid = false
		} else {
			value = null.FloatFrom(float64(*typedValue))
		}
	case uint64:
		value = null.FloatFrom(float64(typedValue))
	case *uint64:
		if typedValue == nil {
			value.Valid = false
		} else {
			value = null.FloatFrom(float64(*typedValue))
		}
	case uint32:
		value = null.FloatFrom(float64(typedValue))
	case *uint32:
		if typedValue == nil {
			value.Valid = false
		} else {
			value = null.FloatFrom(float64(*typedValue))
		}
	case uint16:
		value = null.FloatFrom(float64(typedValue))
	case *uint16:
		if typedValue == nil {
			value.Valid = false
		} else {
			value = null.FloatFrom(float64(*typedValue))
		}
	case uint8:
		value = null.FloatFrom(float64(typedValue))
	case *uint8:
		if typedValue == nil {
			value.Valid = false
		} else {
			value = null.FloatFrom(float64(*typedValue))
		}
	case float64:
		value = null.FloatFrom(typedValue)
	case *float64:
		value = null.FloatFromPtr(typedValue)
	case float32:
		value = null.FloatFrom(float64(typedValue))
	case *float32:
		if typedValue == nil {
			value.Valid = false
		} else {
			value = null.FloatFrom(float64(*typedValue))
		}
	case nil:
		value.Valid = false
	default:
		return null.NewFloat(0, false), fmt.Errorf("Value column must have numeric datatype, column: %s type: %T value: %v", columnName, typedValue, typedValue)
	}

	return value, nil
}

func SetupFillmode(query *tsdb.Query, interval time.Duration, fillmode string) error {
	query.Model.Set("fill", true)
	query.Model.Set("fillInterval", interval.Seconds())
	switch fillmode {
	case "NULL":
		query.Model.Set("fillMode", "null")
	case "previous":
		query.Model.Set("fillMode", "previous")
	default:
		query.Model.Set("fillMode", "value")
		floatVal, err := strconv.ParseFloat(fillmode, 64)
		if err != nil {
			return fmt.Errorf("error parsing fill value %v", fillmode)
		}
		query.Model.Set("fillValue", floatVal)
	}

	return nil
}

type SqlMacroEngineBase struct{}

func NewSqlMacroEngineBase() *SqlMacroEngineBase {
	return &SqlMacroEngineBase{}
}

func (m *SqlMacroEngineBase) ReplaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}
