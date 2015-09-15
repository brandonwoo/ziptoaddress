package main

import (
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
	"github.com/levigross/grequests"
	"log"
	"net/http"
)

const GOOGLE_MAPS_API_ENDPOINT = "http://maps.googleapis.com/maps/api/geocode/json"

type googleAddressComponent struct {
	LongName  string `json:"long_name"`
	ShortName string `json:"short_name"`
}

type googleAddressResult struct {
	AddressComponents []googleAddressComponent `json:"address_components"`
}

type googleAddress struct {
	Status  string                `json:"status"`
	Results []googleAddressResult `json:"results"`
}

type address struct {
	Area    string `json:"area,omitempty"`
	City    string `json:"city,omitempty"`
	Address string `json:"address,omitempty"`
	Error   string `json:"error,omitempty"`
}

func GetAddress(zipcode string) address {
	resp, err := grequests.Get(GOOGLE_MAPS_API_ENDPOINT, &grequests.RequestOptions{
		Params: map[string]string{
			"language": "ja",
			"address":  zipcode,
		},
	})
	if err != nil {
		log.Fatalln("Unable to make request: ", err)
	}

	var a googleAddress
	err = json.Unmarshal(resp.Bytes(), &a)
	if err != nil {
		log.Println("google address json unmarshal error: ", err)
		return address{
			Error: "google address json unmarshal error",
		}
	}

	if len(a.Results) < 1 || len(a.Results[0].AddressComponents) < 5 {
		log.Println("google address json unmarshal error: ", err)
		return address{
			Error: "no results found",
		}
	}

	components := a.Results[0].AddressComponents
	components = components[:len(components)-1] //remove country
	components = components[1:len(components)]  //remove zipcode

	result := address{}
	result.Area = components[len(components)-1].LongName
	result.City = components[len(components)-2].LongName

	//concatenate remaining components to form address string
	addressString := ""
	for i := len(components) - 3; i >= 0; i-- {
		addressString += components[i].LongName
	}
	result.Address = addressString

	return result
}

func main() {
	router := httprouter.New()
	router.GET("/get_address/:zipcode", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		zipcode := ps.ByName("zipcode")
		a := GetAddress(zipcode)
		js, err := json.Marshal(a)
		if err != nil {
			w.Write([]byte("return address json marshal error"))

		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		}
	})

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":3000")
}
