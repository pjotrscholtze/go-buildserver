package controller

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
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

func RegisterUIController(buildRepo repo.PipelineRepo, buildQueue repo.JobQueue, wm websocketmanager.WebsocketManager) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("../../static"))))
	mux.HandleFunc("/repo/", func(w http.ResponseWriter, r *http.Request) {
		repoName := strings.Split(r.RequestURI[6:], "/")[0]
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
		lines := []repo.BuildResultLine{}
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
	// <div class="scroll-options card">
	//   <div class="card-header">
	//     <input type="checkbox" name="auto-scroll" id="auto-scroll" checked>
	//     <label for="auto-scroll">Scroll with output</label>
	//   </div>
	// </div>
	mux.HandleFunc("/ws/", wm.Setup)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
	return mux
}
