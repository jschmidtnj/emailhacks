package main

import ()

var formEmailTemplate = `
<!DOCTYPE html>
<html ⚡4email>
  <head>
    <meta charset="utf-8" />
    <style amp4email-boilerplate>
      body {
        visibility: hidden;
      }
    </style>
    <script async src="https://cdn.ampproject.org/v0.js"></script>
  </head>
  <body>
    <h1>{{ .Title }}</h1>
  </body>
</html>
`

// TODO - get form as email to send to client
// use this: https://gowebexamples.com/templates/
func getFormEmailData(form *Form) (string, error) {
	minifiedData, err := minifier.String("text/html", formEmailTemplate)
	if err != nil {
		return "", err
	}
	return minifiedData, nil
}
