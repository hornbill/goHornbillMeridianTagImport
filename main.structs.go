package main

import (
	"encoding/xml"
	"time"

	apiLib "github.com/hornbill/goApiLib"
)

//----- Constants -----
const (
	version          = "1.0.0"
	xmlmcPageSize    = 100
	meridianPageSize = 100
)

//----- Variables -----
var (
	assetCount     int
	assets         = make(map[string]assetDetailsStruct)
	assetTags      = make(map[string]assetTagStruct)
	counters       counterTypeStruct
	configDryrun   bool
	configFileName string
	configVersion  bool
	espXmlmc       *apiLib.XmlmcInstStruct
	importConf     sqlImportConfStruct
	logFileName    = "meridianAssetImport" + time.Now().Format("20060102150405") + ".log"
)

type counterTypeStruct struct {
	assetsUpdated int
	assetsSkipped int
	updateFailed  int
}

//-- Config Structs
type sqlImportConfStruct struct {
	APIKey           string
	InstanceID       string
	MeridianToken    string
	LocationID       string
	AssetMatchColumn assetColumnStruct
	ImportMapping    map[string]string
}

type assetColumnStruct struct {
	Meridian string
	Hornbill string
}

//-- XMLMC Call Structs
type methodCallResult struct {
	State  stateStruct  `xml:"state"`
	Status string       `xml:"status,attr"`
	Params paramsStruct `xml:"params"`
}
type stateStruct struct {
	Code     string `xml:"code"`
	ErrorRet string `xml:"error"`
}
type paramsStruct struct {
	Count       int                  `xml:"count"`
	Assets      []assetDetailsStruct `xml:"rowData>row"`
	UpdatedCols updatedCols          `xml:"primaryEntityData>record"`
}

type assetDetailsStruct struct {
	AssetID          string `xml:"h_pk_asset_id"`
	AssetDescription string `xml:"asset_description"`
	AssetName        string `xml:"asset_name"`
	AssetTag         string `xml:"h_asset_tag"`
}

type updatedCols struct {
	AssetPK string       `xml:"h_pk_asset_id"`
	ColList []updatedCol `xml:",any"`
}

type updatedCol struct {
	XMLName xml.Name `xml:""`
	Amount  string   `xml:",chardata"`
}

type meridianTagsResponse struct {
	Next     string           `json:"next"`
	Previous string           `json:"previous"`
	Results  []assetTagStruct `json:"results"`
}

type assetTagStruct struct {
	Created    string `json:"created"`
	ExternalID string `json:"external_id"`
	ID         string `json:"id"`
	Location   string `json:"location"`
	Mac        string `json:"mac"`
	Modified   string `json:"modified"`
	Name       string `json:"name"`
	Source     string
}
