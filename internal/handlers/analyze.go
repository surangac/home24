package handlers

import (
	"log/slog"
	"net/http"
	"net/url"
)

// This function handles the form submission to analyze webpages
func (r *Router) analyzeHandler(w http.ResponseWriter, req *http.Request) {
	// Only accept POST requests
	if req.Method != "POST" {
		http.Redirect(w, req, "/", 303) // 303 is http.StatusSeeOther
		return
	}

	// Need to parse the form to get the submitted values
	var err = req.ParseForm()
	if err != nil {
		// Log the error for debugging
		r.log.Error("error parsing form", slog.String("error", err.Error()))

		// Send back a bad request response
		http.Error(w, "Bad Request", 400) // 400 is http.StatusBadRequest
		return
	}

	// Get the URL from the form
	var urlString = req.PostForm.Get("url")
	if urlString == "" {
		// If URL is empty, show an error
		var templateData = map[string]interface{}{
			"Error": "URL cannot be empty",
		}

		err = r.tmpl.ExecuteTemplate(w, "index.html", templateData)
		if err != nil {
			r.log.Error("error rendering template", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", 500) // 500 is http.StatusInternalServerError
		}
		return
	}

	// Make sure the URL is valid
	var parsedURL *url.URL
	parsedURL, err = url.Parse(urlString)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		// If URL is invalid, show an error
		var templateData = map[string]interface{}{
			"Error": "Invalid URL. Please enter a valid URL starting with http:// or https://",
			"URL":   urlString,
		}

		err = r.tmpl.ExecuteTemplate(w, "index.html", templateData)
		if err != nil {
			r.log.Error("error rendering template", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", 500) // 500 is http.StatusInternalServerError
		}
		return
	}

	// Call the analyzer to analyze the webpage
	var result, analyzeErr = r.analyzer.Analyze(req.Context(), urlString)
	if analyzeErr != nil {
		// If analysis fails, log the error and show it to the user
		r.log.Error("error analyzing page",
			slog.String("url", urlString),
			slog.String("error", analyzeErr.Error()),
		)

		var templateData = map[string]interface{}{
			"Error": analyzeErr.Error(),
			"URL":   urlString,
		}

		err = r.tmpl.ExecuteTemplate(w, "index.html", templateData)
		if err != nil {
			r.log.Error("error rendering template", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", 500) // 500 is http.StatusInternalServerError
		}
		return
	}

	// If everything is successful, show the results
	var templateData = map[string]interface{}{
		"URL":    urlString,
		"Result": result,
	}

	err = r.tmpl.ExecuteTemplate(w, "result.html", templateData)
	if err != nil {
		r.log.Error("error rendering template", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", 500) // 500 is http.StatusInternalServerError
	}
}
