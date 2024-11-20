package main

import (
	"log"
	"os"

	"go.temporal.io/sdk/worker"

	"github.com/milovidov983/oms-temporal-demo/workers/activities"
	"github.com/milovidov983/oms-temporal-demo/workers/queue"
	"github.com/milovidov983/oms-temporal-demo/workers/temporal"
	"github.com/milovidov983/oms-temporal-demo/workers/workflows"
	"github.com/spf13/viper"
)

func main() {
	log.SetPrefix("[temporal-worker]:")

	loadConfig()

	temporalConfig := temporal.TemporalClientConfig{
		HostPort:  viper.GetString("temporal.hostPort"),
		Namespace: viper.GetString("temporal.namespace"),
	}
	temporalConfig.Check()
	c, err := temporal.NewClient(temporalConfig)

	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	activitiesConfig := &activities.ActivitiesConfig{
		OmsCoreHostPort: viper.GetString("services.omsCore.hostPort"),
	}
	activitiesConfig.Check()

	a := activities.NewActivities(activitiesConfig)

	w := worker.New(c, queue.TaskQueueNameOrder, worker.Options{})

	w.RegisterWorkflow(workflows.ProcessOrder)

	w.RegisterActivity(a)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("worker exited: %v", err)
	}
}

func loadConfig() {
	env := os.Getenv("APP_ENV")
	configName := "config.dev.yml"
	if env == "production" {
		configName = "config.prod.yml"
	}

	log.Printf("[info] Loading configuration for environment: %s", env)

	viper.SetConfigName(configName)
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("[fatal] Config file not found: %v", err)
		} else {
			log.Fatalf("[fatal] Error reading config file: %v", err)
		}
	}

	viper.SetEnvPrefix("oms")
	viper.AutomaticEnv()

	log.Printf("[info] Configuration loaded successfully from: %s", viper.ConfigFileUsed())
}
