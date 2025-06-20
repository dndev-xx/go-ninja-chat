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
<head>
	<title>Chat Service Debug</title>
	<style>
		:root {
			--primary-color: #3498db;
			--secondary-color: #2980b9;
			--background-color: #f8f9fa;
			--card-bg: #ffffff;
			--text-color: #333333;
			--border-color: #e0e0e0;
		}
		
		body {
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			line-height: 1.6;
			color: var(--text-color);
			background-color: var(--background-color);
			margin: 0;
			padding: 20px;
		}
		
		.container {
			max-width: 1000px;
			margin: 0 auto;
			padding: 20px;
		}
		
		.header {
			background-color: var(--primary-color);
			color: white;
			padding: 20px;
			border-radius: 5px;
			margin-bottom: 20px;
			box-shadow: 0 2px 5px rgba(0,0,0,0.1);
		}
		
		.card {
			background-color: var(--card-bg);
			border-radius: 5px;
			padding: 20px;
			margin-bottom: 20px;
			box-shadow: 0 2px 5px rgba(0,0,0,0.1);
		}
		
		h2 {
			color: var(--primary-color);
			margin-top: 0;
			border-bottom: 1px solid var(--border-color);
			padding-bottom: 10px;
		}
		
		ul {
			list-style-type: none;
			padding: 0;
		}
		
		li {
			padding: 8px 0;
			border-bottom: 1px solid var(--border-color);
		}
		
		li:last-child {
			border-bottom: none;
		}
		
		a {
			color: var(--primary-color);
			text-decoration: none;
			transition: color 0.3s;
		}
		
		a:hover {
			color: var(--secondary-color);
			text-decoration: underline;
		}
		
		form {
			display: flex;
			align-items: center;
			gap: 10px;
		}
		
		select {
			padding: 8px;
			border: 1px solid var(--border-color);
			border-radius: 4px;
			background-color: var(--card-bg);
		}
		
		input[type="submit"] {
			background-color: var(--primary-color);
			color: white;
			border: none;
			padding: 8px 16px;
			border-radius: 4px;
			cursor: pointer;
			transition: background-color 0.3s;
		}
		
		input[type="submit"]:hover {
			background-color: var(--secondary-color);
		}
		
		.log-level-container {
			display: flex;
			align-items: center;
			gap: 15px;
		}
		
		.current-level {
			font-weight: bold;
			color: var(--primary-color);
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Chat Service Debug Console</h1>
		</div>
		
		<div class="card">
			<h2>Debug Endpoints</h2>
			<ul>
			{{range .Pages}}
				<li><a href="{{.Path}}">{{.Description}}</a></li>
			{{end}}
			</ul>
		</div>

		<div class="card">
			<h2>Logging Configuration</h2>
			<div class="log-level-container">
				<div>Current log level: <span class="current-level">{{.LogLevel}}</span></div>
			</div>
			<form onSubmit="putLogLevel(); return false;">
				<select id="log-level-select">
					<option value="debug" {{if eq .LogLevel "debug"}}selected{{end}}>Debug</option>
					<option value="info" {{if eq .LogLevel "info"}}selected{{end}}>Info</option>
					<option value="warn" {{if eq .LogLevel "warn"}}selected{{end}}>Warn</option>
					<option value="error" {{if eq .LogLevel "error"}}selected{{end}}>Error</option>
					<option value="fatal" {{if eq .LogLevel "fatal"}}selected{{end}}>Fatal</option>
					<option value="panic" {{if eq .LogLevel "panic"}}selected{{end}}>Panic</option>
				</select>
				<input type="submit" value="Change Level"></input>
			</form>
		</div>
	</div>
	
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
