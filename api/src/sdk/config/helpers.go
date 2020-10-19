package config

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// FatalIfMissing log with a Fatal error if a field is missing from viper configuration
func FatalIfMissing(moduleName string, fields []string) {
	for _, field := range fields {
		if !viper.IsSet(field) {
			log.Fatal().Msgf("\n~ \033[0;31mMissing configuration field \033[1;91m%s\033[0;31m for the %s module.\033[0m\n", field, moduleName)
		}
	}
}

// Print configuration
func Print(moduleName string, fieldsToHide []string) {
	configMsg := fmt.Sprintf("~ Configuration of the \033[1;35m%s\033[0m module \\o/ \n", moduleName)
	keys := viper.AllKeys()
	sort.Strings(sort.StringSlice(keys))
	for _, key := range keys {
		value := viper.Get(key)
		// replace value by *** if part of secrets
		for _, toHideKey := range fieldsToHide {
			if toHideKey == key {
				value = "*********"
			}
		}
		splitKey := strings.Split(key, ".")
		if len(splitKey) == 2 {
			key = fmt.Sprintf("\033[0;1m%s\033[0m.\033[0;1m%s\033[0m", splitKey[0], splitKey[1])
		}
		configMsg = fmt.Sprintf("%s\t %s = \033[0;94m%v\033[0m\n", configMsg, key, value)
	}
	log.Info().Msgf("\n%s", configMsg)
}
