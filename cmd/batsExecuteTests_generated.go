// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type batsExecuteTestsOptions struct {
	OutputFormat string   `json:"outputFormat,omitempty" validate:"oneof=tap junit"`
	Repository   string   `json:"repository,omitempty"`
	TestPackage  string   `json:"testPackage,omitempty"`
	TestPath     string   `json:"testPath,omitempty"`
	EnvVars      []string `json:"envVars,omitempty"`
}

type batsExecuteTestsInflux struct {
	step_data struct {
		fields struct {
			bats bool
		}
		tags struct {
		}
	}
}

func (i *batsExecuteTestsInflux) persist(path, resourceName string) {
	measurementContent := []struct {
		measurement string
		valType     string
		name        string
		value       interface{}
	}{
		{valType: config.InfluxField, measurement: "step_data", name: "bats", value: i.step_data.fields.bats},
	}

	errCount := 0
	for _, metric := range measurementContent {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(metric.measurement, fmt.Sprintf("%vs", metric.valType), metric.name), metric.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting influx environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Fatal("failed to persist Influx environment")
	}
}

// BatsExecuteTestsCommand This step executes tests using the [Bash Automated Testing System - bats-core](https://github.com/bats-core/bats-core).
func BatsExecuteTestsCommand() *cobra.Command {
	const STEP_NAME = "batsExecuteTests"

	metadata := batsExecuteTestsMetadata()
	var stepConfig batsExecuteTestsOptions
	var startTime time.Time
	var influx batsExecuteTestsInflux
	var logCollector *log.CollectorHook
	splunkClient := &splunk.Splunk{}

	var createBatsExecuteTestsCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step executes tests using the [Bash Automated Testing System - bats-core](https://github.com/bats-core/bats-core).",
		Long:  `Bats is a TAP-compliant testing framework for Bash. It provides a simple way to verify that the UNIX programs you write behave as expected. A Bats test file is a Bash script with special syntax for defining test cases. Under the hood, each test case is just a function with a description.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				influx.persist(GeneralConfig.EnvRootPath, "influx")
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Send(&telemetryData, logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			batsExecuteTests(stepConfig, &telemetryData, &influx)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addBatsExecuteTestsFlags(createBatsExecuteTestsCmd, &stepConfig)
	return createBatsExecuteTestsCmd
}

func addBatsExecuteTestsFlags(cmd *cobra.Command, stepConfig *batsExecuteTestsOptions) {
	cmd.Flags().StringVar(&stepConfig.OutputFormat, "outputFormat", `junit`, "Defines the format of the test result output. junit would be the standard for automated build environments but you could use also the option tap.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", `https://github.com/bats-core/bats-core.git`, "Defines the version of bats-core to be used. By default we use the version from the master branch.")
	cmd.Flags().StringVar(&stepConfig.TestPackage, "testPackage", `piper-bats`, "For the transformation of the test result to xUnit format the node module tap-xunit is used. This parameter defines the name of the test package used in the xUnit result file.")
	cmd.Flags().StringVar(&stepConfig.TestPath, "testPath", `src/test`, "Defines either the directory which contains the test files (*.bats) or a single file. You can find further details in the Bats-core documentation.")
	cmd.Flags().StringSliceVar(&stepConfig.EnvVars, "envVars", []string{}, "Injects environment variables to step execution. Format of value must be ['<KEY1>=<VALUE1>','<KEY2>=<VALUE2>']. Example: ['CONTAINER_NAME=piper-jenskins','IMAGE_NAME=my-image']")

}

// retrieve step metadata
func batsExecuteTestsMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "batsExecuteTests",
			Aliases:     []config.Alias{},
			Description: "This step executes tests using the [Bash Automated Testing System - bats-core](https://github.com/bats-core/bats-core).",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Resources: []config.StepResources{
					{Name: "tests", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "outputFormat",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `junit`,
					},
					{
						Name:        "repository",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `https://github.com/bats-core/bats-core.git`,
					},
					{
						Name:        "testPackage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `piper-bats`,
					},
					{
						Name:        "testPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `src/test`,
					},
					{
						Name:        "envVars",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
				},
			},
			Containers: []config.Container{
				{Name: "bats", Image: "node:lts-stretch", WorkingDir: "/home/node", Conditions: []config.Condition{{ConditionRef: "strings-equal", Params: []config.Param{{Name: "outputFormat", Value: "junit"}}}}},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "influx",
						Type: "influx",
						Parameters: []map[string]interface{}{
							{"Name": "step_data"}, {"fields": []map[string]string{{"name": "bats"}}},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
