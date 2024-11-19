package controller

import (
	"bytes"
	"html/template"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
	"github.com/pjotrscholtze/go-bootstrap/cmd/go-bootstrap/bootstrap"
	"github.com/pjotrscholtze/go-bootstrap/cmd/go-bootstrap/builder"
	"github.com/pjotrscholtze/go-bootstrap/cmd/go-bootstrap/htmlwrapper"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/repo"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/view"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/websocketmanager"
	"github.com/pjotrscholtze/go-buildserver/models"
)

func renderUITemplate(path string, data any) bytes.Buffer {
	t, err := template.ParseFiles(path, "templates/parts/header.html", "templates/parts/footer.html")

	if err != nil {
		panic(err)
	}

	var w bytes.Buffer
	err = t.Execute(&w, data)

	if err != nil {
		panic(err)
	}
	return w
}

func wrapUITemplate(path string) bytes.Buffer {
	return renderUITemplate(path, struct{}{})
}

func RegisterUIController(buildRepo repo.PipelineRepo, buildQueue repo.JobQueue, wm websocketmanager.WebsocketManager) *mux.Router {
	router := mux.NewRouter()

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("../../static"))))
	router.HandleFunc("/repo/{repoName}/live", func(w http.ResponseWriter, r *http.Request) {
		// repoName := strings.Split(r.RequestURI[6:], "/")[0]
		repoName := mux.Vars(r)["repoName"]
		_ = repoName
		// @todo error handling!
		currentRepo := buildRepo.GetRepoBySlug(repoName)
		repoProperties := []htmlwrapper.Elm{
			&htmlwrapper.HTMLElm{
				Tag: "li",
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Name: ")}},
					htmlwrapper.Text(currentRepo.GetName()),
				},
			},
			&htmlwrapper.HTMLElm{
				Tag: "li",
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("In clean build mode: ")}},
					htmlwrapper.Text(map[bool]string{
						true:  "Yes",
						false: "No",
					}[currentRepo.ForceCleanBuild()]),
				},
			},
			&htmlwrapper.HTMLElm{
				Tag: "li",
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Repo path: ")}},
					htmlwrapper.Text(currentRepo.GetPath()),
				},
			},
			&htmlwrapper.HTMLElm{
				Tag: "li",
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("URL: ")}},
					htmlwrapper.Text(currentRepo.GetURL()),
				},
			},
			&htmlwrapper.HTMLElm{
				Tag: "li",
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Build script: ")}},
					htmlwrapper.Text(currentRepo.GetBuildScript()),
				},
			},
		}
		for _, trigger := range currentRepo.GetTriggers() {
			text := trigger.Kind
			if trigger.Schedule != "" {
				text = text + " (" + trigger.Schedule + ")"
			}
			repoProperties = append(repoProperties, &htmlwrapper.HTMLElm{
				Tag: "li",
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Trigger: ")}},
					htmlwrapper.Text(text),
				},
			})
		}

		meta := htmlwrapper.HTMLElm{
			Tag: "ul",
			Attrs: map[string]string{
				"class":     "properties",
				"data-repo": currentRepo.GetName(),
			},
			Contents: repoProperties,
		}
		lastBuilds := currentRepo.GetLastNBuildResults(1)

		lastBuildStatus := ""
		lastBuildReason := ""
		lastBuildStarttime := ""
		lines := []repo.BuildResultLine{{
			Line: "",
			Pipe: "",
			Time: strfmt.UnixZero,
		}}
		if len(lastBuilds) == 1 {
			lastBuild := lastBuilds[0]
			lastBuildStatus = string(lastBuild.Status)
			lastBuildReason = lastBuild.Reason
			lastBuildStarttime = lastBuild.Starttime.Format(time.DateTime)
			lines = lastBuild.Lines
		}
		tb := builder.NewTableBuilder[repo.BuildResultLine](lines)
		tb.GetTypeMapperConv().RegisterCustomFieldMapping("pipe", func(fieldName string, refStruct interface{}) htmlwrapper.Elm {
			line := refStruct.(repo.BuildResultLine)
			return htmlwrapper.Text(line.Pipe)
		})
		tb.GetTypeMapperConv().RegisterCustomFieldMapping("time", func(fieldName string, refStruct interface{}) htmlwrapper.Elm {
			line := refStruct.(repo.BuildResultLine)
			return htmlwrapper.Text(strfmt.DateTime(line.Time).String())
		})

		tb.GetTableMapping().MoveToIndex("Pipe", 0).MoveToIndex("Time", 1).RemoveByFieldName("pipe")

		tb.SetSize(bootstrap.BsTablSizeSmall)
		buildLines := tb.AsElm()
		buildMeta := &htmlwrapper.HTMLElm{
			Tag: "ul",
			Attrs: map[string]string{
				"class": "properties",
			},
			Contents: []htmlwrapper.Elm{
				&htmlwrapper.HTMLElm{
					Tag: "li",
					Attrs: map[string]string{
						"id": "build-status",
					},
					Contents: []htmlwrapper.Elm{
						&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Build status: ")}},
						&htmlwrapper.HTMLElm{Tag: "span", Contents: []htmlwrapper.Elm{htmlwrapper.Text(lastBuildStatus)}},
					},
				},
				&htmlwrapper.HTMLElm{
					Tag: "li",
					Attrs: map[string]string{
						"id": "build-reason",
					},
					Contents: []htmlwrapper.Elm{
						&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Reason: ")}},
						&htmlwrapper.HTMLElm{Tag: "span", Contents: []htmlwrapper.Elm{htmlwrapper.Text(lastBuildReason)}},
					},
				},
				&htmlwrapper.HTMLElm{
					Tag: "li",
					Attrs: map[string]string{
						"id": "build-start-time",
					},
					Contents: []htmlwrapper.Elm{
						&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Start time: ")}},
						&htmlwrapper.HTMLElm{Tag: "span", Contents: []htmlwrapper.Elm{htmlwrapper.Text(lastBuildStarttime)}},
					},
				},
			},
		}
		// repo.GetTriggers()
		// bb := wrapUITemplate("templates/repo_details.html")
		// w.Write(bb.Bytes())
		// buildRepo.GetRepoByName()
		requestURI := r.RequestURI
		title, elm := view.Page(
			requestURI,
			"GoBuildServer",
			buildRepo.List(),
			&htmlwrapper.MultiElm{
				Contents: []htmlwrapper.Elm{
					&meta,
					&htmlwrapper.HTMLElm{
						Tag: "div",
						Attrs: map[string]string{
							"class": "build-result",
						},
						Contents: []htmlwrapper.Elm{
							buildMeta,
							buildLines,
						},
					},
					&htmlwrapper.HTMLElm{
						Tag: "div",
						Attrs: map[string]string{
							"class": "scroll-options",
						},
						Contents: []htmlwrapper.Elm{
							&htmlwrapper.HTMLElm{Tag: "input", Attrs: map[string]string{
								"type":    "checkbox",
								"name":    "auto-scroll",
								"id":      "auto-scroll",
								"checked": "checked",
							}},
							&htmlwrapper.HTMLElm{Tag: "label", Attrs: map[string]string{
								"for": "auto-scroll",
							}, Contents: []htmlwrapper.Elm{htmlwrapper.Text("Scroll with output")}},
						},
					},
				},
			},
		)

		html, _ := elm.AsHTML()
		view.Wrap(w, title, html)
	})

	router.HandleFunc("/build/{buildId}", func(w http.ResponseWriter, r *http.Request) {
		// repoName := strings.Split(r.RequestURI[6:], "/")[0]
		buildIdStr := mux.Vars(r)["buildId"]
		if !strings.HasPrefix(buildIdStr, "build-") {
			// @todo error
		}
		buildIdStrClean := buildIdStr[6:]
		buildId, err := strconv.ParseInt(buildIdStrClean, 10, 64)
		if err != nil {
			// @todo error
		}

		job := buildQueue.GetJobById(buildId)
		if job == nil {
			// @todo error
		}
		jobBuilder := buildRepo.GetRepoByName(job.RepoName)
		buildResult := jobBuilder.GetBuildResultForJobID(job)
		if buildResult == nil {
			buildResult = &repo.BuildResult{
				PipelineName: job.RepoName,
				Lines: []repo.BuildResultLine{
					{
						Line: "",
						Pipe: "",
						Time: strfmt.UnixZero,
					},
				},
				Reason:           job.BuildReason,
				Starttime:        time.Time(job.QueueTime),
				Status:           repo.ResultStatus(job.Status),
				Websocketmanager: nil,
				Job:              job,
			}
		}

		buildMeta := &htmlwrapper.HTMLElm{
			Tag: "ul",
			Attrs: map[string]string{
				"class": "properties",
			},
			Contents: []htmlwrapper.Elm{
				&htmlwrapper.HTMLElm{
					Tag: "li",
					Attrs: map[string]string{
						"id": "build-status",
					},
					Contents: []htmlwrapper.Elm{
						&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Build status: ")}},
						&htmlwrapper.HTMLElm{Tag: "span", Contents: []htmlwrapper.Elm{htmlwrapper.Text(job.Status)}},
					},
				},
				&htmlwrapper.HTMLElm{
					Tag: "li",
					Attrs: map[string]string{
						"id": "build-reason",
					},
					Contents: []htmlwrapper.Elm{
						&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Reason: ")}},
						&htmlwrapper.HTMLElm{Tag: "span", Contents: []htmlwrapper.Elm{htmlwrapper.Text(job.BuildReason)}},
					},
				},
				&htmlwrapper.HTMLElm{
					Tag: "li",
					Attrs: map[string]string{
						"id": "queue-time",
					},
					Contents: []htmlwrapper.Elm{
						&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Queue time: ")}},
						&htmlwrapper.HTMLElm{Tag: "span", Contents: []htmlwrapper.Elm{htmlwrapper.Text(job.QueueTime.String())}},
					},
				},
				&htmlwrapper.HTMLElm{
					Tag: "li",
					Attrs: map[string]string{
						"id": "origin",
					},
					Contents: []htmlwrapper.Elm{
						&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Origin: ")}},
						&htmlwrapper.HTMLElm{Tag: "span", Contents: []htmlwrapper.Elm{htmlwrapper.Text(job.Origin)}},
					},
				},
				&htmlwrapper.HTMLElm{
					Tag: "li",
					Attrs: map[string]string{
						"id": "origin",
					},
					Contents: []htmlwrapper.Elm{
						&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Repo name: ")}},
						&htmlwrapper.HTMLElm{Tag: "span", Contents: []htmlwrapper.Elm{htmlwrapper.Text(job.RepoName)}},
					},
				},
				&htmlwrapper.HTMLElm{
					Tag: "li",
					Attrs: map[string]string{
						"id": "repo-url",
					},
					Contents: []htmlwrapper.Elm{
						&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Repo URL: ")}},
						&htmlwrapper.HTMLElm{Tag: "span", Contents: []htmlwrapper.Elm{htmlwrapper.Text(jobBuilder.GetURL())}},
					},
				},
				&htmlwrapper.HTMLElm{
					Tag: "li",
					Contents: []htmlwrapper.Elm{
						&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Repo path: ")}},
						htmlwrapper.Text(jobBuilder.GetPath()),
					},
				},
				&htmlwrapper.HTMLElm{
					Tag: "li",
					Contents: []htmlwrapper.Elm{
						&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Starttime: ")}},
						htmlwrapper.Text(buildResult.Starttime.String()),
					},
				},
				&htmlwrapper.HTMLElm{
					Tag: "li",
					Contents: []htmlwrapper.Elm{
						&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("build id: ")}},
						&htmlwrapper.HTMLElm{Tag: "span", Attrs: map[string]string{"id": "build-id", "data-build-id": strconv.FormatInt(buildId, 10)}, Contents: []htmlwrapper.Elm{htmlwrapper.Text(strconv.FormatInt(buildId, 10))}},
					},
				},
			},
		}

		tb := builder.NewTableBuilder[repo.BuildResultLine](buildResult.Lines)
		tb.GetTypeMapperConv().RegisterCustomFieldMapping("pipe", func(fieldName string, refStruct interface{}) htmlwrapper.Elm {
			line := refStruct.(repo.BuildResultLine)
			return htmlwrapper.Text(line.Pipe)
		})
		tb.GetTypeMapperConv().RegisterCustomFieldMapping("time", func(fieldName string, refStruct interface{}) htmlwrapper.Elm {
			line := refStruct.(repo.BuildResultLine)
			return htmlwrapper.Text(strfmt.DateTime(line.Time).String())
		})

		tb.GetTableMapping().MoveToIndex("Pipe", 0).MoveToIndex("Time", 1).RemoveByFieldName("pipe")

		tb.SetSize(bootstrap.BsTablSizeSmall)
		buildLines := tb.AsElm()

		// _ = buildId
		requestURI := r.RequestURI
		title, elm := view.Page(
			requestURI,
			"GoBuildServer",
			buildRepo.List(),
			&htmlwrapper.MultiElm{
				Contents: []htmlwrapper.Elm{
					// &meta,
					&htmlwrapper.HTMLElm{
						Tag: "div",
						Attrs: map[string]string{
							"class": "build-result",
						},
						Contents: []htmlwrapper.Elm{
							buildMeta,
							buildLines,
						},
					},
					&htmlwrapper.HTMLElm{
						Tag: "div",
						Attrs: map[string]string{
							"class": "scroll-options",
						},
						Contents: []htmlwrapper.Elm{
							&htmlwrapper.HTMLElm{Tag: "input", Attrs: map[string]string{
								"type":    "checkbox",
								"name":    "auto-scroll",
								"id":      "auto-scroll",
								"checked": "checked",
							}},
							&htmlwrapper.HTMLElm{Tag: "label", Attrs: map[string]string{
								"for": "auto-scroll",
							}, Contents: []htmlwrapper.Elm{htmlwrapper.Text("Scroll with output")}},
						},
					},
				},
			},
		)

		html, _ := elm.AsHTML()
		view.Wrap(w, title, html)
	})

	router.HandleFunc("/repo/{repoName}", func(w http.ResponseWriter, r *http.Request) {
		// repoName := strings.Split(r.RequestURI[6:], "/")[0]
		pageNumberStr := r.URL.Query().Get("pageNumber")
		numberOfResultsPerPageStr := r.URL.Query().Get("numberOfResultsPerPage")
		if pageNumberStr == "" {
			pageNumberStr = "1"
		}
		if numberOfResultsPerPageStr == "" {
			numberOfResultsPerPageStr = "10"
		}

		pageNumber, _ := strconv.ParseInt(pageNumberStr, 10, 64)
		numberOfResultsPerPage, _ := strconv.ParseInt(numberOfResultsPerPageStr, 10, 64)
		if pageNumber < 1 {
			pageNumber = 1
		}
		if numberOfResultsPerPage < 1 {
			numberOfResultsPerPage = 10
		}
		repoName := mux.Vars(r)["repoName"]
		_ = repoName
		// @todo error handling!
		currentRepo := buildRepo.GetRepoBySlug(repoName)
		repoProperties := []htmlwrapper.Elm{
			&htmlwrapper.HTMLElm{
				Tag: "li",
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Name: ")}},
					htmlwrapper.Text(currentRepo.GetName()),
				},
			},
			&htmlwrapper.HTMLElm{
				Tag: "li",
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("In clean build mode: ")}},
					htmlwrapper.Text(map[bool]string{
						true:  "Yes",
						false: "No",
					}[currentRepo.ForceCleanBuild()]),
				},
			},
			&htmlwrapper.HTMLElm{
				Tag: "li",
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Repo path: ")}},
					htmlwrapper.Text(currentRepo.GetPath()),
				},
			},
			&htmlwrapper.HTMLElm{
				Tag: "li",
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("URL: ")}},
					htmlwrapper.Text(currentRepo.GetURL()),
				},
			},
			&htmlwrapper.HTMLElm{
				Tag: "li",
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Build script: ")}},
					htmlwrapper.Text(currentRepo.GetBuildScript()),
				},
			},
		}
		for _, trigger := range currentRepo.GetTriggers() {
			text := trigger.Kind
			if trigger.Schedule != "" {
				text = text + " (" + trigger.Schedule + ")"
			}
			repoProperties = append(repoProperties, &htmlwrapper.HTMLElm{
				Tag: "li",
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{Tag: "strong", Contents: []htmlwrapper.Elm{htmlwrapper.Text("Trigger: ")}},
					htmlwrapper.Text(text),
				},
			})
		}
		repoProperties = append(repoProperties, &htmlwrapper.HTMLElm{
			Tag: "li",
			Contents: []htmlwrapper.Elm{
				&htmlwrapper.HTMLElm{Tag: "a",
					Attrs: map[string]string{
						"href": "/repo/" + repoName + "/live",
					},
					Contents: []htmlwrapper.Elm{
						htmlwrapper.Text("Live build"),
					}},
			},
		})

		meta := htmlwrapper.HTMLElm{
			Tag: "ul",
			Attrs: map[string]string{
				"class":     "properties",
				"data-repo": currentRepo.GetName(),
			},
			Contents: repoProperties,
		}
		lastBuilds := []models.Job{}
		for _, br := range buildQueue.ListAllJobsOfPipeline(currentRepo.GetName()) {
			lastBuilds = append([]models.Job{*br}, lastBuilds...)
		}
		start := numberOfResultsPerPage * (pageNumber - 1)
		if start > int64(len(lastBuilds)) {
			start = int64(len(lastBuilds))
		}
		end := numberOfResultsPerPage * pageNumber
		if end > int64(len(lastBuilds)) {
			end = int64(len(lastBuilds))
		}
		buildsForTable := lastBuilds[start:end]
		noRows := false
		if len(buildsForTable) == 0 {
			noRows = true
			buildsForTable = []models.Job{models.Job{
				BuildReason: "",
				ID:          -1,
				Origin:      "",
				QueueTime:   strfmt.DateTime{},
				RepoName:    repoName,
				Status:      "",
			}}
		}
		tb := builder.NewAdvancedTableBuilder[models.Job](buildsForTable)
		tb.GetPaginationBuilder().SetCurrentPage(int(pageNumber)).
			SetCurrentResultsPerPage(int(numberOfResultsPerPage)).
			SetResultCount(len(lastBuilds)).
			SetResultsPerPageOptions([]int{10, 25, 50, 100})
		tb.GetTableBuilder().GetTypeMapperConv().RegisterCustomFieldMapping("QueueTime", func(fieldName string, refStruct interface{}) htmlwrapper.Elm {
			line := refStruct.(models.Job)
			return htmlwrapper.Text(strfmt.DateTime(line.QueueTime).String())
		})
		fnInt64, _ := tb.GetTableBuilder().GetTypeMapperConv().GetMapping("int64")
		tb.GetTableBuilder().GetTypeMapperConv().RegisterMapping("int64", func(structField reflect.StructField, value reflect.Value, refStruct interface{}) htmlwrapper.Elm {
			if structField.Name == "ID" {
				line := refStruct.(models.Job)
				return &htmlwrapper.HTMLElm{
					Tag: "a",
					Attrs: map[string]string{
						"href": "/build/" + ("build-" + strconv.FormatInt(line.ID, 10)),
					},
					Contents: []htmlwrapper.Elm{htmlwrapper.Text("build-" + strconv.FormatInt(line.ID, 10))},
				}
			}
			return fnInt64(structField, value, refStruct)
		})
		tableElm := tb.AsElm().(*htmlwrapper.HTMLElm)
		tableElm.Attrs = nil
		if noRows {
			tableElm.Contents[1].(*htmlwrapper.HTMLElm).Contents[1].(*htmlwrapper.HTMLElm).Contents = nil
		}

		requestURI := r.RequestURI
		title, elm := view.Page(
			requestURI,
			"GoBuildServer",
			buildRepo.List(),
			&htmlwrapper.MultiElm{
				Contents: []htmlwrapper.Elm{
					&meta,
					&htmlwrapper.HTMLElm{
						Tag: "div",
						Attrs: map[string]string{
							"class": "build-result",
						},
						Contents: []htmlwrapper.Elm{
							tableElm,
							// buildMeta,
							// buildLines,
						},
					},
					&htmlwrapper.HTMLElm{
						Tag: "div",
						Attrs: map[string]string{
							"class": "scroll-options",
						},
						Contents: []htmlwrapper.Elm{
							&htmlwrapper.HTMLElm{Tag: "input", Attrs: map[string]string{
								"type":    "checkbox",
								"name":    "auto-scroll",
								"id":      "auto-scroll",
								"checked": "checked",
							}},
							&htmlwrapper.HTMLElm{Tag: "label", Attrs: map[string]string{
								"for": "auto-scroll",
							}, Contents: []htmlwrapper.Elm{htmlwrapper.Text("Scroll with output")}},
						},
					},
				},
			},
		)

		html, _ := elm.AsHTML()
		view.Wrap(w, title, html)
	}) // <div class="scroll-options card">
	//   <div class="card-header">
	//     <input type="checkbox" name="auto-scroll" id="auto-scroll" checked>
	//     <label for="auto-scroll">Scroll with output</label>
	//   </div>
	// </div>
	router.PathPrefix("/ws/").HandlerFunc(wm.Setup)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		jobs := []models.Job{}
		for _, j := range buildQueue.List() {
			jobs = append(jobs, *j)
		}
		if len(jobs) == 0 {
			jobs = append(jobs, models.Job{
				BuildReason: "-",
				Origin:      "-",
				RepoName:    "-",
				QueueTime:   strfmt.NewDateTime(),
			})
		}
		tb := builder.NewTableBuilder[models.Job](jobs)
		requestURI := r.RequestURI
		title, elm := view.Page(
			requestURI,
			"GoBuildServer",
			buildRepo.List(),
			&htmlwrapper.MultiElm{
				Contents: []htmlwrapper.Elm{
					tb.AsElm(),
				},
			},
		)

		html, _ := elm.AsHTML()
		view.Wrap(w, title, html)
	})
	return router
}
