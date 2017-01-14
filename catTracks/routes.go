package catTracks

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		indexHandler,
	},
	Route{
		"MarkerPopulate",
		"POST",
		"/populate/",
		populatePoint,
	},
	Route{
		"UploadCSV",
		"POST",
		"/upload",
		uploadCSV,
	},
}
