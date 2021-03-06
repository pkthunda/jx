package cmd

import (
	"fmt"
	"io"

	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	"github.com/jenkins-x/jx/pkg/log"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
)

var (
	create_git_server_long = templates.LongDesc(`
		Adds a new Git Server URL
`)

	create_git_server_example = templates.Examples(`
		# Add a new Git server
		jx create git server -kind bitbucket --url http://bitbucket.org

		# Add a new Git server with a name
		jx create git server -k bitbucket -u http://bitbucket.org -n MyBitBucket 

		For more documentation see: [https://jenkins-x.io/developing/git/](https://jenkins-x.io/developing/git/)

	`)

	gitKindToServiceName = map[string]string{
		"gitea": "gitea-gitea",
	}
)

// CreateGitServerOptions the options for the create spring command
type CreateGitServerOptions struct {
	CreateOptions

	Name string
	Kind string
	URL string
}

// NewCmdCreateGitServer creates a command object for the "create" command
func NewCmdCreateGitServer(f Factory, in terminal.FileReader, out terminal.FileWriter, errOut io.Writer) *cobra.Command {
	options := &CreateGitServerOptions{
		CreateOptions: CreateOptions{
			CommonOptions: CommonOptions{
				Factory: f,
				In:      in,

				Out: out,
				Err: errOut,
			},
		},
	}

	cmd := &cobra.Command{
		Use:     "server kind [url]",
		Short:   "Creates a new Git server URL",
		Aliases: []string{"provider"},
		Long:    create_git_server_long,
		Example: create_git_server_example,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&options.Name, "name", "n", "", "The name for the Git server being created")
	cmd.Flags().StringVarP(&options.Kind, "kind", "k", "", "The kind of Git server being created")
	cmd.Flags().StringVarP(&options.URL, "url", "u", "", "The git server URL")
	return cmd
}

// Run implements the command
func (o *CreateGitServerOptions) Run() error {
	args := o.Args
	kind := o.Kind
	if kind == "" {
		if len(args) < 1 {
			return util.MissingOption("kind")
		}
		kind = args[0]
	}
	name := o.Name
	if name == "" {
		name = kind
	}
	gitUrl := o.URL
	if gitUrl == "" {
		if len(args) > 1 {
			gitUrl = args[1]
		} else {
			// lets try find the git URL based on the provider
			serviceName := gitKindToServiceName[kind]
			if serviceName != "" {
				url, err := o.findService(serviceName)
				if err != nil {
					return fmt.Errorf("Failed to find %s Git service %s: %s", kind, serviceName, err)
				}
				gitUrl = url
			}
		}
	}
	if gitUrl == "" {
		return util.MissingOption("url")
	}

	authConfigSvc, err := o.CreateGitAuthConfigService()
	if err != nil {
		return err
	}
	config := authConfigSvc.Config()
	server := config.GetOrCreateServerName(gitUrl, name, kind)
	config.CurrentServer = gitUrl
	err = authConfigSvc.SaveConfig()
	if err != nil {
		return err
	}
	log.Infof("Added Git server %s for URL %s\n", util.ColorInfo(name), util.ColorInfo(gitUrl))

	err = o.ensureGitServiceCRD(server)
	if err != nil {
	  return err
	}
	return nil
}
