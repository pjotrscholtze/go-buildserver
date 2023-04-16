
Example yaml:
```
MaxHistoryInMemory: 1
HTTPPort: 3000
HTTPHost: "0.0.0.0"
WorkspaceDirectory: /tmp/go-buildserver
Repos:
- URL: git@github.com:pjotrscholtze/go-buildserver.git
  Name: Go-Buildserver
  BuildScript: ./build.sh
  SSHKeyLocation: /data/.ssh/github.com
  ForceCleanBuild: True
  Triggers:
  - Kind: Cron
    Schedule: "*/15 * * * * *"
  - Kind: WebHook
```

Field description
| Field                       | Description                                                                                 |
|-----------------------------|---------------------------------------------------------------------------------------------|
| MaxHistoryInMemory          | Maximum number of builds to keep in memory.                                                 |
| HTTPPort                    | HTTP port to listen on.                                                                     |
| HTTPHost                    | Interface to listen on for HTTP.                                                            |
| WorkspaceDirectory          | Directory to clone, and run builds in.                                                      |
| Repos                       | List of repos to be able to build.                                                          |
| Repos[].URL                 | URL of the repo to clone at build time.                                                     |
| Repos[].Name                | Name of the repo/ job                                                                       |
| Repos[].BuildScript         | Build script to run at build time.                                                          |
| Repos[].SSHKeyLocation      | Location of the SSH key which can be used to clone the repo.                                |
| Repos[].ForceCleanBuild     | Force a clean clone of the repository, s.t. there are no residuals.                         |
| Repos[].Triggers            | List of triggers that can start the job                                                     |
| Repos[].Triggers[].Kind     | Either `Cron` or `WebHook`, allowing for a periodic build and a webhook to trigger a build. |
| Repos[].Triggers[].Schedule | Field is only used when kind is `Cron`. Cron format of when to run the job periodically.    |
