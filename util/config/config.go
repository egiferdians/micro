package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/egiferdians/micro/util/config/configetcd"
	"github.com/egiferdians/micro/util/config/configzk"
	"github.com/egiferdians/micro/util/flags"
	"github.com/egiferdians/micro/util/microservice"
)

type StdConfig struct {
	ServiceName string
	ServiceRoot string
	ConfigData map[string]string
	ConfigHosts []string
	EventPath   string
	eventHook   func()
}

const (
	Database = "database"
	DBHost   = "dbhost"
	DBport   = "dbport"
	DBname   = "dbname"
	DBuid    = "dbuid"
	DBpwd    = "dbpwd"
)

var AppConfig StdConfig
var localConfig = false
var configFile = "service.conf"

func Get(key string, defval string) string {

	v, ok := AppConfig.ConfigData[key]
	if !ok {
		if defval != "" {
			AppConfig.ConfigData[key] = defval
			v = defval
		}
	}
	return v
}

func GetA(key string, defval string) []string {

	v := Get(key, defval)

	if len(v) == 0 {
		return nil
	}

	sep := ","

	if len(defval) == 1 {
		sep = defval
	}

	return strings.Split(v, sep)
}

func GetI(key string, defval int) int {

	v := Get(key, strconv.Itoa(defval))

	if res, err := strconv.Atoi(v); err == nil {
		return res
	} else {
		return -1
	}
}

func GetB(key string, defval bool) bool {

	v := Get(key, strconv.FormatBool(defval))

	if res, err := strconv.ParseBool(v); err == nil {
		return res
	} else {
		return false
	}
}

func (sc *StdConfig) ConfigPath() string {
	return fmt.Sprintf("%s/%s", sc.ServiceRoot, sc.ServiceName)
}

func (sc *StdConfig) open() error {
	configFile, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)

	err = jsonParser.Decode(&sc)

	return err
}

func (sc *StdConfig) LoadConfig() bool {

	s, ok := os.LookupEnv(strings.ToUpper("micro_servicename"))
	if ok {
		sc.ServiceName = s
	} else {
		err := sc.open()
		if err != nil {
			log.Printf("error reading service.conf file. %+v\n", err)
			return false
		}
		if sc.ServiceName == "" {
			return false
		}
	}

	s, ok = os.LookupEnv(strings.ToUpper("micro_confighosts"))
	if ok {
		sc.ConfigHosts = strings.Split(s, ",")
	} else {
		sc.open()
	}

	if !localConfig && len(sc.ConfigHosts) > 0 {
		hosts := sc.ConfigHosts
		svcNode := sc.ConfigPath()

		var err error

		var CfgData map[string]string
		switch microservice.GetOsEnv(flags.MICRO_DISCOVERY_ENV_NAME) {
		case flags.MICRO_DISCOVERY_MODE_ZK:
			CfgData, err = configzk.ZKConnectAndListen(hosts, svcNode, sc.onZKChangeEvent)
			if err != nil {
				log.Printf("zk error %v\n", err)
				return false
			}

			if sc.ConfigData == nil {
				sc.ConfigData = make(configzk.ConfigFormat)
			}

			for k, v := range CfgData {
				sc.ConfigData[k] = v
			}

		default:
			CfgData, err = configetcd.ETCDConnectAndListen(hosts, svcNode, sc.onETCDChangeEvent)
			if err != nil {
				log.Printf("etcd  error %v\n", err)
				return false
			}

			if sc.ConfigData == nil {
				sc.ConfigData = make(configetcd.ConfigFormat)
			}

			for k, v := range CfgData {
				sc.ConfigData[k] = v
			}
		}

	}

	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		key := strings.ToLower(pair[0])

		if strings.HasPrefix(key, "micro_") {
			key = strings.Trim(key, "micro_")

			sc.ConfigData[key] = pair[1]
		}
	}
	return true
}
func (sc *StdConfig) onZKChangeEvent(nodename string, dataMap configzk.ConfigFormat) {
	sc.ConfigData = dataMap
	sc.EventPath = nodename
	if sc.eventHook != nil {
		sc.eventHook()
	}
}

func (sc *StdConfig) onETCDChangeEvent(nodename string, dataMap configetcd.ConfigFormat) {
	sc.ConfigData = dataMap
	sc.EventPath = nodename
	if sc.eventHook != nil {
		sc.eventHook()
	}
}