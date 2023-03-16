package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type dataSlice map[string]PullRequest

type PullRequest struct {
	Name        string
	CommitLink  string
	Link        string
	MergedAt    time.Time
	MergeCommit string
}

type kv struct {
	Key   string
	Value time.Time
}

type ResourceTemplate struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Targets []struct {
		Namespace struct {
			Ref string `yaml:"$ref"`
		} `yaml:"namespace"`
		Ref string `yaml:"ref"`
	} `yaml:"targets"`
}

type YamlFile struct {
	ResourceTemplates []ResourceTemplate `yaml:"resourceTemplates"`
}

func main() {
	repo := getEnv("REPO", "insights-rbac")
	fromCommit := getEnv("FROM_COMMIT", "")
	isInProduction := getIntEnv("IS_PR_IN_PRODUCTION", 0)
	owner := getEnv("OWNER", "RedHatInsights")
	commitListRange := getIntEnv("COMMIT_LIST_RANGE", 100)
	accessToken := getEnv("ACCESS_TOKEN", "")
	appInterfaceNamespace := getEnv("APP_INTERFACE_NAMESPACE", "/services/insights/rbac/namespaces/rbac-prod.yml")
	pathToAppInterface := getEnv("PATH_TO_APP_INTERFACE", "/Users/liborpichler/Projects/app-interface/")
	pathToDeployClowderYamlAppInterface := getEnv("PATH_TO_DEPLOY_YAML_APP_INTERFACE", "data/services/insights/rbac/deploy-clowder.yml")

	// Read the YAML file into a byte slice
	yamlFile, err := ioutil.ReadFile(pathToAppInterface + pathToDeployClowderYamlAppInterface)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal the YAML data into a YamlFile struct
	var data YamlFile
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		log.Fatal(err)
	}

	if fromCommit == "" {
		for _, rt := range data.ResourceTemplates {
			for _, target := range rt.Targets {
				if target.Namespace.Ref == appInterfaceNamespace {
					fmt.Println("Current commit in production: " + target.Ref)
					fromCommit = target.Ref
				}
			}
		}
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	isCommitPRInProduction := ""
	if isInProduction != 0 {
		prInput, _, err := client.PullRequests.Get(ctx, owner, repo, isInProduction)
		if err != nil {
			fmt.Printf("Error %v\n", err)
			return
		}

		isCommitPRInProduction = *prInput.MergeCommitSHA
	}

	uniqPRs := make(dataSlice, 1000)
	indPR := make(map[string]time.Time, 1000)

	fmt.Print("Fetching commits...")
	options := github.CommitsListOptions{ListOptions: github.ListOptions{PerPage: commitListRange, Page: 1}}
	processCommits(ctx, client, owner, repo, options, uniqPRs, indPR)

	fmt.Println("Done")

	sortedPRsbyMergedAt := sortMapByValue(indPR)

	if isCommitPRInProduction != "" {
		fmt.Printf("Searching for commit %s to check its presence in production...\n", isCommitPRInProduction)
		prFound := false
		for _, kv := range sortedPRsbyMergedAt {
			pr := uniqPRs[kv.Key]
			if pr.MergeCommit == isCommitPRInProduction {
				prFound = true
				break
			}
			if pr.MergeCommit == fromCommit {
				break
			}
		}

		if prFound {
			fmt.Printf("\nNO, PR is NOT in production.\n")
		} else {
			fmt.Printf("\nYES, PR is production.\n")
		}
	} else {
		for _, kv := range sortedPRsbyMergedAt {
			pr := uniqPRs[kv.Key]
			// list PRs until stop commit(fromCommit) is found
			if pr.MergeCommit == fromCommit {
				break
			}

			printPRInfo(pr)
		}
	}
}

func printPRInfo(pr PullRequest) {
	const separator = "===="
	format := "%s\n%v\n - %v\n - QE\n%s\n\n"
	fmt.Printf(format, separator, "["+pr.MergedAt.Format("2006-01-02")+"] "+pr.Name, pr.Link, separator)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if len(valueStr) == 0 {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		fmt.Printf("Error: invalid value for environment variable %s, using default value (%d)\n", key, defaultValue)
		return defaultValue
	}
	return value
}

func sortMapByValue(m map[string]time.Time) []kv {
	kvs := make([]kv, len(m))
	i := 0
	for k, v := range m {
		kvs[i] = kv{k, v}
		i++
	}

	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Value.After(kvs[j].Value)
	})

	return kvs
}

func processCommits(ctx context.Context, client *github.Client, owner, repo string, options github.CommitsListOptions, uniqPRs map[string]PullRequest, indPR map[string]time.Time) {
	repoCommits, _, err := client.Repositories.ListCommits(ctx, owner, repo, &options)
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}

	for _, v := range repoCommits {
		prs, _, err := client.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, *v.SHA, nil)
		if err != nil {
			fmt.Printf("Error %v\n", err)
			continue
		}

		for _, pr := range prs {
			uniqPRs[*pr.HTMLURL] = PullRequest{MergeCommit: *pr.MergeCommitSHA, MergedAt: *pr.MergedAt, Name: *pr.Title, Link: *pr.HTMLURL}
			indPR[*pr.HTMLURL] = *pr.MergedAt
		}
	}
}
