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

type detectExecuteScanOptions struct {
	Token                      string   `json:"token,omitempty"`
	CodeLocation               string   `json:"codeLocation,omitempty"`
	ProjectName                string   `json:"projectName,omitempty"`
	Scanners                   []string `json:"scanners,omitempty"`
	ScanPaths                  []string `json:"scanPaths,omitempty"`
	DependencyPath             string   `json:"dependencyPath,omitempty"`
	Unmap                      bool     `json:"unmap,omitempty"`
	ScanProperties             []string `json:"scanProperties,omitempty"`
	ServerURL                  string   `json:"serverUrl,omitempty"`
	Groups                     []string `json:"groups,omitempty"`
	FailOn                     []string `json:"failOn,omitempty"`
	VersioningModel            string   `json:"versioningModel,omitempty" validate:"oneof=major major-minor semantic full"`
	Version                    string   `json:"version,omitempty"`
	CustomScanVersion          string   `json:"customScanVersion,omitempty"`
	ProjectSettingsFile        string   `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile         string   `json:"globalSettingsFile,omitempty"`
	M2Path                     string   `json:"m2Path,omitempty"`
	InstallArtifacts           bool     `json:"installArtifacts,omitempty"`
	IncludedPackageManagers    []string `json:"includedPackageManagers,omitempty"`
	ExcludedPackageManagers    []string `json:"excludedPackageManagers,omitempty"`
	MavenExcludedScopes        []string `json:"mavenExcludedScopes,omitempty"`
	DetectTools                []string `json:"detectTools,omitempty"`
	ScanOnChanges              bool     `json:"scanOnChanges,omitempty"`
	CustomEnvironmentVariables []string `json:"customEnvironmentVariables,omitempty"`
}

type detectExecuteScanInflux struct {
	step_data struct {
		fields struct {
			detect bool
		}
		tags struct {
		}
	}
	detect_data struct {
		fields struct {
			vulnerabilities       int
			major_vulnerabilities int
			minor_vulnerabilities int
			components            int
			policy_violations     int
		}
		tags struct {
		}
	}
}

func (i *detectExecuteScanInflux) persist(path, resourceName string) {
	measurementContent := []struct {
		measurement string
		valType     string
		name        string
		value       interface{}
	}{
		{valType: config.InfluxField, measurement: "step_data", name: "detect", value: i.step_data.fields.detect},
		{valType: config.InfluxField, measurement: "detect_data", name: "vulnerabilities", value: i.detect_data.fields.vulnerabilities},
		{valType: config.InfluxField, measurement: "detect_data", name: "major_vulnerabilities", value: i.detect_data.fields.major_vulnerabilities},
		{valType: config.InfluxField, measurement: "detect_data", name: "minor_vulnerabilities", value: i.detect_data.fields.minor_vulnerabilities},
		{valType: config.InfluxField, measurement: "detect_data", name: "components", value: i.detect_data.fields.components},
		{valType: config.InfluxField, measurement: "detect_data", name: "policy_violations", value: i.detect_data.fields.policy_violations},
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

// DetectExecuteScanCommand Executes Synopsys Detect scan
func DetectExecuteScanCommand() *cobra.Command {
	const STEP_NAME = "detectExecuteScan"

	metadata := detectExecuteScanMetadata()
	var stepConfig detectExecuteScanOptions
	var startTime time.Time
	var influx detectExecuteScanInflux
	var logCollector *log.CollectorHook
	splunkClient := &splunk.Splunk{}

	var createDetectExecuteScanCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes Synopsys Detect scan",
		Long: `This step executes [Synopsys Detect](https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/62423113/Synopsys+Detect) scans.
Synopsys Detect command line utlity can be used to run various scans including BlackDuck and Polaris scans. This step allows users to run BlackDuck scans by default.
Please configure your BlackDuck server Url using the serverUrl parameter and the API token of your user using the apiToken parameter for this step.`,
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
			log.RegisterSecret(stepConfig.Token)

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
			detectExecuteScan(stepConfig, &telemetryData, &influx)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addDetectExecuteScanFlags(createDetectExecuteScanCmd, &stepConfig)
	return createDetectExecuteScanCmd
}

func addDetectExecuteScanFlags(cmd *cobra.Command, stepConfig *detectExecuteScanOptions) {
	cmd.Flags().StringVar(&stepConfig.Token, "token", os.Getenv("PIPER_token"), "Api token to be used for connectivity with Synopsis Detect server.")
	cmd.Flags().StringVar(&stepConfig.CodeLocation, "codeLocation", os.Getenv("PIPER_codeLocation"), "An override for the name Detect will use for the scan file it creates.")
	cmd.Flags().StringVar(&stepConfig.ProjectName, "projectName", os.Getenv("PIPER_projectName"), "Name of the Synopsis Detect (formerly BlackDuck) project.")
	cmd.Flags().StringSliceVar(&stepConfig.Scanners, "scanners", []string{`signature`}, "List of scanners to be used for Synopsis Detect (formerly BlackDuck) scan.")
	cmd.Flags().StringSliceVar(&stepConfig.ScanPaths, "scanPaths", []string{`.`}, "List of paths which should be scanned by the Synopsis Detect (formerly BlackDuck) scan.")
	cmd.Flags().StringVar(&stepConfig.DependencyPath, "dependencyPath", `.`, "Absolute Path of the dependency management file of the project. This path represents the folder which contains the pom file, package.json etc. If the project contains multiple pom files, provide the path to the parent pom file or the base folder of the project")
	cmd.Flags().BoolVar(&stepConfig.Unmap, "unmap", false, "Unmap flag will unmap all previous code locations and keep only the current scan results in the specified project version. Set this parameter to true, when the project version needs to store only the latest scan results.")
	cmd.Flags().StringSliceVar(&stepConfig.ScanProperties, "scanProperties", []string{`--blackduck.signature.scanner.memory=4096`, `--detect.timeout=6000`, `--blackduck.trust.cert=true`, `--logging.level.com.synopsys.integration=DEBUG`, `--detect.maven.excluded.scopes=test`}, "Properties passed to the Synopsis Detect (formerly BlackDuck) scan. You can find details in the [Synopsis Detect documentation](https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/622846/Using+Synopsys+Detect+Properties)")
	cmd.Flags().StringVar(&stepConfig.ServerURL, "serverUrl", os.Getenv("PIPER_serverUrl"), "Server URL to the Synopsis Detect (formerly BlackDuck) Server.")
	cmd.Flags().StringSliceVar(&stepConfig.Groups, "groups", []string{}, "Users groups to be assigned for the Project")
	cmd.Flags().StringSliceVar(&stepConfig.FailOn, "failOn", []string{`BLOCKER`}, "Mark the current build as fail based on the policy categories applied.")
	cmd.Flags().StringVar(&stepConfig.VersioningModel, "versioningModel", `major`, "The versioning model used for result reporting (based on the artifact version). Example 1.2.3 using `major` will result in version 1")
	cmd.Flags().StringVar(&stepConfig.Version, "version", os.Getenv("PIPER_version"), "Defines the version number of the artifact being build in the pipeline. It is used as source for the Detect version.")
	cmd.Flags().StringVar(&stepConfig.CustomScanVersion, "customScanVersion", os.Getenv("PIPER_customScanVersion"), "A custom version used along with the uploaded scan results.")
	cmd.Flags().StringVar(&stepConfig.ProjectSettingsFile, "projectSettingsFile", os.Getenv("PIPER_projectSettingsFile"), "Path or url to the mvn settings file that should be used as project settings file.")
	cmd.Flags().StringVar(&stepConfig.GlobalSettingsFile, "globalSettingsFile", os.Getenv("PIPER_globalSettingsFile"), "Path or url to the mvn settings file that should be used as global settings file")
	cmd.Flags().StringVar(&stepConfig.M2Path, "m2Path", os.Getenv("PIPER_m2Path"), "Path to the location of the local repository that should be used.")
	cmd.Flags().BoolVar(&stepConfig.InstallArtifacts, "installArtifacts", false, "If enabled, it will install all artifacts to the local maven repository to make them available before running detect. This is required if any maven module has dependencies to other modules in the repository and they were not installed before.")
	cmd.Flags().StringSliceVar(&stepConfig.IncludedPackageManagers, "includedPackageManagers", []string{}, "The package managers that need to be included for this scan. Providing the package manager names with this parameter will ensure that the build descriptor file of that package manager will be searched in the scan folder For the complete list of possible values for this parameter, please refer [Synopsys detect documentation](https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/631407160/Configuring+Detect+General+Properties#Detector-types-included-(Advanced))")
	cmd.Flags().StringSliceVar(&stepConfig.ExcludedPackageManagers, "excludedPackageManagers", []string{}, "The package managers that need to be excluded for this scan. Providing the package manager names with this parameter will ensure that the build descriptor file of that package manager will be ignored in the scan folder For the complete list of possible values for this parameter, please refer [Synopsys detect documentation](https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/631407160/Configuring+Detect+General+Properties#%5BhardBreak%5DDetector-types-excluded-(Advanced))")
	cmd.Flags().StringSliceVar(&stepConfig.MavenExcludedScopes, "mavenExcludedScopes", []string{}, "The maven scopes that need to be excluded from the scan. For example, setting the value 'test' will exclude all components which are defined with a test scope in maven")
	cmd.Flags().StringSliceVar(&stepConfig.DetectTools, "detectTools", []string{}, "The type of BlackDuck scanners to include while running the BlackDuck scan. By default All scanners are included. For the complete list of possible values, Please refer [Synopsys detect documentation](https://synopsys.atlassian.net/wiki/spaces/INTDOCS/pages/631407160/Configuring+Detect+General+Properties#Detect-tools-included)")
	cmd.Flags().BoolVar(&stepConfig.ScanOnChanges, "scanOnChanges", false, "This flag determines if the scan is submitted to the server. If set to true, then the scan request is submitted to the server only when changes are detected in the Open Source Bill of Materials If the flag is set to false, then the scan request is submitted to server regardless of any changes. For more details please refer to the [documentation](https://github.com/blackducksoftware/detect_rescan/blob/master/README.md)")
	cmd.Flags().StringSliceVar(&stepConfig.CustomEnvironmentVariables, "customEnvironmentVariables", []string{}, "A list of environment variables which can be set to prepare the environment to run a BlackDuck scan.")

	cmd.MarkFlagRequired("token")
	cmd.MarkFlagRequired("projectName")
	cmd.MarkFlagRequired("serverUrl")
}

// retrieve step metadata
func detectExecuteScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "detectExecuteScan",
			Aliases:     []config.Alias{},
			Description: "Executes Synopsys Detect scan",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "detectTokenCredentialsId", Description: "Jenkins 'Secret text' credentials ID containing the API token used to authenticate with the Synopsis Detect (formerly BlackDuck) Server.", Type: "jenkins", Aliases: []config.Alias{{Name: "apiTokenCredentialsId", Deprecated: false}}},
				},
				Resources: []config.StepResources{
					{Name: "buildDescriptor", Type: "stash"},
					{Name: "checkmarx", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{
						Name: "token",
						ResourceRef: []config.ResourceReference{
							{
								Name: "detectTokenCredentialsId",
								Type: "secret",
							},

							{
								Name:    "detectVaultSecretName",
								Type:    "vaultSecret",
								Default: "detect",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{{Name: "blackduckToken"}, {Name: "detectToken"}, {Name: "apiToken"}, {Name: "detect/apiToken"}},
						Default:   os.Getenv("PIPER_token"),
					},
					{
						Name:        "codeLocation",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_codeLocation"),
					},
					{
						Name:        "projectName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "detect/projectName"}},
						Default:     os.Getenv("PIPER_projectName"),
					},
					{
						Name:        "scanners",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/scanners"}},
						Default:     []string{`signature`},
					},
					{
						Name:        "scanPaths",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/scanPaths"}},
						Default:     []string{`.`},
					},
					{
						Name:        "dependencyPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/dependencyPath"}},
						Default:     `.`,
					},
					{
						Name:        "unmap",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/unmap"}},
						Default:     false,
					},
					{
						Name:        "scanProperties",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/scanProperties"}},
						Default:     []string{`--blackduck.signature.scanner.memory=4096`, `--detect.timeout=6000`, `--blackduck.trust.cert=true`, `--logging.level.com.synopsys.integration=DEBUG`, `--detect.maven.excluded.scopes=test`},
					},
					{
						Name:        "serverUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "detect/serverUrl"}},
						Default:     os.Getenv("PIPER_serverUrl"),
					},
					{
						Name:        "groups",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/groups"}},
						Default:     []string{},
					},
					{
						Name:        "failOn",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/failOn"}},
						Default:     []string{`BLOCKER`},
					},
					{
						Name:        "versioningModel",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "GENERAL", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `major`,
					},
					{
						Name: "version",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "projectVersion"}, {Name: "detect/projectVersion"}},
						Default:   os.Getenv("PIPER_version"),
					},
					{
						Name:        "customScanVersion",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STAGES", "STEPS", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_customScanVersion"),
					},
					{
						Name:        "projectSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/projectSettingsFile"}},
						Default:     os.Getenv("PIPER_projectSettingsFile"),
					},
					{
						Name:        "globalSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/globalSettingsFile"}},
						Default:     os.Getenv("PIPER_globalSettingsFile"),
					},
					{
						Name:        "m2Path",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/m2Path"}},
						Default:     os.Getenv("PIPER_m2Path"),
					},
					{
						Name:        "installArtifacts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "includedPackageManagers",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/includedPackageManagers"}},
						Default:     []string{},
					},
					{
						Name:        "excludedPackageManagers",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/excludedPackageManagers"}},
						Default:     []string{},
					},
					{
						Name:        "mavenExcludedScopes",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/mavenExcludedScopes"}},
						Default:     []string{},
					},
					{
						Name:        "detectTools",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "detect/detectTools"}},
						Default:     []string{},
					},
					{
						Name:        "scanOnChanges",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "customEnvironmentVariables",
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
				{Name: "openjdk", Image: "openjdk:11", WorkingDir: "/root", Options: []config.Option{{Name: "-u", Value: "0"}}},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "influx",
						Type: "influx",
						Parameters: []map[string]interface{}{
							{"Name": "step_data"}, {"fields": []map[string]string{{"name": "detect"}}},
							{"Name": "detect_data"}, {"fields": []map[string]string{{"name": "vulnerabilities"}, {"name": "major_vulnerabilities"}, {"name": "minor_vulnerabilities"}, {"name": "components"}, {"name": "policy_violations"}}},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
