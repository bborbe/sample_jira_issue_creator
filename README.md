# Jira Issue Creator

Simple script to create a Jira Issue


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
