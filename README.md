# gogs-filter

This service filters webhook requests from Gogs and only passes the requests through to a Tekton EventListener when the git ref matches what's been defined in the rules.

This enables you to kick different pipelineruns off depending on the git branch.