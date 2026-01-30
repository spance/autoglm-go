package constants

import (
	_ "embed"
	"encoding/json"
	"errors"
	"sync"
)

//go:embed app_aliases.json
var aliasesJSON []byte

var (
	pkg2AliasesMap map[string][]string
	alias2PkgMap   map[string]string
	errLoad        error
	once           = new(sync.Once)
)

// Load loads the app mapping from the embedded JSON
func Load() (map[string][]string, error) {
	once.Do(func() {
		pkg2AliasesMap = make(map[string][]string)
		if err := json.Unmarshal(aliasesJSON, &pkg2AliasesMap); err != nil {
			errLoad = errors.Join(err, errors.New("failed to unmarshal embedded app_aliases.json"))
			return
		}

		alias2PkgMap = make(map[string]string)
		for pkg, aliases := range pkg2AliasesMap {
			for _, alias := range aliases {
				alias2PkgMap[alias] = pkg
			}
		}
	})
	return pkg2AliasesMap, errLoad
}

// GetPackageByAlias returns the package name for a given alias
func GetPackageByAlias(alias string) (string, bool) {
	_, err := Load()
	if err != nil {
		return "", false
	}
	pkg, ok := alias2PkgMap[alias]
	return pkg, ok
}

// GetAliasesByPackage returns the aliases for a given package name
func GetAliasesByPackage(pkg string) ([]string, bool) {
	_, err := Load()
	if err != nil {
		return nil, false
	}
	aliases, ok := pkg2AliasesMap[pkg]
	return aliases, ok
}

// GetAliasByPackage returns the first alias for a given package name
func GetAliasByPackage(pkg string) (string, bool) {
	_, err := Load()
	if err != nil {
		return "", false
	}
	aliases, ok := pkg2AliasesMap[pkg]
	if !ok || len(aliases) == 0 {
		return "", false
	}
	return aliases[0], true
}
