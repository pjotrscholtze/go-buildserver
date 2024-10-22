package view

import (
	"path/filepath"

	"github.com/pjotrscholtze/go-bootstrap/cmd/go-bootstrap/htmlwrapper"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/repo"
)

func Page(currentPath, basePath string, repos []repo.Repo, content htmlwrapper.Elm) (string, htmlwrapper.Elm) {
	menu := Menu(repos)
	base := filepath.Base(currentPath)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}

	title := StringToName(base)
	return title, &htmlwrapper.HTMLElm{
		Tag: "div",
		Attrs: map[string]string{
			"class": "wrapper",
		},
		Contents: []htmlwrapper.Elm{
			&htmlwrapper.HTMLElm{
				Tag: "nav",
				Attrs: map[string]string{
					"id": "sidebar",
				},
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{
						Tag: "div",
						Attrs: map[string]string{
							"class": "sidebar-header",
						},
						Contents: []htmlwrapper.Elm{
							&htmlwrapper.HTMLElm{
								Tag:   "h3",
								Attrs: map[string]string{},
								Contents: []htmlwrapper.Elm{
									htmlwrapper.Text(StringToName(filepath.Base(basePath))),
								},
							},
						},
					},
					&htmlwrapper.HTMLElm{
						Tag: "div",
						Attrs: map[string]string{
							"class": "list-unstyled components",
						},
						Contents: []htmlwrapper.Elm{
							menu,
						},
					},
				},
			},
			&htmlwrapper.HTMLElm{
				Tag: "div",
				Attrs: map[string]string{
					"id": "content",
				},
				Contents: []htmlwrapper.Elm{
					&htmlwrapper.HTMLElm{
						Tag: "h1",
						Contents: []htmlwrapper.Elm{
							htmlwrapper.Text(title),
						},
					},
					content,
				},
			},
		},
	}
}
