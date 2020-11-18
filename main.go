package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/bborbe/argument"
	"github.com/golang/glog"
)

type application struct {
	JiraURL         string `required:"true" arg:"url" env:"URL"`
	JiraUser        string `required:"true" arg:"username" env:"USERNAME"`
	JiraPassword    string `required:"true" arg:"password" env:"PASSWORD"`
	JiraProjectKey  string `required:"true" arg:"project-key" env:"PROJECT_KEY"`
	JiraIssueType   string `required:"true" arg:"issue-type" env:"ISSUE_TYPE"`
	JiraSummary     string `required:"true" arg:"summary" env:"SUMMARY"`
	JiraDescription string `required:"true" arg:"description" env:"DESCRIPTION"`
	ParentIssueKey  string `required:"false" arg:"parent-issue-key" env:"PARENT_ISSUE_KEY"`
	Assignee        string `required:"false" arg:"assignee" env:"ASSIGNEE"`
}

type ProjectKey string

func (i ProjectKey) String() string {
	return string(i)
}

type IssueKey string

func (i IssueKey) String() string {
	return string(i)
}

type ProjectKeys []ProjectKey

func (t ProjectKeys) Contains(projectKey ProjectKey) bool {
	for _, i := range t {
		if i == projectKey {
			return true
		}
	}
	return false
}

type IssueType string

func (i IssueType) String() string {
	return string(i)
}

type IssueTypes []IssueType

func (t IssueTypes) Contains(issueType IssueType) bool {
	for _, i := range t {
		if i == issueType {
			return true
		}
	}
	return false
}

type Summary string

func (i Summary) String() string {
	return string(i)
}

type Description string

func (i Description) String() string {
	return string(i)
}

type IssueTypeResult struct {
	MaxResults int  `json:"maxResults"`
	StartAt    int  `json:"startAt"`
	Total      int  `json:"total"`
	IsLast     bool `json:"isLast"`
	Values     []struct {
		Self        string    `json:"self"`
		ID          string    `json:"id"`
		Description string    `json:"description"`
		IconURL     string    `json:"iconUrl"`
		Name        IssueType `json:"name"`
		Subtask     bool      `json:"subtask"`
	} `json:"values"`
}

type ProjectResult []struct {
	Expand     string     `json:"expand"`
	Self       string     `json:"self"`
	ID         string     `json:"id"`
	Key        ProjectKey `json:"key"`
	Name       string     `json:"name"`
	AvatarUrls struct {
		Four8X48  string `json:"48x48"`
		Two4X24   string `json:"24x24"`
		One6X16   string `json:"16x16"`
		Three2X32 string `json:"32x32"`
	} `json:"avatarUrls"`
	ProjectCategory struct {
		Self        string `json:"self"`
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"projectCategory,omitempty"`
	ProjectTypeKey string `json:"projectTypeKey"`
}

type Issue struct {
	Fields struct {
		Project struct {
			Key ProjectKey `json:"key"`
		} `json:"project"`
		Parent struct {
			Key IssueKey `json:"key"`
		} `json:"parent"`
		Assignee struct {
			Name Username `json:"name"`
		} `json:"assignee"`
		Summary     Summary     `json:"summary"`
		Description Description `json:"description"`
		Issuetype   struct {
			Name IssueType `json:"name"`
		} `json:"issuetype"`
	} `json:"fields"`
}

type CreateIssueResponse struct {
	ID   string   `json:"id"`
	Key  IssueKey `json:"key"`
	Self string   `json:"self"`
}

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())
	_ = flag.Set("logtostderr", "true")
	var app application
	if err := argument.Parse(&app); err != nil {
		glog.Fatalf("parse args failed: %v", err)
	}

	app.Run()
}

func (a *application) Run() {
	projectKey := ProjectKey(a.JiraProjectKey)
	if !a.getProjectKeys().Contains(projectKey) {
		glog.Fatalf("projectKey '%s' does not exists", projectKey)
	}
	issueType := IssueType(a.JiraIssueType)
	if !a.getIssueTypes(projectKey).Contains(issueType) {
		glog.Fatalf("issueType '%s' does not exists", issueType)
	}
	a.createIssue(
		projectKey,
		issueType,
		Summary(a.JiraSummary),
		Description(a.JiraDescription),
		IssueKey(a.ParentIssueKey),
		Username(a.Assignee),
	)
}

func (a *application) getProjectKeys() ProjectKeys {
	glog.V(2).Infof("get project keys started")
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"%s/rest/api/2/project",
			a.JiraURL,
		),
		nil,
	)
	if err != nil {
		glog.Fatalf("create request failed: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(a.JiraUser, a.JiraPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		glog.Fatalf("request failed with status %d: %s", resp.StatusCode, resp.Status)
	}
	var projectResult ProjectResult
	if err := json.NewDecoder(resp.Body).Decode(&projectResult); err != nil {
		glog.Fatalf("parse projects failed: %v", err)
	}
	glog.V(2).Infof("found %d issueTypes", len(projectResult))

	var result ProjectKeys
	for _, value := range projectResult {
		glog.V(2).Infof("ProjectKey: %s", value.Name)
		result = append(result, value.Key)
	}
	return result
}

func (a *application) getIssueTypes(projectKey ProjectKey) IssueTypes {
	glog.V(2).Infof("get issue types started")
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"%s/rest/api/2/issue/createmeta/%s/issuetypes",
			a.JiraURL,
			projectKey,
		),
		nil,
	)
	if err != nil {
		glog.Fatalf("create request failed: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(a.JiraUser, a.JiraPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		glog.Fatalf("request failed with status %d: %s", resp.StatusCode, resp.Status)
	}
	var issueTypeResult IssueTypeResult
	if err := json.NewDecoder(resp.Body).Decode(&issueTypeResult); err != nil {
		glog.Fatalf("parse issuetypes failed: %v", err)
	}

	glog.V(2).Infof("found %d issueTypes", len(issueTypeResult.Values))
	var result IssueTypes
	for _, value := range issueTypeResult.Values {
		glog.V(2).Infof("IssueType: %s", value.Name)
		result = append(result, value.Name)
	}
	return result
}

type Username string

func (u Username) String() string {
	return string(u)
}

func (a *application) createIssue(
	projectKey ProjectKey,
	issueType IssueType,
	summary Summary,
	description Description,
	parentIssueKey IssueKey,
	assignee Username,
) {
	glog.V(2).Infof("create issue started")

	var issue Issue
	issue.Fields.Project.Key = projectKey
	issue.Fields.Parent.Key = parentIssueKey
	issue.Fields.Assignee.Name = assignee
	issue.Fields.Summary = summary
	issue.Fields.Description = description
	issue.Fields.Issuetype.Name = issueType

	content, err := json.Marshal(issue)
	if err != nil {
		glog.V(2).Infof("marshal json failed: %v", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/rest/api/2/issue/", a.JiraURL),
		bytes.NewBuffer(content),
	)
	if err != nil {
		glog.Fatalf("create request failed: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(a.JiraUser, a.JiraPassword)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.V(2).Infof("create issue request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		content, _ := ioutil.ReadAll(resp.Body)

		glog.Fatalf("request failed with status %d: %s: %s", resp.StatusCode, resp.Status, string(content))
	}

	var createIssueResponse CreateIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&createIssueResponse); err != nil {
		glog.Fatalf("parse create issue result failed: %v", err)
	}

	glog.V(2).Infof("issue create %s/browse/%s", a.JiraURL, createIssueResponse.Key)

}
