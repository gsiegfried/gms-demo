package main

import (
	"encoding/json"
	"log"
	"net/http"

	"sync"

	"cloud.google.com/go/trace"
	"github.com/gsiegfried/gms-demo/pb/geo"
	"github.com/gsiegfried/gms-demo/pb/profile"
	"github.com/gsiegfried/gms-demo/pb/rate"
)

type response struct {
	Type     string    `json:"type"`
	Features []feature `json:"features"`
}

type feature struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Properties struct {
		Name        string `json:"name"`
		PhoneNumber string `json:"phone_number"`
	} `json:"properties"`
	Geometry struct {
		Type        string    `json:"type"`
		Coordinates []float32 `json:"coordinates"`
	} `json:"geometry"`
}

func main() {
	e := newEnv()
	http.Handle("/", requestHandler(e))
	log.Fatal(http.ListenAndServe(e.serviceAddr(), nil))
}

func requestHandler(e *env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		span := e.Tracer.SpanFromRequest(r)
		defer span.Finish()
		ctx := trace.NewContext(r.Context(), span)

		// in/out dates from query params
		inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
		if inDate == "" || outDate == "" {
			http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
			return
		}

		// finds nearby hotels
		// TODO(hw): use lat/lon from request params
		nearby, err := e.GeoClient.Nearby(ctx, &geo.Request{
			Lat: 37.7749,
			Lon: -122.4194,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var profileRes *profile.Result
		var rateRes *rate.Result

		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			profileRes, err = e.ProfileClient.GetProfiles(ctx, &profile.Request{
				HotelIds: nearby.HotelIds,
				Locale:   "en",
			})
		}()

		go func() {
			defer wg.Done()
			rateRes, err = e.RateClient.GetRates(ctx, &rate.Request{
				HotelIds: nearby.HotelIds,
				InDate:   inDate,
				OutDate:  outDate,
			})
		}()

		wg.Wait()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(
			geoJSONResponse(profileRes.Hotels),
		)
	})
}

func geoJSONResponse(hotels []*profile.Hotel) response {
	r := response{
		Type: "FeatureCollection",
	}

	for _, hotel := range hotels {
		f := feature{
			Type: "Feature",
			ID:   hotel.Id,
		}
		f.Properties.Name = hotel.Name
		f.Properties.PhoneNumber = hotel.PhoneNumber
		f.Geometry.Type = "Point"
		f.Geometry.Coordinates = []float32{
			hotel.Address.Lon,
			hotel.Address.Lat,
		}
		r.Features = append(r.Features, f)
	}

	return r
}
