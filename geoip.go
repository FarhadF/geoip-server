package main

import (
	"compress/gzip"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/julienschmidt/httprouter"
	"github.com/oschwald/geoip2-golang"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
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

var dbBytes []byte

type maxmind struct {
	mutex sync.RWMutex
	db    *geoip2.Reader
}

var m maxmind

func main() {
	var (
		bindIP         string
		bindPort       string
		prefix         string
		license        string
		accountid      int
		updateInterval int
		edition        string
	)
	pflag.StringVarP(&bindIP, "bindip", "b", "0.0.0.0", "the ip address to bind to")
	pflag.StringVarP(&bindPort, "port", "p", "8080", "port to listen on")
	pflag.IntVarP(&updateInterval, "updateinterval", "u", 24, "intervals (hour) to check for database updates")
	pflag.StringVarP(&license, "license", "l", "", "sign up and generate this at maxmind website")
	pflag.IntVarP(&accountid, "accountid", "a", 0, "sign up and generate this at maxmind website")
	pflag.StringVarP(&edition, "edition", "e", "GeoLite2-City", "edition of database to download")
	pflag.StringVarP(&prefix, "routeprefix", "r", "/geoip", "route prefix for geoip service, cant be empty")

	pflag.Parse()
	url := fmt.Sprintf("https://updates.maxmind.com/geoip/databases/%s/update", edition)
	db, err := download(url, accountid, license)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	log.Info().Msg("download finished")
	err = reload(db)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	defer m.db.Close()
	go func() {
		for {
			time.Sleep(time.Duration(updateInterval) * time.Hour)
			db, err := download(url, accountid, license)
			if err != nil {
				log.Error().Err(err).Msg("downloading update failed")
			}
			log.Info().Msg("download finished")
			err = reload(db)
			if err != nil {
				log.Error().Err(err).Msg("reload failed")
			}
		}
	}()
	router := httprouter.New()
	router.GET(prefix+"/json/:ip", contentTypeMiddleware(geoHandler))
	router.GET(prefix+"/healthcheck", healthcheck)
	log.Fatal().Err(http.ListenAndServe(bindIP+":"+bindPort, router)).Msg("")
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
	_, err := w.Write([]byte(`{"err": "` + errStr + `"}`))
	if err != nil {
		log.Error().Err(err).Msg("")
	}
}

func geoResponse(w http.ResponseWriter, geo geoResponseStruct) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	j, err := json.Marshal(geo)
	if err != nil {
		errResponse(w, http.StatusInternalServerError, "")
		return
	}
	_, err = w.Write(j)
	if err != nil {
		log.Error().Err(err).Msg("")
	}
}

func geoHandler(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	ipStr := ps.ByName("ip")
	ip := net.ParseIP(ipStr)
	if ip == nil {
		errResponse(w, http.StatusBadRequest, "bad ip")
		return
	}
	m.mutex.RLock()
	geo, err := m.db.City(ip)
	m.mutex.RUnlock()
	if err != nil {
		log.Err(err).Msg("")
		errResponse(w, http.StatusInternalServerError, "lookup error")
		return
	}
	stateName := ""
	if len(geo.Subdivisions) > 0 {
		stateName = geo.Subdivisions[0].Names["en"]
	}
	resp := geoResponseStruct{
		IP:          ipStr,
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

func download(url string, accountid int, license string) ([]byte, error) {
	log.Info().Msg("Starting to download the database")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(fmt.Sprintf("%d", accountid), license)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tempBytes, err := ioutil.ReadAll(gzr)
	return tempBytes, nil
}

func reload(newDB []byte) error {
	newReader, err := geoip2.FromBytes(newDB)
	if err != nil {
		return err
	}
	m.mutex.Lock()
	m.db = newReader
	m.mutex.Unlock()
	return nil

}
