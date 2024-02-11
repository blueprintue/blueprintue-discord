package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/blueprintue/discord-bot/welcome"
)

var (
	ErrDiscordName  = errors.New("invalid json value: discord.name is empty")
	ErrDiscordToken = errors.New("invalid json value: discord.token is empty")
	ErrLogFilename  = errors.New("invalid json value: log.filename is empty")
	ErrLogLevel     = errors.New("invalid json value: log.level is invalid")
)

// Support only field's type string, int, bool, []string

// Configuration contains Discord, Log and Modules parameters.
type Configuration struct {
	Discord `json:"discord"`
	Log     `json:"log"`
	Modules `json:"modules"`
}

type Discord struct {
	Name  string `env:"DBOT_DISCORD_NAME"  json:"name"`
	Token string `env:"DBOT_DISCORD_TOKEN" json:"token"`
}

type Log struct {
	Filename string `env:"DBOT_LOG_FILENAME" json:"filename"`
	Level    string `env:"DBOT_LOG_LEVEL"    json:"level"`
}

type Modules struct {
	WelcomeConfiguration welcome.Configuration `json:"welcome"`
}

// ReadConfiguration read `config.json` file and update values with env if found.
func ReadConfiguration(fsys fs.FS, filename string) (*Configuration, error) {
	filedata, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	config := Configuration{}

	err = json.Unmarshal(filedata, &config)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	eraseConfigurationValuesWithEnv(&config)

	err = checkBasicConfiguration(config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

//nolint:cyclop
func eraseConfigurationValuesWithEnv(obj any) {
	var val reflect.Value
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		val = reflect.ValueOf(obj).Elem()
	} else {
		val = reflect.ValueOf(obj)
	}

	for idxNumField := 0; idxNumField < val.NumField(); idxNumField++ {
		if val.Field(idxNumField).Kind() == reflect.Struct {
			eraseConfigurationValuesWithEnv(val.Field(idxNumField).Addr().Interface())

			continue
		}

		envKey := reflect.TypeOf(obj).Elem().Field(idxNumField).Tag.Get("env")

		envValue, ok := os.LookupEnv(envKey)
		if !ok {
			continue
		}

		//nolint:exhaustive
		switch val.Field(idxNumField).Kind() {
		case reflect.String:
			val.Field(idxNumField).SetString(envValue)
		case reflect.Int:
			intEnvValue, err := strconv.Atoi(envValue)
			if err == nil {
				val.Field(idxNumField).SetInt(int64(intEnvValue))
			}
		case reflect.Bool:
			boolEnvValue, err := strconv.ParseBool(envValue)
			if err == nil {
				val.Field(idxNumField).SetBool(boolEnvValue)
			}
		case reflect.Slice:
			splitter := ":"
			if runtime.GOOS == "windows" {
				splitter = ";"
			}

			stringEnvValues := strings.Split(envValue, splitter)

			val.Field(idxNumField).Set(reflect.MakeSlice(val.Field(idxNumField).Type(), len(stringEnvValues), len(stringEnvValues)))

			for idxSlice := 0; idxSlice < len(stringEnvValues); idxSlice++ {
				val.Field(idxNumField).Index(idxSlice).SetString(stringEnvValues[idxSlice])
			}
		}
	}
}

func checkBasicConfiguration(config Configuration) error {
	if config.Discord.Name == "" {
		return ErrDiscordName
	}

	if config.Discord.Token == "" {
		return ErrDiscordToken
	}

	if config.Log.Filename == "" {
		return ErrLogFilename
	}

	isValidLogLevel := false

	validLogLevelValue := []string{"", "trace", "debug", "info", "warn", "error", "fatal", "panic"}
	for _, levelValue := range validLogLevelValue {
		if config.Log.Level == levelValue {
			isValidLogLevel = true
		}
	}

	if !isValidLogLevel {
		return ErrLogLevel
	}

	return nil
}
