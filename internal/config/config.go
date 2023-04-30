package config

//go:generate python3 config-gen.py
//go:generate python3 config-env.py

import (
	"edm/internal/core"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Config struct defines server-side config of the app which is in cfg file
type Config struct {
	ServerSystem  string
	ServerRoot    string
	ServerHost    string
	ServerPort    string
	DomainName    string
	DefaultLang   string
	StartPage     string
	RemoveAllowed string
	RunBrowser    string
	UseTLS        string
	SSLCertFile   string
	SSLKeyFile    string
	CreateDB      string
	DBType        string
	DBName        string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	REDISConnect  string
	REDISPassword string
	REDISFlush    string
	SMTPEmail     string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPassword  string
}

// ReadConfig reads config into memory
func (cfg *Config) ReadConfig(configPath string, serverRoot string) error {
	const appDir = ".edm" // in this directory everything usually stored
	const defaultiniFname = "edm-system.cfg"
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "." // trying to write to the current dir if getting homedir failed
	}
	if configPath == "" {
		configPath = filepath.Join(homeDir, appDir, defaultiniFname) //default config path
	}
	if serverRoot == "" {
		cfg.ServerRoot = filepath.Join(homeDir, appDir) //default server root path
	} else {
		cfg.ServerRoot = serverRoot
	}
	ConfigFileFull := configPath
	fileInfo, _ := os.Stat(ConfigFileFull)
	isDir := false
	if fileInfo != nil {
		isDir = fileInfo.IsDir()
	}
	if strings.HasSuffix(configPath, "\\") || strings.HasSuffix(configPath, "/") || isDir {
		ConfigFileFull = filepath.Join(configPath, defaultiniFname)
	}
	if core.DEBUG {
		log.Println("Using config file:", ConfigFileFull)
	}
	path, _ := filepath.Split(ConfigFileFull)
	if path != "." && path != "" {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				err := os.Mkdir(path, 0700)
				if err != nil {
					return err
				}
			}
		}
	}
	if _, err := os.Stat(ConfigFileFull); err != nil {
		if os.IsNotExist(err) {
			f, err := os.Create(ConfigFileFull)
			if err != nil {
				return err
			}
			defer f.Close()

			// Writing config from a struct to file
			f.WriteString(makeStringToWriteToINI(cfg))
			f.Sync()
		}
	} else {
		mapOfConfig, err := readini(ConfigFileFull)
		if err != nil {
			return err
		}

		// Reading Config from a map to a struct
		readMapToCfgStruct(mapOfConfig, cfg)

	}
	cfg.readOSEnv()
	return nil
}

// WriteConfig writes config to disk
func (cfg *Config) WriteConfig(configPath string) error {
	const appDir = ".edm" // in this directory everything usually stored
	const defaultiniFname = "edm-system.cfg"
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "." // trying to write to the current dir if getting homedir failed
	}
	if configPath == "" {
		configPath = filepath.Join(homeDir, appDir) //default config path
		cfg.ServerRoot = configPath
	}
	ConfigFile := filepath.Join(configPath, defaultiniFname)
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(configPath, 0700)
			if err != nil {
				return err
			}
		}
	}

	f, err := os.Create(ConfigFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Writing config from a struct to file
	f.WriteString(makeStringToWriteToINI(cfg))

	f.Sync()

	return nil
}

func readini(fname string) (map[string]string, error) {
	mapConfig := make(map[string]string)
	content, err := ioutil.ReadFile(fname)
	if err != nil {
		return mapConfig, err
	}
	arrayLines := strings.Split(string(content), "\n")
	var configName, configVal string
	var validConfigLine = regexp.MustCompile("^[^#].+=.*")
	for _, element := range arrayLines {
		element = strings.TrimSpace(element)
		if validConfigLine.MatchString(element) {
			arr := strings.SplitN(element, "=", 2)
			configName = strings.TrimSpace(arr[0])
			configVal = strings.TrimSpace(arr[1])
			mapConfig[configName] = configVal
		}
	}
	return mapConfig, nil
}
