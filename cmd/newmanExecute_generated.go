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

type newmanExecuteOptions struct {
	NewmanCollection     string   `json:"newmanCollection,omitempty"`
	NewmanRunCommand     string   `json:"newmanRunCommand,omitempty"`
	RunOptions           []string `json:"runOptions,omitempty"`
	NewmanInstallCommand string   `json:"newmanInstallCommand,omitempty"`
	NewmanEnvironment    string   `json:"newmanEnvironment,omitempty"`
	NewmanGlobals        string   `json:"newmanGlobals,omitempty"`
	FailOnError          bool     `json:"failOnError,omitempty"`
	CfAppsWithSecrets    []string `json:"cfAppsWithSecrets,omitempty"`
}

type newmanExecuteInflux struct {
	step_data struct {
		fields struct {
			newman bool
		}
		tags struct {
		}
	}
}

func (i *newmanExecuteInflux) persist(path, resourceName string) {
	measurementContent := []struct {
		measurement string
		valType     string
		name        string
		value       interface{}
	}{
		{valType: config.InfluxField, measurement: "step_data", name: "newman", value: i.step_data.fields.newman},
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

// NewmanExecuteCommand Installs newman and executes specified newman collections.
func NewmanExecuteCommand() *cobra.Command {
	const STEP_NAME = "newmanExecute"

	metadata := newmanExecuteMetadata()
	var stepConfig newmanExecuteOptions
	var startTime time.Time
	var influx newmanExecuteInflux
	var logCollector *log.CollectorHook

	var createNewmanExecuteCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Installs newman and executes specified newman collections.",
		Long:  `This script executes [Postman](https://www.getpostman.com) tests from a collection via the [Newman](https://www.getpostman.com/docs/v6/postman/collection_runs/command_line_integration_with_newman) command line tool.`,
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
					splunk.Send(&telemetryData, logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunk.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			newmanExecute(stepConfig, &telemetryData, &influx)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addNewmanExecuteFlags(createNewmanExecuteCmd, &stepConfig)
	return createNewmanExecuteCmd
}

func addNewmanExecuteFlags(cmd *cobra.Command, stepConfig *newmanExecuteOptions) {
	cmd.Flags().StringVar(&stepConfig.NewmanCollection, "newmanCollection", `**/*.postman_collection.json`, "The test collection that should be executed. This could also be a file pattern.")
	cmd.Flags().StringVar(&stepConfig.NewmanRunCommand, "newmanRunCommand", os.Getenv("PIPER_newmanRunCommand"), "+++ Deprecated +++ Please use list parameter `runOptions` instead.")
	cmd.Flags().StringSliceVar(&stepConfig.RunOptions, "runOptions", []string{`run`, `{{.NewmanCollection}}`, `--reporters`, `cli,junit,html`, `--reporter-junit-export`, `target/newman/TEST-{{.CollectionDisplayName}}.xml`, `--reporter-html-export`, `target/newman/TEST-{{.CollectionDisplayName}}.html`}, "The newman command that will be executed inside the docker container.")
	cmd.Flags().StringVar(&stepConfig.NewmanInstallCommand, "newmanInstallCommand", `npm install newman newman-reporter-html --global --quiet`, "The shell command that will be executed inside the docker container to install Newman.")
	cmd.Flags().StringVar(&stepConfig.NewmanEnvironment, "newmanEnvironment", os.Getenv("PIPER_newmanEnvironment"), "Specify an environment file path or URL. Environments provide a set of variables that one can use within collections.")
	cmd.Flags().StringVar(&stepConfig.NewmanGlobals, "newmanGlobals", os.Getenv("PIPER_newmanGlobals"), "Specify the file path or URL for global variables. Global variables are similar to environment variables but have a lower precedence and can be overridden by environment variables having the same name.")
	cmd.Flags().BoolVar(&stepConfig.FailOnError, "failOnError", true, "Defines the behavior, in case tests fail.")
	cmd.Flags().StringSliceVar(&stepConfig.CfAppsWithSecrets, "cfAppsWithSecrets", []string{}, "List of CloudFoundry apps with secrets")

}

// retrieve step metadata
func newmanExecuteMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "newmanExecute",
			Aliases:     []config.Alias{},
			Description: "Installs newman and executes specified newman collections.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Resources: []config.StepResources{
					{Name: "tests", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "newmanCollection",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `**/*.postman_collection.json`,
					},
					{
						Name:        "newmanRunCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_newmanRunCommand"),
					},
					{
						Name:        "runOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{`run`, `{{.NewmanCollection}}`, `--reporters`, `cli,junit,html`, `--reporter-junit-export`, `target/newman/TEST-{{.CollectionDisplayName}}.xml`, `--reporter-html-export`, `target/newman/TEST-{{.CollectionDisplayName}}.html`},
					},
					{
						Name:        "newmanInstallCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `npm install newman newman-reporter-html --global --quiet`,
					},
					{
						Name:        "newmanEnvironment",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_newmanEnvironment"),
					},
					{
						Name:        "newmanGlobals",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_newmanGlobals"),
					},
					{
						Name:        "failOnError",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "cfAppsWithSecrets",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
				},
			},
			Containers: []config.Container{
				{Name: "newman", Image: "node:lts-stretch", WorkingDir: "/home/node"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "influx",
						Type: "influx",
						Parameters: []map[string]interface{}{
							{"Name": "step_data"}, {"fields": []map[string]string{{"name": "newman"}}},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
