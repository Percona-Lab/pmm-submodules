package graphite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"golang.org/x/net/context/ctxhttp"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/tsdb"
	"github.com/opentracing/opentracing-go"
)

type GraphiteExecutor struct {
	HttpClient *http.Client
}

func NewGraphiteExecutor(datasource *models.DataSource) (tsdb.TsdbQueryEndpoint, error) {
	return &GraphiteExecutor{}, nil
}

var glog = log.New("tsdb.graphite")

func init() {
	tsdb.RegisterTsdbQueryEndpoint("graphite", NewGraphiteExecutor)
}

func (e *GraphiteExecutor) Query(ctx context.Context, dsInfo *models.DataSource, tsdbQuery *tsdb.TsdbQuery) (*tsdb.Response, error) {
	result := &tsdb.Response{}

	from := "-" + formatTimeRange(tsdbQuery.TimeRange.From)
	until := formatTimeRange(tsdbQuery.TimeRange.To)
	var target string

	formData := url.Values{
		"from":          []string{from},
		"until":         []string{until},
		"format":        []string{"json"},
		"maxDataPoints": []string{"500"},
	}

	emptyQueries := make([]string, 0)
	for _, query := range tsdbQuery.Queries {
		glog.Debug("graphite", "query", query.Model)
		currTarget := ""
		if fullTarget, err := query.Model.Get("targetFull").String(); err == nil {
			currTarget = fullTarget
		} else {
			currTarget = query.Model.Get("target").MustString()
		}
		if currTarget == "" {
			glog.Debug("graphite", "empty query target", query.Model)
			emptyQueries = append(emptyQueries, fmt.Sprintf("Query: %v has no target", query.Model))
			continue
		}
		target = fixIntervalFormat(currTarget)
	}

	if target == "" {
		glog.Error("No targets in query model", "models without targets", strings.Join(emptyQueries, "\n"))
		return nil, errors.New("No query target found for the alert rule")
	}

	formData["target"] = []string{target}

	if setting.Env == setting.Dev {
		glog.Debug("Graphite request", "params", formData)
	}

	req, err := e.createRequest(dsInfo, formData)
	if err != nil {
		return nil, err
	}

	httpClient, err := dsInfo.GetHttpClient()
	if err != nil {
		return nil, err
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "graphite query")
	span.SetTag("target", target)
	span.SetTag("from", from)
	span.SetTag("until", until)
	span.SetTag("datasource_id", dsInfo.Id)
	span.SetTag("org_id", dsInfo.OrgId)

	defer span.Finish()

	if err := opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header)); err != nil {
		return nil, err
	}

	res, err := ctxhttp.Do(ctx, httpClient, req)
	if err != nil {
		return nil, err
	}

	data, err := e.parseResponse(res)
	if err != nil {
		return nil, err
	}

	result.Results = make(map[string]*tsdb.QueryResult)
	queryRes := tsdb.NewQueryResult()

	for _, series := range data {
		queryRes.Series = append(queryRes.Series, &tsdb.TimeSeries{
			Name:   series.Target,
			Points: series.DataPoints,
		})

		if setting.Env == setting.Dev {
			glog.Debug("Graphite response", "target", series.Target, "datapoints", len(series.DataPoints))
		}
	}

	result.Results["A"] = queryRes
	return result, nil
}

func (e *GraphiteExecutor) parseResponse(res *http.Response) ([]TargetResponseDTO, error) {
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	if res.StatusCode/100 != 2 {
		glog.Info("Request failed", "status", res.Status, "body", string(body))
		return nil, fmt.Errorf("Request failed status: %v", res.Status)
	}

	var data []TargetResponseDTO
	err = json.Unmarshal(body, &data)
	if err != nil {
		glog.Info("Failed to unmarshal graphite response", "error", err, "status", res.Status, "body", string(body))
		return nil, err
	}

	return data, nil
}

func (e *GraphiteExecutor) createRequest(dsInfo *models.DataSource, data url.Values) (*http.Request, error) {
	u, err := url.Parse(dsInfo.Url)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "render")

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(data.Encode()))
	if err != nil {
		glog.Info("Failed to create request", "error", err)
		return nil, fmt.Errorf("Failed to create request. error: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if dsInfo.BasicAuth {
		req.SetBasicAuth(dsInfo.BasicAuthUser, dsInfo.DecryptedBasicAuthPassword())
	}

	return req, err
}

func formatTimeRange(input string) string {
	if input == "now" {
		return input
	}
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(input, "now", ""), "m", "min"), "M", "mon")
}

func fixIntervalFormat(target string) string {
	rMinute := regexp.MustCompile(`'(\d+)m'`)
	target = rMinute.ReplaceAllStringFunc(target, func(m string) string {
		return strings.ReplaceAll(m, "m", "min")
	})
	rMonth := regexp.MustCompile(`'(\d+)M'`)
	target = rMonth.ReplaceAllStringFunc(target, func(M string) string {
		return strings.ReplaceAll(M, "M", "mon")
	})
	return target
}
