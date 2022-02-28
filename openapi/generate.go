package openapi

import (
	"os"
	"path"
)

func GenerateNamedActions(pathToPlugin string, openapi string, mask string) error {
	fullPath := getFullPath(pathToPlugin)

	actions, err := GetNamedActionsFromOpenapi(
		path.Join(fullPath, mask),
		path.Join(fullPath, openapi),
	)
	if err != nil {
		return err
	}

	err = os.WriteFile(fullPath+"/named_actions.yaml", actions, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getFullPath(p string) string {
	if path.IsAbs(p) {
		return p
	}
	currDir, err := os.Getwd()
	if err != nil {
		return p
	}
	return path.Join(currDir, p)
}
