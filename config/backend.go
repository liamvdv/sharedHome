package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	backendConfigurationFileTemplate = "%s-configuration.json"
	backendCredentialsFileTemplate   = "%s-credentials.json"
	backendTokenFileTemplate         = "%s-token.json"
)

// BackendConfigFile returns the ConfigFile for that backend. It has read and write permissions and must be closed by consumer.
func BackendConfigFile(backend string) (*os.File, error) {
	fp := filepath.Join(BackendConfigFolder, fmt.Sprintf(backendConfigurationFileTemplate, backend))
	return os.OpenFile(fp, os.O_CREATE|os.O_RDWR, 0600)
}

// BackendCredentialsFile returns the ConfigFile for that backend. It has read and write permissions and must be closed by consumer.
func BackendCredentialsFile(backend string) (*os.File, error) {
	fp := filepath.Join(BackendConfigFolder, fmt.Sprintf(backendCredentialsFileTemplate, backend))
	return os.OpenFile(fp, os.O_CREATE|os.O_RDWR, 0600)
}

// BackendTokenFile returns the ConfigFile for that backend. It has read and write permissions and must be closed by consumer.
func BackendTokenFile(backend string) (*os.File, error) {
	fp := filepath.Join(BackendConfigFolder, fmt.Sprintf(backendTokenFileTemplate, backend))
	return os.OpenFile(fp, os.O_CREATE|os.O_RDWR, 0600)
}
