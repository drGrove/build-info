package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"text/template"

	"github.com/shurcooL/githubv4"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
	"k8s.io/utils/strings/slices"
)

var labelsQuery struct {
	Repository struct {
		PullRequest struct {
			Labels struct {
				PageInfo struct {
					EndCursor   githubv4.String
					StartCursor githubv4.String
					HasNextPage githubv4.Boolean
				} `graphql:"pageInfo"`
				Edges []struct {
					Node struct {
						Name string
					} `graphql:"node"`
				} `graphql:"edges"`
			} `graphql:"labels(first: 100, after: $labelsCursor)"`
		} `graphql:"pullRequest(number: $prnumber)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

var (
	repositoryOwner   string
	repositoryName    string
	pullRequestNumber int64
)

// Build information. Populated at build-time.
var (
	Version   string
	BuildDate string
	GoVersion = runtime.Version()
)

// versionInfoTmpl contains the template used by Info.
var versionInfoTmpl = `
{{.program}}, version {{.version}}
  build date:       {{.buildDate}}
  go version:       {{.goVersion}}
  platform:         {{.platform}}
`

func hasLabel(cCtx *cli.Context) error {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return cli.Exit("Authentication token for Github not provided", 126)
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	var client *githubv4.Client
	if os.Getenv("GITHUB_ENDPOINT") != "" {
		client = githubv4.NewEnterpriseClient(os.Getenv("GITHUB_ENDPOINT"), httpClient)
	} else {
		client = githubv4.NewClient(httpClient)
	}
	labels := []string{}
	variables := map[string]interface{}{
		"owner":        githubv4.String(repositoryOwner),
		"name":         githubv4.String(repositoryName),
		"prnumber":     githubv4.Int(pullRequestNumber),
		"labelsCursor": (*githubv4.String)(nil),
	}
	for {
		err := client.Query(context.Background(), &labelsQuery, variables)
		if err != nil {
			return cli.Exit("Could not query for labels", 126)
		}
		for _, label := range labelsQuery.Repository.PullRequest.Labels.Edges {
			labels = append(labels, label.Node.Name)
		}
		if !labelsQuery.Repository.PullRequest.Labels.PageInfo.HasNextPage {
			break
		}
		variables["labelsCursor"] = githubv4.NewString(labelsQuery.Repository.PullRequest.Labels.PageInfo.EndCursor)
	}
	search := cCtx.Args().Get(0)
	if !slices.Contains(labels, search) {
		if !cCtx.Bool("quiet") {
			fmt.Fprintf(cCtx.App.Writer, "%v not found\n", search)
		}
		return cli.Exit("", 1)
	}
	if !cCtx.Bool("quiet") {
		fmt.Fprintf(cCtx.App.Writer, "%v found\n", search)
	}
	return cli.Exit("", 0)
}

var authors = []*cli.Author{
	&cli.Author{
		Name:  "Danny Grove",
		Email: "danny@drgrovellc.com",
	},
}
var copyright = "(c) 2022 Danny Grove"

func init() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		m := map[string]string{
			"program":   "build-info",
			"version":   Version,
			"buildDate": BuildDate,
			"goVersion": GoVersion,
			"platform":  runtime.GOOS + "/" + runtime.GOARCH,
		}
		t := template.Must(template.New("version").Parse(versionInfoTmpl))

		var buf bytes.Buffer
		if err := t.ExecuteTemplate(&buf, "version", m); err != nil {
			panic(err)
		}
		fmt.Fprintln(cCtx.App.Writer, buf.String())
	}
}

func baseAction(cCtx *cli.Context) error {
	return nil
}

var hasLabelCommand = &cli.Command{
	Name:  "has-label",
	Usage: "check if a PR has a label",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "quiet",
			Aliases: []string{"q"},
			Value:   false,
			Usage:   "Output to stdout",
		},
	},
	Action: hasLabel,
}

var pullRequestSubCommand = &cli.Command{
	Name:    "pull-request",
	Aliases: []string{"pr"},
	Usage:   "Get info on pull-request",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:        "number",
			Usage:       "Set the pull request number",
			Destination: &pullRequestNumber,
			EnvVars:     []string{"GITHUB_PR", "DRONE_PULL_REQUEST"},
		},
	},
	Subcommands: []*cli.Command{hasLabelCommand},
}

func main() {
	app := &cli.App{
		Name:                   "build-info",
		Usage:                  "Get info from a build",
		Version:                Version,
		HideVersion:            false,
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,
		Authors:                authors,
		Copyright:              copyright,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "repo-owner",
				Usage:       "Set the owner of the repository",
				Destination: &repositoryOwner,
				EnvVars:     []string{"GITHUB_REPO_OWNER", "DRONE_REPO_OWNER"},
			},
			&cli.StringFlag{
				Name:        "repo-name",
				Usage:       "Set the name of the repository to query",
				Destination: &repositoryName,
				EnvVars:     []string{"GITHUB_REPO_NAME", "DRONE_REPO_NAME"},
			},
		},
		Commands: []*cli.Command{pullRequestSubCommand},
	}
	app.Run(os.Args)
}
