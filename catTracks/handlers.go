package catTracks

//Handles
import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/rotblauer/trackpoints/trackPoint"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

//var funcMap = template.FuncMap{
//	"eq": func(a, b interface{}) bool {
//		return a == b
//	},
//}

// the html stuff of this thing
var templates = template.Must(template.ParseGlob("templates/*.html"))

////For passing to the template , might not need to pass?
//type Data struct {
//	TrackPoints     []*trackPoint.TrackPoint
//	TrackPointsJSON []byte
//}

//Welcome, loads and servers all (currently) data pointers
func indexHandler(w http.ResponseWriter, r *http.Request) {

	//templates.Funcs(funcMap)
	templates.ExecuteTemplate(w, "base", nil)
}

func getPointsJSON(w http.ResponseWriter, r *http.Request) {
	var query query
	q := r.URL.Query()
	e := q.Get("epsilon")
	if e == "" {
		e = "0.001"
	}
	eps, er := strconv.ParseFloat(e, 64)
	if er != nil {
		fmt.Println("shit parsefloat eps")
		query.Epsilon = 0.001
	} else {
		query.Epsilon = eps
	}

	data, eq := getData(query)
	if eq != nil {
		http.Error(w, eq.Error(), http.StatusInternalServerError)
	}
	fmt.Println("Receive ajax get data string ")
	w.Write([]byte(data))
}

func receiveAjax(w http.ResponseWriter, r *http.Request) {
	var query query
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}
	data, e := getData(query)

	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	fmt.Println("Receive ajax post data string ")
	w.Write([]byte(data))

}

func getMap(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "map", nil)
}

func getData(query query) ([]byte, error) {
	var data []byte
	allPoints, e := getAllPoints(query)
	if e != nil {
		return data, e
	}
	data, err := json.Marshal(allPoints)
	if err != nil {
		return data, err
	}
	return data, nil
}

//TODO populate a population of points
func populatePoints(w http.ResponseWriter, r *http.Request) {
	var trackPoints trackPoint.TrackPoints

	if r.Body == nil {
		http.Error(w, "Please send a request body", 500)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&trackPoints)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	errS := storePoints(trackPoints)
	if errS != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//return json of trakcpoint if stored succcess
	if errW := json.NewEncoder(w).Encode(&trackPoints); errW != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func uploadCSV(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 30)
	file, _, err := r.FormFile("uploadfile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, line := range lines {
		var tp trackPoint.TrackPoint

		tp.Name = line[0]

		if tp.Time, err = time.Parse(time.UnixDate, line[1]); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if tp.Lat, err = strconv.ParseFloat(line[2], 64); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if tp.Lng, err = strconv.ParseFloat(line[3], 64); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		errS := storePoint(tp)
		if errS != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	http.Redirect(w, r, "/", 302) //the 300

}
