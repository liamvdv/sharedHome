package config

import (
	"path/filepath"
	"os"
	"github.com/liamvdv/sharedHome/osx"
)

var (
	LoadBackendConfig      loadFunc
	LoadBackendCredentials loadFunc
	LoadBackendToken       loadFunc

	StoreBackendConfig      storeFunc
	StoreBackendCredentials storeFunc
	StoreBackendToken       storeFunc
)

type loadFunc func(string) ([]byte, error)
type storeFunc func(string, []byte) error

// we need this setup to inject a fs.
func initBackendFuncs(fs osx.Fs) {
	mkLoadFunc := func(template string) loadFunc {
		return func(backend string) ([]byte, error) {
			fp := filepath.Join(template, backend)
			return fs.ReadFile(fp)
		}
	}
	mkStoreFunc := func(template string, perm os.FileMode) storeFunc {
		return func(backend string, raw []byte) error {
			fp := filepath.Join(template, backend)
			return fs.WriteFile(fp, raw, perm)
		}
	}
	LoadBackendConfig = mkLoadFunc("%s-configuration.json")
	StoreBackendConfig = mkStoreFunc("%s-configuration.json", 0600)

	LoadBackendCredentials = mkLoadFunc("%s-credentials.json")
	StoreBackendCredentials = mkStoreFunc("%s-credentials.json", 0600)

	LoadBackendToken = mkLoadFunc("%s-token.json")
	StoreBackendToken = mkStoreFunc("%s-token.json", 0600)
}
