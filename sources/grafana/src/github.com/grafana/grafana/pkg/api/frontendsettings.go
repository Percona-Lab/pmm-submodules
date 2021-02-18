package api

import (
	"strconv"

	"github.com/grafana/grafana/pkg/models"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/util"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/setting"
)

func getFSDataSources(c *models.ReqContext, enabledPlugins *plugins.EnabledPlugins) (map[string]interface{}, error) {
	orgDataSources := make([]*models.DataSource, 0)

	if c.OrgId != 0 {
		query := models.GetDataSourcesQuery{OrgId: c.OrgId}
		err := bus.Dispatch(&query)

		if err != nil {
			return nil, err
		}

		dsFilterQuery := models.DatasourcesPermissionFilterQuery{
			User:        c.SignedInUser,
			Datasources: query.Result,
		}

		if err := bus.Dispatch(&dsFilterQuery); err != nil {
			if err != bus.ErrHandlerNotFound {
				return nil, err
			}

			orgDataSources = query.Result
		} else {
			orgDataSources = dsFilterQuery.Result
		}
	}

	dataSources := make(map[string]interface{})

	for _, ds := range orgDataSources {
		url := ds.Url

		if ds.Access == models.DS_ACCESS_PROXY {
			url = "/api/datasources/proxy/" + strconv.FormatInt(ds.Id, 10)
		}

		dsMap := map[string]interface{}{
			"id":        ds.Id,
			"uid":       ds.Uid,
			"type":      ds.Type,
			"name":      ds.Name,
			"url":       url,
			"isDefault": ds.IsDefault,
		}

		meta, exists := enabledPlugins.DataSources[ds.Type]
		if !exists {
			log.Errorf(3, "Could not find plugin definition for data source: %v", ds.Type)
			continue
		}
		dsMap["meta"] = meta

		jsonData := ds.JsonData
		if jsonData == nil {
			jsonData = simplejson.New()
		}

		dsMap["jsonData"] = jsonData

		if ds.Access == models.DS_ACCESS_DIRECT {
			if ds.BasicAuth {
				dsMap["basicAuth"] = util.GetBasicAuthHeader(ds.BasicAuthUser, ds.DecryptedBasicAuthPassword())
			}
			if ds.WithCredentials {
				dsMap["withCredentials"] = ds.WithCredentials
			}

			if ds.Type == models.DS_INFLUXDB_08 {
				dsMap["username"] = ds.User
				dsMap["password"] = ds.DecryptedPassword()
				dsMap["url"] = url + "/db/" + ds.Database
			}

			if ds.Type == models.DS_INFLUXDB {
				dsMap["username"] = ds.User
				dsMap["password"] = ds.DecryptedPassword()
				dsMap["url"] = url
			}
		}

		if (ds.Type == models.DS_INFLUXDB) || (ds.Type == models.DS_ES) {
			dsMap["database"] = ds.Database
		}

		if ds.Type == models.DS_PROMETHEUS {
			// add unproxied server URL for link to Prometheus web UI
			jsonData.Set("directUrl", ds.Url)
		}

		dataSources[ds.Name] = dsMap
	}

	// add data sources that are built in (meaning they are not added via data sources page, nor have any entry in
	// the datasource table)
	for _, ds := range plugins.DataSources {
		if ds.BuiltIn {
			dataSources[ds.Name] = map[string]interface{}{
				"type": ds.Type,
				"name": ds.Name,
				"meta": plugins.DataSources[ds.Id],
			}
		}
	}

	return dataSources, nil
}

// getFrontendSettingsMap returns a json object with all the settings needed for front end initialisation.
func (hs *HTTPServer) getFrontendSettingsMap(c *models.ReqContext) (map[string]interface{}, error) {
	enabledPlugins, err := plugins.GetEnabledPlugins(c.OrgId)
	if err != nil {
		return nil, err
	}
	pluginsToPreload := []string{}
	for _, app := range enabledPlugins.Apps {
		if app.Preload {
			pluginsToPreload = append(pluginsToPreload, app.Module)
		}
	}

	dataSources, err := getFSDataSources(c, enabledPlugins)
	if err != nil {
		return nil, err
	}

	defaultDS := "-- Grafana --"
	for n, ds := range dataSources {
		dsM := ds.(map[string]interface{})
		if isDefault, _ := dsM["isDefault"].(bool); isDefault {
			defaultDS = n
		}
		delete(dsM, "isDefault")

		meta := dsM["meta"].(*plugins.DataSourcePlugin)
		if meta.Preload {
			pluginsToPreload = append(pluginsToPreload, meta.Module)
		}
	}

	panels := map[string]interface{}{}
	for _, panel := range enabledPlugins.Panels {
		if panel.State == plugins.PluginStateAlpha && !hs.Cfg.PluginsEnableAlpha {
			continue
		}

		if panel.Preload {
			pluginsToPreload = append(pluginsToPreload, panel.Module)
		}

		panels[panel.Id] = map[string]interface{}{
			"module":        panel.Module,
			"baseUrl":       panel.BaseUrl,
			"name":          panel.Name,
			"id":            panel.Id,
			"info":          panel.Info,
			"hideFromList":  panel.HideFromList,
			"sort":          getPanelSort(panel.Id),
			"skipDataQuery": panel.SkipDataQuery,
			"state":         panel.State,
			"signature":     panel.Signature,
		}
	}

	hideVersion := hs.Cfg.AnonymousHideVersion && !c.IsSignedIn
	version := setting.BuildVersion
	commit := setting.BuildCommit
	buildstamp := setting.BuildStamp

	if hideVersion {
		version = ""
		commit = ""
		buildstamp = 0
	}

	jsonObj := map[string]interface{}{
		"defaultDatasource":          defaultDS,
		"datasources":                dataSources,
		"minRefreshInterval":         setting.MinRefreshInterval,
		"panels":                     panels,
		"appUrl":                     setting.AppUrl,
		"appSubUrl":                  setting.AppSubUrl,
		"allowOrgCreate":             (setting.AllowUserOrgCreate && c.IsSignedIn) || c.IsGrafanaAdmin,
		"authProxyEnabled":           setting.AuthProxyEnabled,
		"ldapEnabled":                setting.LDAPEnabled,
		"alertingEnabled":            setting.AlertingEnabled,
		"alertingErrorOrTimeout":     setting.AlertingErrorOrTimeout,
		"alertingNoDataOrNullValues": setting.AlertingNoDataOrNullValues,
		"alertingMinInterval":        setting.AlertingMinInterval,
		"autoAssignOrg":              setting.AutoAssignOrg,
		"verifyEmailEnabled":         setting.VerifyEmailEnabled,
		"sigV4AuthEnabled":           setting.SigV4AuthEnabled,
		"exploreEnabled":             setting.ExploreEnabled,
		"googleAnalyticsId":          setting.GoogleAnalyticsId,
		"disableLoginForm":           setting.DisableLoginForm,
		"disableUserSignUp":          !setting.AllowUserSignUp,
		"loginHint":                  setting.LoginHint,
		"passwordHint":               setting.PasswordHint,
		"externalUserMngInfo":        setting.ExternalUserMngInfo,
		"externalUserMngLinkUrl":     setting.ExternalUserMngLinkUrl,
		"externalUserMngLinkName":    setting.ExternalUserMngLinkName,
		"viewersCanEdit":             setting.ViewersCanEdit,
		"editorsCanAdmin":            hs.Cfg.EditorsCanAdmin,
		"disableSanitizeHtml":        hs.Cfg.DisableSanitizeHtml,
		"pluginsToPreload":           pluginsToPreload,
		"buildInfo": map[string]interface{}{
			"hideVersion":   hideVersion,
			"version":       version,
			"commit":        commit,
			"buildstamp":    buildstamp,
			"edition":       hs.License.Edition(),
			"latestVersion": plugins.GrafanaLatestVersion,
			"hasUpdate":     plugins.GrafanaHasUpdate,
			"env":           setting.Env,
			"isEnterprise":  hs.License.HasValidLicense(),
		},
		"licenseInfo": map[string]interface{}{
			"hasLicense":      hs.License.HasLicense(),
			"hasValidLicense": hs.License.HasValidLicense(),
			"expiry":          hs.License.Expiry(),
			"stateInfo":       hs.License.StateInfo(),
			"licenseUrl":      hs.License.LicenseURL(c.SignedInUser),
			"edition":         hs.License.Edition(),
		},
		"featureToggles":    hs.Cfg.FeatureToggles,
		"rendererAvailable": hs.RenderService.IsAvailable(),
		"http2Enabled":      hs.Cfg.Protocol == setting.HTTP2Scheme,
		"marketplaceUrl":    hs.Cfg.MarketplaceURL,
	}

	return jsonObj, nil
}

func getPanelSort(id string) int {
	sort := 100
	switch id {
	case "graph":
		sort = 1
	case "stat":
		sort = 2
	case "gauge":
		sort = 3
	case "bargauge":
		sort = 4
	case "table":
		sort = 5
	case "singlestat":
		sort = 6
	case "text":
		sort = 7
	case "heatmap":
		sort = 8
	case "alertlist":
		sort = 9
	case "dashlist":
		sort = 10
	case "news":
		sort = 10
	}
	return sort
}

func (hs *HTTPServer) GetFrontendSettings(c *models.ReqContext) {
	settings, err := hs.getFrontendSettingsMap(c)
	if err != nil {
		c.JsonApiErr(400, "Failed to get frontend settings", err)
		return
	}

	c.JSON(200, settings)
}
