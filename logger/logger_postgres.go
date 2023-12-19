package logger

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type pgHook struct {
	config pgCfg
	db     *gorm.DB
	level  logrus.Level
}

type pgCfg struct {
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	DBName    string `yaml:"dbName"`
	TableName string `yaml:"tableName"`
	Sslmode   string `yaml:"sslmode"`
}

// PGLog postgresql log
type PGLog struct {
	ID      uint      `json:",omitempty" gorm:"primary_key"`                  // 关系型数据库主键，内部使用
	Time    time.Time `json:"time" bson:"time" example:"2006-01-02 15:04:05"` // 日志时间 string类型
	Service string    `json:"service" bson:"service" example:"webserver"`     // 微服务
	Level   string    `json:"level" bson:"level" example:"error"`             // 日志级别，全小写
	Message string    `json:"message" bson:"message"`                         // 日志内容
	RawTime string    `json:",omitempty" bson:"raw_time"`                     // 日志时间，简单格式
}

var configPath = "/etc/config/db/postgres.yml"

// NewPGLogger create a new postgreSQL logger
func NewPGLogger(c Config, fields ...string) (Logger, error) {
	pg := &pgHook{}
	// 默认配置
	pg.config = pg.defaultCfg()
	// 若存在配置文件，根据配置文件配置PG
	if fi, err := os.Stat(configPath); err == nil && !fi.IsDir() {
		cfg := pgCfg{}
		if data, err := ioutil.ReadFile(configPath); err == nil {
			if err := yaml.Unmarshal(data, &cfg); err == nil {
				pg.config = cfg
			}
		}
	}
	// 设置日志级别
	level, err := logrus.ParseLevel(c.Level)
	if err != nil {
		log.Printf("failed to parse log level (%s), use default level (info)\n", c.Level)
		level = logrus.InfoLevel
	}
	pg.level = level

	logrusFields := logrus.Fields{}
	for index := 0; index < len(fields)-1; index = index + 2 {
		logrusFields[fields[index]] = fields[index+1]
	}

	// 连接postgreSQL
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		pg.config.Host, pg.config.Port, pg.config.User, pg.config.DBName, pg.config.Password, pg.config.Sslmode)
	db, err := gorm.Open("postgres", dbInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgre： %v", err)
	}
	// 创建表
	if !db.HasTable(pg.config.TableName) {
		pg.db = db.Table(pg.config.TableName).CreateTable(&PGLog{})
	} else {
		pg.db = db.Table(pg.config.TableName)
	}

	// 初始化entry
	entry := logrus.NewEntry(logrus.New())
	entry.Level = level
	entry.Logger.Level = level
	entry.Logger.Formatter = newFormatter(c.Format)
	entry.Logger.Hooks.Add(pg)

	return &logger{entry.WithFields(logrusFields)}, nil
}

func (pg *pgHook) defaultCfg() pgCfg {
	return pgCfg{
		Host:      "h3c-postgres",
		Port:      "5432",
		User:      "h3c_postgres",
		Password:  "h3c_iot_edge_postgres",
		DBName:    "h3c_log",
		TableName: "system",
		Sslmode:   "disable",
	}
}

func (pg *pgHook) Fire(entry *logrus.Entry) error {
	if pg.level < entry.Level {
		return nil
	}

	fields := make(logrus.Fields)
	for k, v := range entry.Data {
		if errInfo, isErr := v.(error); k == logrus.ErrorKey && v != nil && isErr {
			fields[k] = errInfo.Error()
		} else {
			fields[k] = v
		}
	}

	log := PGLog{
		Time:    entry.Time,
		RawTime: entry.Time.Format("2006-01-02 15:04:05"),
		Level:   entry.Level.String(),
		Message: entry.Message,
	}
	if service, ok := fields["service"]; ok {
		log.Service = service.(string)
	}

	pg.db.Create(&log)

	return nil
}

func (pg *pgHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (pg *pgHook) Close() error {
	return pg.db.Close()
}
