package main

import (
	"flag"
	"io"
	"log"
	"os"
	"testing"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/features/steps"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

const componentLogFile = "component-output.txt"

var componentFlag = flag.Bool("component", false, "perform component tests")

type ComponentTest struct {
	MongoFeature *componenttest.MongoFeature
	t            testing.TB
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	component := steps.NewComponent(f.t)

	ctx.BeforeScenario(func(*godog.Scenario) {
		if err := component.Reset(); err != nil {
			log.Panicf("unable to initialise scenario: %s", err)
		}
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {
		component.Close()
	})

	component.RegisterSteps(ctx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {

}

func TestComponent(t *testing.T) {
	if *componentFlag {
		status := 0

		cfg, err := config.Get()
		if err != nil {
			t.Fatalf("failed to get service config: %s", err)
		}

		var output io.Writer = os.Stdout

		if cfg.ComponentTestUseLogFile {
			logfile, err := os.Create(componentLogFile)
			if err != nil {
				t.Fatalf("could not create logs file: %s", err)
			}

			defer func() {
				if err := logfile.Close(); err != nil {
					t.Fatalf("failed to close logs file: %s", err)
				}
			}()
			output = logfile
		}

		var opts = godog.Options{
			Output: colors.Colored(output),
			Format: "pretty",
			Paths:  flag.Args(),
		}

		f := &ComponentTest{t: t}

		status = godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip("component flag required to run component tests")
	}
}
