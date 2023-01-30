package core

import (
	"edm/pkg/accs"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

// GetThemeList loads a list of available frontend themes
func GetThemeList(serverSystem string) []string {
	files, err := ioutil.ReadDir(filepath.Join(serverSystem, "static", "themes"))
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		return []string{}
	}
	var res []string
	var fname string
	for _, file := range files {
		fname = file.Name()
		if ext := filepath.Ext(fname); ext == ".css" && !strings.HasPrefix(fname, "system-") {
			fname = fname[0 : len(fname)-len(ext)]
			res = append(res, fname)
		}
	}
	return res
}
