# geoip-server
No bullshit blazing fast geoip server

## Usage:
1. Sign up(free) for maxmind [geolite2 (city)](https://dev.maxmind.com/geoip/geoip2/geolite2/) database. 
2. Login and go to my licenses and generate a new license(free). 
3. Build : ```go build geoip.go
3. use the flags to provide token and accountid from previous step and run.
4. if you are using default routeprefix try: ```curl localhost:8080/geoip/json/50.19.0.1```
```
geoip:
  -a, --accountid int        sign up and generate this at maxmind website
  -b, --bindip string        the ip address to bind to (default "0.0.0.0")
  -e, --edition string       edition of database to download (default "GeoLite2-City")
  -l, --license string       sign up and generate this at maxmind website
  -p, --port string          port to listen on (default "8080")
  -r, --routeprefix string   route prefix for geoip service, cant be empty (default "/geoip")
  -u, --updateinterval int   intervals (hour) to check for database updates (default 24)
```
