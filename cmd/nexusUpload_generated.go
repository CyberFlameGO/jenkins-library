// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type nexusUploadOptions struct {
	Version            string `json:"version,omitempty" validate:"oneof=nexus2 nexus3"`
	Format             string `json:"format,omitempty" validate:"oneof=maven npm"`
	Url                string `json:"url,omitempty"`
	MavenRepository    string `json:"mavenRepository,omitempty"`
	NpmRepository      string `json:"npmRepository,omitempty"`
	GroupID            string `json:"groupId,omitempty"`
	ArtifactID         string `json:"artifactId,omitempty"`
	GlobalSettingsFile string `json:"globalSettingsFile,omitempty"`
	M2Path             string `json:"m2Path,omitempty"`
	Username           string `json:"username,omitempty"`
	Password           string `json:"password,omitempty"`
}

// NexusUploadCommand Upload artifacts to Nexus Repository Manager
func NexusUploadCommand() *cobra.Command {
	const STEP_NAME = "nexusUpload"

	metadata := nexusUploadMetadata()
	var stepConfig nexusUploadOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createNexusUploadCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Upload artifacts to Nexus Repository Manager",
		Long: `Upload build artifacts to a Nexus Repository Manager.

Supports MTA, npm and (multi-module) Maven projects.
MTA files will be uploaded to a Maven repository.

The uploaded file-type depends on your project structure and step configuration.
To upload Maven projects, you need a pom.xml in the project root and set the mavenRepository option.
To upload MTA projects, you need a mta.yaml in the project root and set the mavenRepository option.
To upload npm projects, you need a package.json in the project root and set the npmRepository option.

If the 'format' option is set, the 'URL' can contain the full path including the repository ID. Providing the 'npmRepository' or the 'mavenRepository' parameter(s) is not necessary.

npm:
Publishing npm projects makes use of npm's "publish" command.
It requires a "package.json" file in the project's root directory which has "version" set and is not delared as "private".
To find out what will be published, run "npm publish --dry-run" in the project's root folder.
It will use your gitignore file to exclude the mached files from publishing.
Note: npm's gitignore parser might yield different results from your git client, to ignore a "foo" directory globally use the glob pattern "**/foo".

If an image for mavenExecute is configured, and npm packages are to be published, the image must have npm installed.`,
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
			log.RegisterSecret(stepConfig.Username)
			log.RegisterSecret(stepConfig.Password)

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
			nexusUpload(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addNexusUploadFlags(createNexusUploadCmd, &stepConfig)
	return createNexusUploadCmd
}

func addNexusUploadFlags(cmd *cobra.Command, stepConfig *nexusUploadOptions) {
	cmd.Flags().StringVar(&stepConfig.Version, "version", `nexus3`, "The Nexus Repository Manager version. Currently supported are 'nexus2' and 'nexus3'.")
	cmd.Flags().StringVar(&stepConfig.Format, "format", `maven`, "The format/registry type. Currently supported are 'maven' and 'npm'.")
	cmd.Flags().StringVar(&stepConfig.Url, "url", os.Getenv("PIPER_url"), "URL of the nexus. The scheme part of the URL will not be considered, because only http is supported. If the 'format' option is set, the 'URL' can contain the full path including the repository ID and providing the 'npmRepository' or the 'mavenRepository' parameter(s) is not necessary.")
	cmd.Flags().StringVar(&stepConfig.MavenRepository, "mavenRepository", os.Getenv("PIPER_mavenRepository"), "Name of the nexus repository for Maven and MTA deployments. If this is not provided, Maven and MTA deployment is implicitly disabled.")
	cmd.Flags().StringVar(&stepConfig.NpmRepository, "npmRepository", os.Getenv("PIPER_npmRepository"), "Name of the nexus repository for npm deployments. If this is not provided, npm deployment is implicitly disabled.")
	cmd.Flags().StringVar(&stepConfig.GroupID, "groupId", os.Getenv("PIPER_groupId"), "Group ID of the artifacts. Only used in MTA projects, ignored for Maven.")
	cmd.Flags().StringVar(&stepConfig.ArtifactID, "artifactId", os.Getenv("PIPER_artifactId"), "The artifact ID used for both the .mtar and mta.yaml files deployed for MTA projects, ignored for Maven.")
	cmd.Flags().StringVar(&stepConfig.GlobalSettingsFile, "globalSettingsFile", os.Getenv("PIPER_globalSettingsFile"), "Path to the mvn settings file that should be used as global settings file.")
	cmd.Flags().StringVar(&stepConfig.M2Path, "m2Path", os.Getenv("PIPER_m2Path"), "The path to the local .m2 directory, only used for Maven projects.")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "Username for accessing the Nexus endpoint.")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password for accessing the Nexus endpoint.")

	cmd.MarkFlagRequired("url")
}

// retrieve step metadata
func nexusUploadMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "nexusUpload",
			Aliases:     []config.Alias{{Name: "mavenExecute", Deprecated: false}},
			Description: "Upload artifacts to Nexus Repository Manager",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "nexusCredentialsId", Description: "Jenkins 'Username with password' credentials ID containing the technical username/password credential for accessing the nexus endpoint.", Type: "jenkins", Aliases: []config.Alias{{Name: "nexus/credentialsId", Deprecated: false}}},
				},
				Resources: []config.StepResources{
					{Name: "buildDescriptor", Type: "stash"},
					{Name: "buildResult", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "version",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "nexus/version"}},
						Default:     `nexus3`,
					},
					{
						Name: "format",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryFormat",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   `maven`,
					},
					{
						Name: "url",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUrl",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{{Name: "nexus/url"}},
						Default:   os.Getenv("PIPER_url"),
					},
					{
						Name:        "mavenRepository",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "nexus/mavenRepository"}, {Name: "nexus/repository"}},
						Default:     os.Getenv("PIPER_mavenRepository"),
					},
					{
						Name:        "npmRepository",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "nexus/npmRepository"}},
						Default:     os.Getenv("PIPER_npmRepository"),
					},
					{
						Name:        "groupId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "nexus/groupId"}},
						Default:     os.Getenv("PIPER_groupId"),
					},
					{
						Name:        "artifactId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_artifactId"),
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
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/m2Path"}},
						Default:     os.Getenv("PIPER_m2Path"),
					},
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "nexusCredentialsId",
								Param: "username",
								Type:  "secret",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUsername",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_username"),
					},
					{
						Name: "password",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "nexusCredentialsId",
								Param: "password",
								Type:  "secret",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryPassword",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_password"),
					},
				},
			},
			Containers: []config.Container{
				{Name: "mvn-npm", Image: "devxci/mbtci:1.1.1"},
			},
		},
	}
	return theMetaData
}
