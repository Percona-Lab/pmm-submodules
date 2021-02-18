package cloudwatch

import (
	"fmt"
	"sort"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/tsdb"
)

// returns a map of queries with query id as key. In the case a q request query
// has more than one statistic defined, one cloudwatchQuery will be created for each statistic.
// If the query doesn't have an Id defined by the user, we'll give it an with format `query[RefId]`. In the case
// the incoming query had more than one stat, it will ge an id like `query[RefId]_[StatName]`, eg queryC_Average
func (e *cloudWatchExecutor) transformRequestQueriesToCloudWatchQueries(requestQueries []*requestQuery) (
	map[string]*cloudWatchQuery, error) {
	plog.Debug("Transforming CloudWatch request queries")
	cloudwatchQueries := make(map[string]*cloudWatchQuery)
	for _, requestQuery := range requestQueries {
		for _, stat := range requestQuery.Statistics {
			id := requestQuery.Id
			if id == "" {
				id = fmt.Sprintf("query%s", requestQuery.RefId)
			}
			if len(requestQuery.Statistics) > 1 {
				id = fmt.Sprintf("%s_%v", id, strings.ReplaceAll(*stat, ".", "_"))
			}

			if _, ok := cloudwatchQueries[id]; ok {
				return nil, fmt.Errorf("error in query %q - query ID %q is not unique", requestQuery.RefId, id)
			}

			query := &cloudWatchQuery{
				Id:         id,
				RefId:      requestQuery.RefId,
				Region:     requestQuery.Region,
				Namespace:  requestQuery.Namespace,
				MetricName: requestQuery.MetricName,
				Dimensions: requestQuery.Dimensions,
				Stats:      *stat,
				Period:     requestQuery.Period,
				Alias:      requestQuery.Alias,
				Expression: requestQuery.Expression,
				ReturnData: requestQuery.ReturnData,
				MatchExact: requestQuery.MatchExact,
			}
			cloudwatchQueries[id] = query
		}
	}

	return cloudwatchQueries, nil
}

func (e *cloudWatchExecutor) transformQueryResponseToQueryResult(cloudwatchResponses []*cloudwatchResponse) map[string]*tsdb.QueryResult {
	responsesByRefID := make(map[string][]*cloudwatchResponse)
	refIDs := sort.StringSlice{}
	for _, res := range cloudwatchResponses {
		refIDs = append(refIDs, res.RefId)
		responsesByRefID[res.RefId] = append(responsesByRefID[res.RefId], res)
	}
	// Ensure stable results
	refIDs.Sort()

	results := make(map[string]*tsdb.QueryResult)
	for _, refID := range refIDs {
		responses := responsesByRefID[refID]
		queryResult := tsdb.NewQueryResult()
		queryResult.RefId = refID
		queryResult.Meta = simplejson.New()
		queryResult.Series = tsdb.TimeSeriesSlice{}
		frames := make(data.Frames, 0, len(responses))

		requestExceededMaxLimit := false
		partialData := false
		queryMeta := []struct {
			Expression, ID string
			Period         int
		}{}

		for _, response := range responses {
			frames = append(frames, response.DataFrames...)
			requestExceededMaxLimit = requestExceededMaxLimit || response.RequestExceededMaxLimit
			partialData = partialData || response.PartialData
			queryMeta = append(queryMeta, struct {
				Expression, ID string
				Period         int
			}{
				Expression: response.Expression,
				ID:         response.Id,
				Period:     response.Period,
			})
		}

		sort.Slice(frames, func(i, j int) bool {
			return frames[i].Name < frames[j].Name
		})

		if requestExceededMaxLimit {
			queryResult.ErrorString = "Cloudwatch GetMetricData error: Maximum number of allowed metrics exceeded. Your search may have been limited."
		}
		if partialData {
			queryResult.ErrorString = "Cloudwatch GetMetricData error: Too many datapoints requested - your search has been limited. Please try to reduce the time range"
		}

		queryResult.Dataframes = tsdb.NewDecodedDataFrames(frames)
		queryResult.Meta.Set("gmdMeta", queryMeta)
		results[refID] = queryResult
	}

	return results
}
