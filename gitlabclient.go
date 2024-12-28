package gitlabclient

import (
	"encoding/base64"
	"errors"
	"os"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var BaseURL string = "https://gitlab.com"
var BaseBranch string = "master"
var client *gitlab.Client

type GitlabClient struct {
	URL     string
	Project string
	Branch  string
	Path    string
	Message string
}

func getClient(url string) (*gitlab.Client, error) {
	if client != nil {
		return client, nil
	}
	token, ok := os.LookupEnv("GITLAB_API_TOKEN")
	if !ok {
		return nil, errors.New("GITLAB_API_TOKEN not set\n")
	}
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(url))
	return client, err
}

func CreateBranch(g *GitlabClient) (*gitlab.Branch, error) {
	client, err := getClient(g.URL)
	if err != nil {
		return nil, err
	}
	branch, _, err := client.Branches.CreateBranch(g.Project, &gitlab.CreateBranchOptions{
		Branch: &g.Branch,
		Ref:    &BaseBranch,
	})
	return branch, err
}

func CreateCommit(g *GitlabClient, files map[string]string) (*gitlab.Commit, error) {
	client, err := getClient(g.URL)
	if err != nil {
		return nil, err
	}
	_, _, err = client.Branches.GetBranch(g.Project, g.Branch)
	if err != nil {
		// TODO if "404 Not Found"...
		_, err = CreateBranch(g)
		if err != nil {
			return nil, err
		}
	}
	var actions []*gitlab.CommitActionOptions
	for k, v := range files {
		actions = append(actions, &gitlab.CommitActionOptions{
			Action:   gitlab.Ptr(gitlab.FileUpdate),
			FilePath: &k,
			Content:  &v,
		})
	}
	commit, _, err := client.Commits.CreateCommit(g.Project, &gitlab.CreateCommitOptions{
		Branch:        &g.Branch,
		CommitMessage: &g.Message,
		Actions:       actions,
	})
	return commit, err
}

func CreateMergeRequest(g *GitlabClient) (*gitlab.MergeRequest, error) {
	client, err := getClient(g.URL)
	if err != nil {
		return nil, err
	}
	forkParent, err := GetForkParent(g)
	if err != nil {
		return nil, err
	}
	mergeRequest, _, err := client.MergeRequests.CreateMergeRequest(g.Project, &gitlab.CreateMergeRequestOptions{
		Title:           &g.Message,
		SourceBranch:    &g.Branch,
		TargetBranch:    &BaseBranch,
		TargetProjectID: &forkParent.ID,
	})
	return mergeRequest, nil
}

func DeleteBranch(g *GitlabClient) (*gitlab.Response, error) {
	client, err := getClient(g.URL)
	if err != nil {
		return nil, err
	}
	return client.Branches.DeleteBranch(g.Project, g.Branch)
}

func GetFile(g *GitlabClient) ([]byte, error) {
	client, err := getClient(g.URL)
	if err != nil {
		return nil, err
	}
	file, _, err := client.RepositoryFiles.GetFile(g.Project, g.Path, &gitlab.GetFileOptions{
		Ref: &g.Branch,
	})
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(file.Content)
}

func GetForkParent(g *GitlabClient) (*gitlab.ForkParent, error) {
	project, err := GetProject(g)
	if err != nil {
		return nil, err
	}
	return project.ForkedFromProject, nil
}

func GetProject(g *GitlabClient) (*gitlab.Project, error) {
	client, err := getClient(g.URL)
	if err != nil {
		return nil, err
	}
	project, _, err := client.Projects.GetProject(g.Project, &gitlab.GetProjectOptions{})
	if err != nil {
		return nil, err
	}
	return project, nil
}
