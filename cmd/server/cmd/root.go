package cmd

import (
	"bytes"
	"io/ioutil"
	"iot-ble-server/internal/config"
	"os"
	"reflect"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "iot-ble-server",
	Short: "iot-ble-server",
	RunE:  run,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to configuration file (optional)")
	rootCmd.PersistentFlags().Int("log-level", 4, "debug=5, info=4, error=2, fatal=1, panic=0")
	// bind flag to config vars
	viper.BindPFlag("general.log_level", rootCmd.PersistentFlags().Lookup("log-level"))

	// defaults 
	rootCmd.AddCommand(configCmd)
}

 
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	cfgFile = "D:\\Code\\iot-ble-server\\packaging\\files\\iot-ble-server.toml"
	if cfgFile != "" {
		b, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
		viper.SetConfigType("toml")
		if err := viper.ReadConfig(bytes.NewBuffer(b)); err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
	} else {
		viper.SetConfigName("iot-ble-server")
		viper.AddConfigPath(".")
		viper.AddConfigPath("iot-ble-server/packaging/files/iot-ble-server.toml")
		viper.AddConfigPath("/etc/iot-ble-server")
		if err := viper.ReadInConfig(); err != nil {
			switch err.(type) {
			case viper.ConfigFileNotFoundError:
				log.Warnln("No configuration file found, using defaults.")
			default:
				log.WithError(err).Fatalln("read configuration file error")
			}
		}
	}

	for _, pair := range os.Environ() {
		d := strings.SplitN(pair, "=", 2)
		if strings.Contains(d[0], ".") {
			log.Warnln("Using dots in env variable is illegal and deprecated. Please use double underscore `__` for: ", d[0])
			underscoreName := strings.ReplaceAll(d[0], ".", "__")
			// Set only when the underscore version doesn't already exist.
			if _, exists := os.LookupEnv(underscoreName); !exists {
				os.Setenv(underscoreName, d[1])
			}
		}
	}

	viperBindEnvs(config.C)

	if err := viper.Unmarshal(&config.C); err != nil {
		log.WithError(err).Fatal("unmarshal config error")
	}

	if config.C.Redis.URL != "" {
		opt, err := redis.ParseURL(config.C.Redis.URL)
		if err != nil {
			log.WithError(err).Fatal("redis url error")
		}

		config.C.Redis.Servers = []string{opt.Addr}
		config.C.Redis.Database = opt.DB
		config.C.Redis.Password = opt.Password
	}
}

func viperBindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			tv = strings.ToLower(t.Name)
		}
		if tv == "-" {
			continue
		}

		switch v.Kind() {
		case reflect.Struct:
			viperBindEnvs(v.Interface(), append(parts, tv)...)
		default:
			// Bash doesn't allow env variable names with a dot so
			// bind the double underscore version.
			keyDot := strings.Join(append(parts, tv), ".")
			keyUnderscore := strings.Join(append(parts, tv), "__")
			viper.BindEnv(keyDot, strings.ToUpper(keyUnderscore))
		}
	}
}
