package main

import (
	"github.com/json-iterator/go"
	"github.com/julienschmidt/httprouter"
	"github.com/oschwald/geoip2-golang"
	"github.com/spf13/pflag"
	"log"
	"net"
	"net/http"
)

type geoResponseStruct struct {
	IP          string  `json:"ip"`
	Continent   string  `json:"continent"`
	CountryName string  `json:"country_name"`
	CountryCode string  `json:"country_code"`
	StateName   string  `json:"state_name"`
	CityName    string  `json:"city_name"`
	PostalCode  string  `json:"zip_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	TimeZone    string  `json:"time_zone"`
}

var db *geoip2.Reader

func main() {
	var (
		bindIP string
		bindPort string
		dbPath string
		prefix string
	)
	pflag.StringVarP(&bindIP, "bindip", "b", "0.0.0.0", "the ip address to bind to")
	pflag.StringVarP(&bindPort, "port", "p", "8080", "port to listen on")
	pflag.StringVarP(&dbPath, "dbpath", "d", "./GeoLite2-City.mmdb", "full db file path")
	pflag.StringVarP(&prefix, "routeprefix", "r", "/geoip", "route prefix for geoip service, cant be empty")

	pflag.Parse()

	var err error
	db, err = geoip2.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	router := httprouter.New()
	router.GET(prefix + "/json/:ip", contentTypeMiddleware(geoHandler))
	router.GET(prefix  + "/healthcheck", healthcheck)
	log.Fatal(http.ListenAndServe(bindIP + ":" + bindPort, router))
}

func healthcheck(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	return
}

func contentTypeMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		next(w, r, ps)
	}
}

func errResponse(w http.ResponseWriter, statusCode int, errStr string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"err": "` + errStr + `"}`))
}

func geoResponse(w http.ResponseWriter, geo geoResponseStruct) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	j, err := json.Marshal(geo)
	if err != nil {
		errResponse(w, http.StatusInternalServerError, "")
		return
	}
	w.Write(j)
}

func geoHandler(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	ipStr := ps.ByName("ip")
	ip := net.ParseIP(ipStr)
	if ip == nil {
			errResponse(w, http.StatusBadRequest, "bad ip")
			return
	}
	geo, err := db.City(ip)
	if err != nil {
		errResponse(w, http.StatusInternalServerError, "lookup error")
		return
	}

	stateName := ""
	if len(geo.Subdivisions) > 0 {
		stateName = geo.Subdivisions[0].Names["en"]
	}
	resp := geoResponseStruct{
		Continent:   geo.Continent.Names["en"],
		CountryName: geo.Country.Names["en"],
		CountryCode: geo.Country.IsoCode,
		StateName:   stateName,
		CityName:    geo.City.Names["en"],
		PostalCode:  geo.Postal.Code,
		Latitude:    geo.Location.Latitude,
		Longitude:   geo.Location.Longitude,
		TimeZone:    geo.Location.TimeZone,
	}

	geoResponse(w, resp)
}
