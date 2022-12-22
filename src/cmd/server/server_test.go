package main

import (
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_realMain(t *testing.T) {
	logger := helpers.GetTestLogger()
	t.Run("invalid config: wrong port", func(t *testing.T) {
		conf, _ := config.LoadConfig()
		conf.Server.Port = "wrongPort"
		assert.Panics(t, func() {
			go func() {
				time.Sleep(1 * time.Second)
				stop <- true
				time.Sleep(1 * time.Second)
			}()
			realMain(conf, logger)
		}, "starting runner should panic")
	})
	t.Run("invalid config: wrong server type", func(t *testing.T) {
		conf, _ := config.LoadConfig()
		conf.Server.Type = "wrongtype"
		assert.Panics(t, func() {
			go func() {
				time.Sleep(1 * time.Second)
				stop <- true
				time.Sleep(1 * time.Second)
			}()
			realMain(conf, logger)
		}, "starting runner should panic")
	})
	t.Run("valid runner", func(t *testing.T) {
		conf, _ := config.LoadConfig()

		conf.Server.Port = "5000"
		assert.NotPanics(t, func() {
			go func() {
				time.Sleep(1 * time.Second)
				stop <- true
				time.Sleep(1 * time.Second)
			}()
			realMain(conf, logger)
		}, "starting runner shouldn't panic")
	})
}
