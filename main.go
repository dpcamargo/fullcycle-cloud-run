package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/valyala/fastjson"
)

type ZIP struct {
	Location string `json:"localidade"`
}

type Weather struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

const (
	CEP_SIZE = 8
)

func main() {
	http.HandleFunc("/", getData)
	log.Fatal((http.ListenAndServe(":8080", nil)))
}

func getData(w http.ResponseWriter, r *http.Request) {
	zip := r.URL.Query().Get("zip")
	apiKey := r.Header.Get("api_key")
	if apiKey == "" {
		http.Error(w, "api_key is required", http.StatusBadRequest)
		return
	}

	client := &http.Client{}

	zip, err := validateZip(zip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	location, err := getLocation(client, zip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	weather, err := getTemp(client, location, apiKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(weather)
}

func validateZip(zip string) (string, error) {
	re := regexp.MustCompile(`\d+`)
	match := re.FindAllString(zip, -1)
	zip = strings.Join(match, "")

	if len(zip) != CEP_SIZE {
		return "", errors.New("invalid zipcode")
	}
	return zip, nil
}

func getLocation(client *http.Client, zip string) (string, error) {
	baseURL := "http://viacep.com.br/ws/%s/json/"
	URL := fmt.Sprintf(baseURL, zip)

	res, err := client.Get(URL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	r, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	location := fastjson.GetString(r, "localidade")
	if location == "" {
		return "", errors.New("can not find zipcode")
	}

	if location == "" {
		return "", errors.New("error getting location from viacep")
	}

	return location, nil
}

func getTemp(client *http.Client, location string, apiKey string) (Weather, error) {
	weather := Weather{}
	baseURL := "http://api.weatherapi.com/v1/current.json"
	u, err := url.Parse(baseURL)
	if err != nil {
		return Weather{}, err
	}

	q := u.Query()
	q.Set("q", location)
	q.Set("key", apiKey)
	u.RawQuery = q.Encode()

	res, err := client.Get(u.String())
	if err != nil {
		return Weather{}, err
	}
	defer res.Body.Close()

	r, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Weather{}, err
	}

	var p fastjson.Parser
	v, err := p.Parse(string(r))
	if err != nil {
		return Weather{}, err
	}
	weather.TempC = v.GetFloat64("current", "temp_c")
	weather.TempF = v.GetFloat64("current", "temp_f")
	weather.TempK = weather.TempC + 273
	if weather.TempC == 0 {
		return Weather{}, errors.New("error getting weather, invalid API key")
	}

	return weather, nil
}
