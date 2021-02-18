package plugins

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/plugins/backendplugin"
	"github.com/grafana/grafana/pkg/setting"
)

var (
	PluginTypeApp        = "app"
	PluginTypeDatasource = "datasource"
	PluginTypePanel      = "panel"
	PluginTypeDashboard  = "dashboard"
)

type PluginState string

var (
	PluginStateAlpha PluginState = "alpha"
	PluginStateBeta  PluginState = "beta"
)

type PluginSignature string

const (
	PluginSignatureInternal PluginSignature = "internal" // core plugin, no signature
	PluginSignatureValid    PluginSignature = "valid"    // signed and accurate MANIFEST
	PluginSignatureInvalid  PluginSignature = "invalid"  // invalid signature
	PluginSignatureModified PluginSignature = "modified" // valid signature, but content mismatch
	PluginSignatureUnsigned PluginSignature = "unsigned" // no MANIFEST file
)

type PluginNotFoundError struct {
	PluginID string
}

func (e PluginNotFoundError) Error() string {
	return fmt.Sprintf("plugin with ID %q not found", e.PluginID)
}

type duplicatePluginError struct {
	Plugin         *PluginBase
	ExistingPlugin *PluginBase
}

func (e duplicatePluginError) Error() string {
	return fmt.Sprintf("plugin with ID %q already loaded from %q", e.Plugin.Id, e.ExistingPlugin.PluginDir)
}

func (e duplicatePluginError) Is(err error) bool {
	_, ok := err.(duplicatePluginError)
	return ok
}

// PluginLoader can load a plugin.
type PluginLoader interface {
	// Load loads a plugin and registers it with the manager.
	Load(decoder *json.Decoder, base *PluginBase, backendPluginManager backendplugin.Manager) error
}

// PluginBase is the base plugin type.
type PluginBase struct {
	Type         string             `json:"type"`
	Name         string             `json:"name"`
	Id           string             `json:"id"`
	Info         PluginInfo         `json:"info"`
	Dependencies PluginDependencies `json:"dependencies"`
	Includes     []*PluginInclude   `json:"includes"`
	Module       string             `json:"module"`
	BaseUrl      string             `json:"baseUrl"`
	Category     string             `json:"category"`
	HideFromList bool               `json:"hideFromList,omitempty"`
	Preload      bool               `json:"preload"`
	State        PluginState        `json:"state,omitempty"`
	Signature    PluginSignature    `json:"signature"`
	Backend      bool               `json:"backend"`

	IncludedInAppId string `json:"-"`
	PluginDir       string `json:"-"`
	DefaultNavUrl   string `json:"-"`
	IsCorePlugin    bool   `json:"-"`

	GrafanaNetVersion   string `json:"-"`
	GrafanaNetHasUpdate bool   `json:"-"`

	Root *PluginBase
}

func (pb *PluginBase) registerPlugin(base *PluginBase) error {
	if p, exists := Plugins[pb.Id]; exists {
		return duplicatePluginError{Plugin: pb, ExistingPlugin: p}
	}

	if !strings.HasPrefix(base.PluginDir, setting.StaticRootPath) {
		plog.Info("Registering plugin", "id", pb.Id)
	}

	if len(pb.Dependencies.Plugins) == 0 {
		pb.Dependencies.Plugins = []PluginDependencyItem{}
	}

	if pb.Dependencies.GrafanaVersion == "" {
		pb.Dependencies.GrafanaVersion = "*"
	}

	for _, include := range pb.Includes {
		if include.Role == "" {
			include.Role = models.ROLE_VIEWER
		}
	}

	// Copy relevant fields from the base
	pb.PluginDir = base.PluginDir
	pb.Signature = base.Signature

	Plugins[pb.Id] = pb
	return nil
}

type PluginDependencies struct {
	GrafanaVersion string                 `json:"grafanaVersion"`
	Plugins        []PluginDependencyItem `json:"plugins"`
}

type PluginInclude struct {
	Name       string          `json:"name"`
	Path       string          `json:"path"`
	Type       string          `json:"type"`
	Component  string          `json:"component"`
	Role       models.RoleType `json:"role"`
	AddToNav   bool            `json:"addToNav"`
	DefaultNav bool            `json:"defaultNav"`
	Slug       string          `json:"slug"`

	Id string `json:"-"`
}

type PluginDependencyItem struct {
	Type    string `json:"type"`
	Id      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type PluginBuildInfo struct {
	Time   int64  `json:"time,omitempty"`
	Repo   string `json:"repo,omitempty"`
	Branch string `json:"branch,omitempty"`
	Hash   string `json:"hash,omitempty"`
}

type PluginInfo struct {
	Author      PluginInfoLink      `json:"author"`
	Description string              `json:"description"`
	Links       []PluginInfoLink    `json:"links"`
	Logos       PluginLogos         `json:"logos"`
	Build       PluginBuildInfo     `json:"build"`
	Screenshots []PluginScreenshots `json:"screenshots"`
	Version     string              `json:"version"`
	Updated     string              `json:"updated"`
}

type PluginInfoLink struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type PluginLogos struct {
	Small string `json:"small"`
	Large string `json:"large"`
}

type PluginScreenshots struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type PluginStaticRoute struct {
	Directory string
	PluginId  string
}

type EnabledPlugins struct {
	Panels      []*PanelPlugin
	DataSources map[string]*DataSourcePlugin
	Apps        []*AppPlugin
}

func NewEnabledPlugins() EnabledPlugins {
	return EnabledPlugins{
		Panels:      make([]*PanelPlugin, 0),
		DataSources: make(map[string]*DataSourcePlugin),
		Apps:        make([]*AppPlugin, 0),
	}
}
