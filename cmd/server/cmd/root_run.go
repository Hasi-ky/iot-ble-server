package cmd

import (
	"context"
	"iot-ble-server/internal/api"
	"iot-ble-server/internal/bleudp"
	"iot-ble-server/internal/config"
	"iot-ble-server/internal/device"
	"iot-ble-server/internal/packets"
	"iot-ble-server/internal/storage"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tasks := []func() error{
		setLogLevel,
		setServices,
		setCharacteristics,
		setDescriptors,
		printStartMessage,
		setupStorage,
		startBleUdp,
		startHttpServer,
		startDevKeepAlive,
	}

	for _, t := range tasks {
		if err := t(); err != nil {
			log.Fatal(err)
		}
	}

	sigChan := make(chan os.Signal, 1)
	exitChan := make(chan struct{})
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.WithField("signal", <-sigChan).Info("signal received")
	go func() {
		log.Warning("stopping iot-ble-server")
		// todo: handle graceful shutdown?
		exitChan <- struct{}{}
	}()
	select {
	case <-exitChan:
	case s := <-sigChan:
		log.WithField("signal", s).Info("signal received, stopping immediately")
	}

	return nil
}

func setLogLevel() error {
	log.SetLevel(log.Level(uint8(config.C.General.LogLevel)))
	return nil
}

func setDescriptors() error {
	packets.SetDescriptors()
	return nil
}

func setServices() error {
	packets.SetServices()
	return nil
}

func setCharacteristics() error {
	packets.SetCharacteristics()
	return nil
}

func printStartMessage() error {
	log.Info("starting IOT BLE Server")
	return nil
}

func setupStorage() error {
	if err := storage.Setup(config.C); err != nil {
		return errors.Wrap(err, "setup storage error")
	}

	return nil
}

func startBleUdp() error {
	if err := bleudp.Start(); err != nil {
		return errors.Wrap(err, "start ble udp error")
	}
	return nil
}

func startHttpServer() error {
	if err := api.Start(); err != nil {
		return errors.Wrap(err, "start http server error")
	}
	return nil
}

func startDevKeepAlive() error {
	if err := device.KeepAlive(); err != nil {
		return errors.Wrap(err, "start devices keep alive error")
	}
	return nil
}
