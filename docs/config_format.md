
Example yaml:
```
MaxHistoryInMemory: 1
HTTPPort: 3000
HTTPHost: "0.0.0.0"
WorkspaceDirectory: /tmp/go-buildserver
Repos:
- Path: ../../
  Name: Path_based_example
  BuildScript: ./build2.sh
  Triggers:
  - Kind: Cron
    Schedule: "*/10 * * * * *"
  - Kind: WebHook
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
| Repos[].Path                | Path of the script to run (Path based job). (NA for Git based jobs)                         |
| Repos[].URL                 | URL of the repo to clone at build time. (NA for path based jobs)                            |
| Repos[].Name                | Name of the repo/ job                                                                       |
| Repos[].BuildScript         | Build script to run at build time.                                                          |
| Repos[].SSHKeyLocation      | Location of the SSH key which can be used to clone the repo. (NA for path based jobs)       |
| Repos[].ForceCleanBuild     | Force a clean clone of the repository, s.t. there are no residuals. (NA for path based jobs)|
| Repos[].Triggers            | List of triggers that can start the job                                                     |
| Repos[].Triggers[].Kind     | Either `Cron` or `WebHook`, allowing for a periodic build and a webhook to trigger a build. |
| Repos[].Triggers[].Schedule | Field is only used when kind is `Cron`. Cron format of when to run the job periodically.    |
