package hereapi

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/artemnikitin/android-go-shared/builder"
	"github.com/artemnikitin/android-go-shared/logger"
	"github.com/artemnikitin/android-go-shared/model"
)

// GetCoordinates implements HERE Geocoding API for convert text address to GPS coordinates
func GetCoordinates(appID, appToken, searchText string) string {
	url := builder.NewGeocodingService().
		SetAppID(appID).SetAppToken(appToken).
		SetSearchPhrase(searchText).Build()
	resp := sendRequest(url)
	defer closeAfter(resp)
	if resp.StatusCode != 200 {
		return ""
	}
	bytes := getBody(resp)
	lat, lon := getCoordinatesFromJSON(bytes)
	return createStringFromCoordinates(lat, lon)
}

// GetPicture returns map tile for specific set of GPS coordinates
func GetPicture(appID, appToken string, lat, lon float64, h, w, dpi int) []byte {
	url := builder.NewMapTileService().
		SetAppID(appID).SetAppToken(appToken).
		SetLatitude(lat).SetLongitude(lon).
		SetWidth(w).SetHeight(h).SetDpi(dpi).Build()
	resp := sendRequest(url)
	defer closeAfter(resp)
	if resp.StatusCode != 200 {
		return make([]byte, 0)
	}
	return getBody(resp)
}

// GetAutosuggestions returns a struct that represent lisr of suggestion for input from HERE API
func GetAutosuggestions(appID, appToken, searchPhrase string, lat, lon float64) *model.AutosuggestionResponse {
	res := &model.AutosuggestionResponse{}
	url := builder.NewAutosuggestionService().
		SetAppID(appID).SetAppToken(appToken).
		SetLatitude(lat).SetLongitude(lon).
		SetSearchPhrase(searchPhrase).Build()
	resp := sendRequest(url)
	defer closeAfter(resp)
	if resp.StatusCode != 200 {
		return res
	}
	err := json.Unmarshal(getBody(resp), res)
	if err != nil {
		log.Println("Can't parse JSON:", err)
		return res
	}
	return res
}

func sendRequest(data string) *http.Response {
	resp, err := http.Get(data)
	logger.ProcessError("Can't execute HTTP request", err)
	return resp
}

func getBody(resp *http.Response) []byte {
	response, err := ioutil.ReadAll(resp.Body)
	logger.ProcessError("Can't get a body of HTTP response", err)
	return response
}

func closeAfter(resp *http.Response) {
	// Drain and close the body to let the Transport reuse the connection
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}

func getCoordinatesFromJSON(response []byte) (float64, float64) {
	var geocode = &model.GeocodingResponse{}
	var lat, lon float64
	err := json.Unmarshal(response, geocode)
	if err != nil {
		log.Println("Can't parse JSON:", err)
		return lat, lon
	}
	lat = geocode.Response.View[0].Result[0].Location.DisplayPosition.Latitude
	lon = geocode.Response.View[0].Result[0].Location.DisplayPosition.Longitude
	return lat, lon
}

func createStringFromCoordinates(lat, lon float64) string {
	var bytes bytes.Buffer
	bytes.WriteString(strconv.FormatFloat(lat, 'f', -1, 64))
	bytes.WriteString("||")
	bytes.WriteString(strconv.FormatFloat(lon, 'f', -1, 64))
	return bytes.String()
}
