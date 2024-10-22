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

	cssDracula := strings.Join(trimStringArray([]string{
		`/* Dracula Theme v1.2.5`,
		` *`,
		` * https://github.com/dracula/highlightjs`,
		` *`,
		` * Copyright 2016-present, All rights reserved`,
		` *`,
		` * Code licensed under the MIT license`,
		` *`,
		` * @author Denis Ciccale <dciccale@gmail.com>`,
		` * @author Zeno Rocha <hi@zenorocha.com>`,
		` */`,
		``,
		` .hljs {`,
		`	display: block;`,
		`	overflow-x: auto;`,
		`	padding: 0.5em;`,
		`	background: #282a36;`,
		`  }`,
		`  `,
		`  .hljs-built_in,`,
		`  .hljs-selector-tag,`,
		`  .hljs-section,`,
		`  .hljs-link {`,
		`	color: #8be9fd;`,
		`  }`,
		`  `,
		`  .hljs-keyword {`,
		`	color: #ff79c6;`,
		`  }`,
		`  `,
		`  .hljs,`,
		`  .hljs-subst {`,
		`	color: #f8f8f2;`,
		`  }`,
		`  `,
		`  .hljs-title,`,
		`  .hljs-attr,`,
		`  .hljs-meta-keyword {`,
		`	font-style: italic;`,
		`	color: #50fa7b;`,
		`  }`,
		`  `,
		`  .hljs-string,`,
		`  .hljs-meta,`,
		`  .hljs-name,`,
		`  .hljs-type,`,
		`  .hljs-symbol,`,
		`  .hljs-bullet,`,
		`  .hljs-addition,`,
		`  .hljs-variable,`,
		`  .hljs-template-tag,`,
		`  .hljs-template-variable {`,
		`	color: #f1fa8c;`,
		`  }`,
		`  `,
		`  .hljs-comment,`,
		`  .hljs-quote,`,
		`  .hljs-deletion {`,
		`	color: #6272a4;`,
		`  }`,
		`  `,
		`  .hljs-keyword,`,
		`  .hljs-selector-tag,`,
		`  .hljs-literal,`,
		`  .hljs-title,`,
		`  .hljs-section,`,
		`  .hljs-doctag,`,
		`  .hljs-type,`,
		`  .hljs-name,`,
		`  .hljs-strong {`,
		`	font-weight: bold;`,
		`  }`,
		`  `,
		`  .hljs-literal,`,
		`  .hljs-number {`,
		`	color: #bd93f9;`,
		`  }`,
		`  `,
		`  .hljs-emphasis {`,
		`	font-style: italic;`,
		`  }`,
	}), "")
	fmt.Fprint(w, strings.Join([]string{
		`<!doctype html>`,
		`<html lang="en">`,
		`<head>`,
		`<link href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css" rel="stylesheet">`,
		`<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.2/css/all.min.css" crossorigin="anonymous" referrerpolicy="no-referrer" />`,
		`<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/default.min.css">`,
		`<style>` + css + cssDracula + `</style>`,
		`<script src="https://code.jquery.com/jquery-3.5.1.slim.min.js"></script>`,
		`<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.bundle.min.js"></script>`,
		`<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>`,
		``,
		`<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/go.min.js"></script>`,
		`<script type="text/javascript"`,
		`src="https://cdn.mathjax.org/mathjax/latest/MathJax.js?config=TeX-AMS-MML_HTMLorMML">`,
		`</script>`,
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
		`hljs.highlightAll();`,
		`});`,
		`</script>`,
		`<script src="/static/script.js"></script>`,
		`<script type="module">`,
		`import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';`,
		`mermaid.initialize({ startOnLoad: true });`,
		`</script>`,
		`<script>`,
		`document.addEventListener('DOMContentLoaded', (event) => {`,
		`document.querySelectorAll('code.inline').forEach((el) => {`,
		`hljs.highlightElement(el);`,
		`});`,
		`});`,
		``,
		`</script>`,
		`<script type="text/x-mathjax-config">`,
		`MathJax.Hub.Config({`,
		`tex2jax: {`,
		`inlineMath: [ ['$','$'], ["\\(","\\)"] ],`,
		`processEscapes: true`,
		`}`,
		`});`,
		`</script>`,
		``,
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
func Menu(repos []repo.Repo) htmlwrapper.Elm {
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
						"href": "/repo/" + util.StringToSlug(repo.GetName()),
					},
					Contents: []htmlwrapper.Elm{
						htmlwrapper.Text(repo.GetName()),
					}}}})
	}

	return &htmlwrapper.MultiElm{Contents: contents}
}
