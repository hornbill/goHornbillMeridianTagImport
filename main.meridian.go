package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"time"

	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
)

func getTags() (int, error) {

	nextPage := "https://edit.meridianapps.com/api/locations/" + importConf.LocationID + "/asset-beacons?page_size=" + strconv.Itoa(meridianPageSize)
	var assetTagsPage []assetTagStruct
	var err error
	getNextPage := true
	for getNextPage {
		nextPage, assetTagsPage, err = getPageTags(importConf.LocationID, nextPage)

		if err != nil {
			return 0, err
		}
		if nextPage == "" {
			getNextPage = false
		}
		for _, v := range assetTagsPage {
			assetTags[getTagID(v)] = v
		}
	}

	return len(assetTags), nil
}

func getPageTags(location, pageURL string) (string, []assetTagStruct, error) {
	var tagArray []assetTagStruct
	hornbillHelpers.Logger(3, "Tags Page URL: "+pageURL, false, logFileName)

	req, err := http.NewRequest("GET", pageURL, nil)

	if err != nil {
		return "", tagArray, errors.New("getPageTags:http:NewRequest:" + err.Error())
	}
	req.Header.Set("Authorization", "Token "+importConf.MeridianToken)
	req.Header.Set("User-Agent", "Go-http-client/1.1")

	duration := time.Second * time.Duration(30)

	var netTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout: duration,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Timeout:   duration,
		Transport: netTransport,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", tagArray, errors.New("getPageTags:client:Do:" + err.Error())
	}
	defer resp.Body.Close()

	//-- Check for HTTP Response
	if resp.StatusCode != 200 {
		io.Copy(ioutil.Discard, resp.Body)
		return "", tagArray, errors.New("getPageTags:http:InvalidStatusCode:" + strconv.Itoa(resp.StatusCode))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", tagArray, errors.New("Cant read the body of the response")
	}
	var jsonTagsResponse meridianTagsResponse
	err = json.Unmarshal([]byte(string(body)), &jsonTagsResponse)
	if err != nil {
		return "", tagArray, errors.New("getPageTags:Unmarshal:" + err.Error())
	}

	return fmt.Sprint(jsonTagsResponse.Next), jsonTagsResponse.Results, nil
}

func getTagID(tag assetTagStruct) string {
	v := reflect.ValueOf(tag)
	typeOfTag := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if typeOfTag.Field(i).Name == importConf.AssetMatchColumn.Meridian {
			return fmt.Sprint(v.Field(i).Interface())
		}
	}
	return ""
}
