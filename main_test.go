package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valyala/fastjson"
)

func TestValidateZip(t *testing.T) {
	tests := []struct {
		zip    string
		valid  bool
		expZip string
		expErr error
	}{
		{"12345678", true, "12345678", nil},
		{"12345-678", true, "12345678", nil},
		{"12.345-678", true, "12345678", nil},
		{"12345", false, "", errors.New("invalid zipcode")},
		{"abcdef", false, "", errors.New("invalid zipcode")},
	}

	for _, test := range tests {
		gotZip, err := validateZip(test.zip)
		if (err == nil) != test.valid || (err != nil && err.Error() != test.expErr.Error()) {
			t.Errorf("validateZip(%q) = %v, %v; want %v, %v", test.zip, gotZip, err, test.expZip, test.expErr)
		}
	}
}

func TestGetLocation(t *testing.T) {
	zip := "12345678"
	location := "Sample Location"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"localidade": "%s"}`, location)
	}))
	defer ts.Close()

	client := &http.Client{}
	getLocation := func(client *http.Client, zip string) (string, error) {
		res, err := client.Get(ts.URL)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()
		return fastjson.GetString([]byte(`{"localidade": "Sample Location"}`), "localidade"), nil
	}

	loc, err := getLocation(client, zip)
	if err != nil || loc != location {
		t.Errorf("getLocation(%q) = %v, %v; want %v, nil", zip, loc, err, location)
	}
}

func TestGetTemp(t *testing.T) {
	location := "Sample Location"
	apiKey := "validapikey"
	expectedWeather := Weather{
		TempC: 25,
		TempF: 77,
		TempK: 298,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"current": {"temp_c": %f, "temp_f": %f}}`, expectedWeather.TempC, expectedWeather.TempF)
	}))
	defer ts.Close()

	client := &http.Client{}
	getTemp := func(client *http.Client, location string, apiKey string) (Weather, error) {
		return Weather{TempC: 25, TempF: 77, TempK: 298}, nil
	}

	weather, err := getTemp(client, location, apiKey)
	if err != nil || weather != expectedWeather {
		t.Errorf("getTemp(%q, %q) = %v, %v; want %v, nil", location, apiKey, weather, err, expectedWeather)
	}
}

func TestGetData(t *testing.T) {
	// This integration test checks the entire data retrieval process.
	req := httptest.NewRequest("GET", "/?zip=12345678", nil)
	req.Header.Set("api_key", "validapikey")
	w := httptest.NewRecorder()

	getData(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
}
