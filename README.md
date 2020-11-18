# Jira Issue Creator

Simple script to create a Jira Issue

## Create Issue

```bash
go run main.go \
-url=https://jira-test.apps.seibert-media.net \
-username=$(teamvault-username --teamvault-config ~/.teamvault-sm.json --teamvault-key=gXMy4m) \
-password=$(teamvault-password --teamvault-config ~/.teamvault-sm.json --teamvault-key=gXMy4m) \
-project-key=BRO \
-issue-type=Task \
-summary=test-summary \
-description=test-description \
-v=2
```

## Create Subtask

```bash
go run main.go \
-url=https://jira-test.apps.seibert-media.net \
-username=$(teamvault-username --teamvault-config ~/.teamvault-sm.json --teamvault-key=gXMy4m) \
-password=$(teamvault-password --teamvault-config ~/.teamvault-sm.json --teamvault-key=gXMy4m) \
-project-key=BRO \
-issue-type=Subtask \
-summary=test-sub-summary \
-description=test-sub-description \
-parent-issue-key=BRO-2867 \
-v=2
```
