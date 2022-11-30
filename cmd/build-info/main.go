package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
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
	label             string
)

// Build information. Populated at build-time.
var (
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
	GoVersion = runtime.Version()
)

// versionInfoTmpl contains the template used by Info.
var versionInfoTmpl = `
{{.program}}, version {{.version}} (branch: {{.branch}}, revision: {{.revision}})
  build user:       {{.buildUser}}
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
			return err
		}
		for _, label := range labelsQuery.Repository.PullRequest.Labels.Edges {
			fmt.Printf("%v", label.Node)
			labels = append(labels, label.Node.Name)
		}
		if !labelsQuery.Repository.PullRequest.Labels.PageInfo.HasNextPage {
			break
		}
		variables["labelsCursor"] = githubv4.NewString(labelsQuery.Repository.PullRequest.Labels.PageInfo.EndCursor)
	}
	if !slices.Contains(labels, label) {
		return cli.Exit("", 1)
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

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		m := map[string]string{
			"program":   "build-info",
			"version":   Version,
			"revision":  Revision,
			"branch":    Branch,
			"buildUser": BuildUser,
			"buildDate": BuildDate,
			"goVersion": GoVersion,
			"platform":  runtime.GOOS + "/" + runtime.GOARCH,
		}
		t := template.Must(template.New("version").Parse(versionInfoTmpl))

		var buf bytes.Buffer
		if err := t.ExecuteTemplate(&buf, "version", m); err != nil {
			panic(err)
		}
		fmt.Printf(strings.TrimSpace(buf.String()))
	}

	app := &cli.App{
		Name:                   "build-info",
		Usage:                  "Get info from a build",
		Version:                Version,
		HideVersion:            false,
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,
		Authors:                authors,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "repo-owner",
				Usage:       "Set the owner of the repository",
				Destination: &repositoryOwner,
				EnvVars:     []string{"GITHUB_REPO_OWNER"},
			},
			&cli.StringFlag{
				Name:        "repo-name",
				Usage:       "Set the name of the repository to query",
				Destination: &repositoryName,
				EnvVars:     []string{"GITHUB_REPO_NAME"},
			},
			&cli.Int64Flag{
				Name:        "pr",
				Usage:       "Set the pull request number to get the labels from",
				Destination: &pullRequestNumber,
				EnvVars:     []string{"GITHUB_PR"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "has-label",
				Usage: "check if a PR has a label",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "label",
						Aliases:     []string{"l"},
						Value:       "",
						Usage:       "Check if PR has a label",
						Destination: &label,
					},
				},
				BashComplete: func(cCtx *cli.Context) {
					fmt.Fprintf(cCtx.App.Writer, "--label\n")
				},
				Action: hasLabel,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
