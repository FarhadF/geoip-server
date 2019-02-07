# geoip-server
No bullshit blazing fast geoip server

## Usage:
Download and uncompress the maxmind [geolite2 (city)](https://dev.maxmind.com/geoip/geoip2/geolite2/) database.
```
geoip:
  -b, --bindip string        the ip address to bind to (default "0.0.0.0")
  -d, --dbpath string        full db file path (default "./GeoLite2-City.mmdb")
  -p, --port string          port to listen on (default "8080")
  -r, --routeprefix string   route prefix for geoip service, cant be empty (default "/geoip")
```
