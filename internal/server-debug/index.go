package serverdebug

import (
	"html/template"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type page struct {
	Path        string
	Description string
}

type indexPage struct {
	pages []page
}

func newIndexPage() *indexPage {
	return &indexPage{}
}

func (i *indexPage) addPage(path string, description string) {
	i.pages = append(i.pages, page{Path: path, Description: description})
}

func (i indexPage) handler(eCtx echo.Context) error {
	currentLevel := zap.L().Level().String()

	return template.Must(template.New("index").Parse(`<html>
	<title>Chat Service Debug</title>
<body>
	<h2>Chat Service Debug</h2>
	<ul>
	{{range .Pages}}
		<li><a href="{{.Path}}">{{.Description}}</a></li>
	{{end}}
	</ul>

	<h2>Log Level</h2>
	<form onSubmit="putLogLevel(); return false;">
		<select id="log-level-select">
			<option value="debug" {{if eq .LogLevel "debug"}}selected{{end}}>Debug</option>
			<option value="info" {{if eq .LogLevel "info"}}selected{{end}}>Info</option>
			<option value="warn" {{if eq .LogLevel "warn"}}selected{{end}}>Warn</option>
			<option value="error" {{if eq .LogLevel "error"}}selected{{end}}>Error</option>
			<option value="fatal" {{if eq .LogLevel "fatal"}}selected{{end}}>Fatal</option>
			<option value="panic" {{if eq .LogLevel "panic"}}selected{{end}}>Panic</option>
		</select>
		<input type="submit" value="Change"></input>
	</form>
	
	<script>
		function putLogLevel() {
			const req = new XMLHttpRequest();
			req.open('PUT', '/log/level', true);
			req.setRequestHeader('Content-Type', 'application/json');
			req.onload = function() { window.location.reload(); };
			req.send(JSON.stringify({ level: document.getElementById('log-level-select').value }));
		};
	</script>
</body>
</html>
`)).Execute(eCtx.Response(), struct {
		Pages    []page
		LogLevel string
	}{
		Pages:    i.pages,
		LogLevel: currentLevel,
	})
}
