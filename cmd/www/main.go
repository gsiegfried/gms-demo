package main

import (
	"io"
	"os"
	"fmt"
	"net/url"
	"net/http"
	"net/http/httputil"
)

func main() {

	remoteApi, err := url.Parse("http://api:8080")
	if err != nil {
		panic(err);
	}
	proxy := httputil.NewSingleHostReverseProxy(remoteApi)
	gmk := os.Getenv("GMAPI")
	
	htmlTail := fmt.Sprintf(htmlBodyB, gmk)
	htmlBody := fmt.Sprintf("%s%s", htmlBodyA, htmlTail)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, htmlBody)
	})
	http.HandleFunc("/api", handler(proxy))

	http.ListenAndServe(":5000", nil)
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		r.URL.Path = ""
		w.Header().Set("X-Rudy", "Grs")
		p.ServeHTTP(w, r)
	}

}

const htmlBodyA = `
<!DOCTYPE html>
<html>
  <head>
    <style>
      html, body {
        height: 100%;
        margin: 0;
        padding: 0;
      }
      #map {
        height: 100%;
      }
    </style>
  </head>
  <body>
    <div id="map"></div>
    <script>
      var map;
      
      function initMap() {
        var infowindow = new google.maps.InfoWindow();

        map = new google.maps.Map(document.getElementById('map'), {
          zoom: 13,
          center: new google.maps.LatLng(37.7879, -122.4075)
        });

        google.maps.event.addListener(map, 'click', function() {
          infowindow.close();
        });

        map.data.addListener('click', function(event) {
          infowindow.setContent(event.feature.getProperty('name')+"<br>"+event.feature.getProperty('phone_number'));
          infowindow.setPosition(event.latLng);
          infowindow.setOptions({pixelOffset: new google.maps.Size(0,-34)});
          infowindow.open(map);
        });

        map.data.loadGeoJson('/api?inDate=2015-04-09&outDate=2015-04-10');
      }
    </script>
`
const htmlBodyB = `
    <script type="text/javascript" src="http://maps.google.com/maps/api/js?callback=initMap&key=%s" async defer></script>
  </body>
</html>
`
