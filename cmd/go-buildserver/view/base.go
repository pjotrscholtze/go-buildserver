package view

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pjotrscholtze/go-bootstrap/cmd/go-bootstrap/htmlwrapper"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/repo"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/util"
)

func Error(errorMsg string, w http.ResponseWriter) {
	fmt.Fprint(w, strings.Join([]string{
		`<!doctype html>`,
		`<html lang="en">`,
		`<head>`,
		`<link href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css" rel="stylesheet">`,
		`<meta charset="utf-8">`,
		`<meta name="viewport" content="width=device-width, initial-scale=1">`,
		`<title>Error while rending the page!</title>`,
		`</head>`,
		`<body>`,
		`<h1>Error while rending the page!</h1>`,
		`<pre>` + errorMsg + `</pre>`,
		`</body>`,
		`</html>`,
	}, ""))

}

func NotFound(w http.ResponseWriter) {
	fmt.Fprint(w, strings.Join([]string{
		`<!doctype html>`,
		`<html lang="en">`,
		`<head>`,
		`<link href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css" rel="stylesheet">`,
		`<meta charset="utf-8">`,
		`<meta name="viewport" content="width=device-width, initial-scale=1">`,
		`<title>Page not found!</title>`,
		`</head>`,
		`<body>`,
		`<h1>Page not found!</h1>`,
		`</body>`,
		`</html>`,
	}, ""))
}

func Wrap(w http.ResponseWriter, title, html string) {
	css := strings.Join([]string{
		`body {`,
		`font-family: 'Poppins', sans-serif;`,
		`background: #f0f0f1;`,
		`color:#3c434a;`,
		`}`,
		`a {`,
		`text-decoration: none;`,
		`}`,
		`.wrapper {`,
		`display: flex;`,
		`}`,
		`#sidebar {`,
		`min-width: 300px;`,
		`max-width: 300px;`,
		`background: #282a36;`,
		// `/* #3c434a; */`,
		`color: #f0f0f1;`,
		`transition: all 0.3s;`,
		`}`,
		`#sidebar li a:hover {`,
		`color:  #72aee6;`,
		`}`,
		`#sidebar li a {`,
		`padding: 0.5em;`,
		`display: block;`,
		`color: #f0f0f1;`,
		`}`,
		`#sidebar li ul {`,
		`padding: 0 0.5em;`,
		`}`,
		`#sidebar.active {`,
		`margin-left: -250px;`,
		`}`,
		`#sidebarToggle {`,
		`transition: all 0.3s;`,
		`}`,
		`#sidebarToggle.active {`,
		`transform: rotate(90deg);`,
		`}`,
		`#content {`,
		`width: 100%;`,
		`padding: 20px;`,
		`min-height: 100vh;`,
		`transition: all 0.3s;`,
		`}`,
		// `#content.active {`,
		// `/* margin-left: 250px; */`,
		// `}`,
		`img {`,
		`display: block;`,
		`}`,
		`pre code{`,
		`display: block;`,
		`overflow-x: auto;`,
		`padding: 1em;`,
		`background: #282a36;`,
		`}`,
		`pre,code {`,
		`-webkit-border-radius: 0.5em;`,
		`border-radius: 0.5em;`,
		`}`,
		`code.inline {`,
		`-webkit-border-radius: 0.25em;`,
		`border-radius: 0.25em;`,
		`margin-bottom: -0.3em;`,
		`top: 0.1em;`,
		`position: relative;`,
		``,
		`display: inline-block;`,
		`padding:0 0.3em;`,
		`font-size:0.9em;`,
		`}`,
		`blockquote, dl {`,
		`background: #FFF;`,
		`padding: 0.75em 0 0.75em 1em;`,
		`-webkit-border-radius: 0.25em;`,
		`border-radius: 0.25em;`,
		`border-left: 0.3em solid #555;`,
		`}`,
		``,
		`.footnote-definition{`,
		`background: #FFF;`,
		`padding: 0.75em 0 0.75em 1em;`,
		`-webkit-border-radius: 0.25em;`,
		`border-radius: 0.25em;`,
		`border-left: 0.3em solid #555;  `,
		`}`,
		`.sidebar-header h3 {`,
		`padding: 0.3em`,
		`}`,
		`ul.properties{display: grid;grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap:0 1em; padding: 0;border: 0.1em solid #DDD; background: #F5F5F5;}`,
		`.scroll-options {width: 13em;position: fixed;right: 1em;bottom: 1em; padding: 1em;border: 0.1em solid #DDD; background: #F5F5F5;}`,
		`.scroll-options input {margin-right: 0.6em;}`,
		`ul.properties li { list-style: none; padding: 0.5em;}`,
		`.build-result table.table{margin-bottom:7rem}`,
	}, "")

	fmt.Fprint(w, strings.Join([]string{
		`<!doctype html>`,
		`<html lang="en">`,
		`<head>`,
		// `<link href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css" rel="stylesheet">`,
		`<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet" crossorigin="anonymous">`,
		`<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.2/css/all.min.css" crossorigin="anonymous" referrerpolicy="no-referrer" />`,
		`<style>` + css + `</style>`,
		`<script src="https://code.jquery.com/jquery-3.5.1.slim.min.js"></script>`,
		``,
		`<meta charset="utf-8">`,
		`<meta name="viewport" content="width=device-width, initial-scale=1">`,
		`<title>` + title + `</title>`,
		`</head>`,
		`<body>`,
		`` + html + ``,
		`<script>`,
		`$(document).ready(function () {`,
		`$('#sidebarCollapse').on('click', function () {`,
		`$('#sidebar').toggleClass('active');`,
		`$('#content').toggleClass('active');`,
		`});`,
		`});`,
		`</script>`,
		`<script src="/static/script.js"></script>`,
		`<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/js/bootstrap.bundle.min.js" crossorigin="anonymous"></script>`,
		// `<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.bundle.min.js"></script>`,
		`</body>`,
		`</html>`,
	}, ""))

}
func trimStringArray(in []string) []string {
	out := make([]string, len(in))
	for i := range in {
		out[i] = strings.Trim(in[i], " \t")
	}
	return out
}
func StringToName(in string) string {
	id := in
	for _, old := range []string{"#"} {
		id = strings.ReplaceAll(id, old, "")
	}

	id = strings.Trim(id, " \t")
	for _, old := range []string{" ", "_", ".", "/"} {
		id = strings.ReplaceAll(id, old, " ")
	}

	if len(id) > 0 {
		id = strings.ToUpper(id[:1]) + id[1:]
	}

	return id
}
func Menu(repos []repo.Pipeline) htmlwrapper.Elm {
	contents := []htmlwrapper.Elm{}
	contents = append(contents,
		&htmlwrapper.HTMLElm{
			Tag: "li",

			Contents: []htmlwrapper.Elm{
				&htmlwrapper.HTMLElm{
					Tag: "a",
					Attrs: map[string]string{
						"href": "/",
					},
					Contents: []htmlwrapper.Elm{
						htmlwrapper.Text(StringToName("Queue overview")),
					}}}})
	contents = append(contents,
		&htmlwrapper.HTMLElm{
			Tag: "hr",
		})
	for _, repo := range repos {
		contents = append(contents, &htmlwrapper.HTMLElm{
			Tag: "li",

			Contents: []htmlwrapper.Elm{
				&htmlwrapper.HTMLElm{
					Tag: "a",
					Attrs: map[string]string{
						"href": "/repo/" + util.StringToSlug(repo.GetPipelineConfig().Name),
					},
					Contents: []htmlwrapper.Elm{
						htmlwrapper.Text(repo.GetPipelineConfig().Name),
					}}}})
	}

	return &htmlwrapper.MultiElm{Contents: contents}
}
