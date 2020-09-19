package pkg

import "testing"

func TestSimpleRules(t *testing.T) {
	rules := `[
			{"ref":"refs/heads/new-feature","target":"http://abc"},
			{"ref":"", "target":"fallthrough"}
		]`
	tables := []struct {
		ref            string
		expectedTarget string
	}{
		{"refs/heads/main", "fallthrough"},
		{"refs/heads/new-feature", "http://abc"},
	}
	fc := InitFilterConfig(rules)

	for _, table := range tables {
		target := fc.targetForRef(table.ref, map[string]interface{}{})
		if target != table.expectedTarget {
			t.Errorf("expected target %s from ref %s but got %s instead", table.expectedTarget, table.ref, target)
		}
	}
}

func TestTemplateRule(t *testing.T) {
	rules := `[{"ref":"ref/heads/master", "target":"http://el-pipeline.{{ (index .commits 0).committer.username }}-stage.svc.cluster.local:8080"}]`

	body := `{
		"ref": "refs/heads/develop",
		"before": "28e1879d029cb852e4844d9c718537df08844e03",
		"after": "bffeb74224043ba2feb48d137756c8a9331c449a",
		"compare_url": "http://localhost:3000/unknwon/webhooks/compare/28e1879d029cb852e4844d9c718537df08844e03...bffeb74224043ba2feb48d137756c8a9331c449a",
		"commits": [
		  {
			"id": "bffeb74224043ba2feb48d137756c8a9331c449a",
			"message": "!@#0^%\u003e\u003e\u003e\u003e\u003c\u003c\u003c\u003c\u003e\u003e\u003e\u003e\n",
			"url": "http://localhost:3000/unknwon/webhooks/commit/bffeb74224043ba2feb48d137756c8a9331c449a",
			"author": {
			  "name": "Unknwon",
			  "email": "u@gogs.io",
			  "username": "authorusername"
			},
			"committer": {
			  "name": "Unknwon",
			  "email": "u@gogs.io",
			  "username": "committerusername"
			},
			"timestamp": "2017-03-13T13:52:11-04:00"
		  }
		],
		"repository": {
		  "id": 140,
		  "owner": {
			"id": 1,
			"login": "unknwon",
			"full_name": "Unknwon",
			"email": "u@gogs.io",
			"avatar_url": "https://secure.gravatar.com/avatar/d8b2871cdac01b57bbda23716cc03b96",
			"username": "ownerusername"
		  },
		  "name": "webhooks",
		  "full_name": "unknwon/webhooks",
		  "description": "",
		  "private": false,
		  "fork": false,
		  "html_url": "http://localhost:3000/unknwon/webhooks",
		  "ssh_url": "ssh://unknwon@localhost:2222/unknwon/webhooks.git",
		  "clone_url": "http://localhost:3000/unknwon/webhooks.git",
		  "website": "",
		  "stars_count": 0,
		  "forks_count": 1,
		  "watchers_count": 1,
		  "open_issues_count": 7,
		  "default_branch": "master",
		  "created_at": "2017-02-26T04:29:06-05:00",
		  "updated_at": "2017-03-13T13:51:58-04:00"
		},
		"pusher": {
		  "id": 1,
		  "login": "unknwon",
		  "full_name": "Unknwon",
		  "email": "u@gogs.io",
		  "avatar_url": "https://secure.gravatar.com/avatar/d8b2871cdac01b57bbda23716cc03b96",
		  "username": "pusherusername"
		},
		"sender": {
		  "id": 1,
		  "login": "unknwon",
		  "full_name": "Unknwon",
		  "email": "u@gogs.io",
		  "avatar_url": "https://secure.gravatar.com/avatar/d8b2871cdac01b57bbda23716cc03b96",
		  "username": "senderusername"
		}
	  }`

	fc := InitFilterConfig(rules)
	parsed, err := parseJSON([]byte(body))
	if err != nil {
		t.Errorf("error parsing body JSON: %v", err)
		return
	}
	target := fc.targetForRef("ref/heads/master", parsed)
	if target != "http://el-pipeline.committerusername-stage.svc.cluster.local:8080" {
		t.Errorf("template execution failed - got: %s", target)
	}
}

func TestIsTemplate(t *testing.T) {
	tables := []struct {
		data     string
		expected bool
	}{
		{"", false},
		{"abc", false},
		{"http://{abc}", false},
		{"http://{{abc", false},
		{"http://abc}}", false},
		{"http://{{abc}", false},
		{"http://{{abc}}", true},
		{"http://}}abc{{", false},
		{"http://{{}}", true},
		{"http://{{abc}}/def", true},
		{"{{.abc}}", true},
	}

	for _, table := range tables {
		result := isTemplate(table.data)
		if result != table.expected {
			t.Errorf("expected %v from string %s but got %v", table.expected, table.data, result)
		}
	}
}
