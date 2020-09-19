# gogs-filter

This service filters webhook requests from Gogs and only passes the requests through to a Tekton EventListener when the git ref matches what's been defined in the rules.

This was developed because OpenShift Pipelines 1.0 did not support [CEL interceptors](https://bigkevmcd.github.io/kubernetes/tekton/pipeline/2020/02/05/cel-interception.html).

This enables you to kick different pipelineruns off depending on the git branch.

The service is configured with a command line argument `-rulesjson` or an environment variable `RULESJSON`. This configuration option expects a string in the following format (formatted for readability):

```
[
  {"ref":"refs/heads/new-feature", "target":"http://abc:8080"},
  {"ref":"refs/heads/master", "target":"http://el-pipeline.{{ (index .commits 0).committer.username }}.svc.cluster.local:8080"},
  {"ref":"", "target":"http://def:8080"}
]
```

The rules are evaluated in order and the first match is returned. If an entry with an empty `ref` field exists, that entry is treated as a catch-all and will match all incoming `ref`s.

If none of the rules match and there is no catch-all, the `gogs-filter` will drop the request and not pass it through to any target service.

If a `target` field contains `{{` and `}}` it is treated as a Go template and the incoming request body is evaluated against that template.
