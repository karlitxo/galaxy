package registry

import (
	"strconv"

	"github.com/litl/galaxy/utils"
)

type ServiceConfig struct {
	// ID is used for ordering and conflict resolution.
	// Usualy set to time.Now().UnixNano()
	Name            string `redis:"name"`
	versionVMap     *utils.VersionedMap
	environmentVMap *utils.VersionedMap
	portsVMap       *utils.VersionedMap
}

func NewServiceConfig(app, version string) *ServiceConfig {
	svcCfg := &ServiceConfig{
		Name:            app,
		versionVMap:     utils.NewVersionedMap(),
		environmentVMap: utils.NewVersionedMap(),
		portsVMap:       utils.NewVersionedMap(),
	}
	svcCfg.SetVersion(version)

	return svcCfg
}

func NewServiceConfigWithEnv(app, version string, env map[string]string) *ServiceConfig {
	svcCfg := NewServiceConfig(app, version)

	for k, v := range env {
		svcCfg.environmentVMap.Set(k, v)
	}

	return svcCfg
}

// Env returns a map representing the runtime environment for the container.
// Changes to this map have no effect.
func (s *ServiceConfig) Env() map[string]string {
	env := map[string]string{}
	for _, k := range s.environmentVMap.Keys() {
		val := s.environmentVMap.Get(k)
		if val != "" {
			env[k] = val
		}
	}
	return env
}

func (s *ServiceConfig) EnvSet(key, value string) {
	if s.environmentVMap.Get(key) != value {
		s.environmentVMap.SetVersion(key, value, s.nextID())
	}
}

func (s *ServiceConfig) EnvGet(key string) string {
	return s.environmentVMap.Get(key)
}

func (s *ServiceConfig) Version() string {
	return s.versionVMap.Get("version")
}

func (s *ServiceConfig) SetVersion(version string) {
	if s.versionVMap.Get("version") != version {
		s.versionVMap.SetVersion("version", version, s.nextID())
	}
}

func (s *ServiceConfig) Ports() map[string]string {
	ports := map[string]string{}
	for _, k := range s.portsVMap.Keys() {
		val := s.portsVMap.Get(k)
		if val != "" {
			ports[k] = val
		}
	}
	return ports
}

func (s *ServiceConfig) ClearPorts() {
	for _, k := range s.portsVMap.Keys() {
		s.portsVMap.SetVersion(k, "", s.nextID())
	}
}

func (s *ServiceConfig) AddPort(port, portType string) {
	s.portsVMap.Set(port, portType)
}

func (s *ServiceConfig) ID() int64 {
	id := int64(0)
	for _, vmap := range []*utils.VersionedMap{
		s.environmentVMap,
		s.versionVMap,
		s.portsVMap,
	} {
		if vmap.LatestVersion() > id {
			id = vmap.LatestVersion()
		}
	}
	return id
}

func (s *ServiceConfig) ContainerName() string {
	return s.Name + "_" + strconv.FormatInt(s.ID(), 10)
}

func (s *ServiceConfig) nextID() int64 {
	return s.ID() + 1
}