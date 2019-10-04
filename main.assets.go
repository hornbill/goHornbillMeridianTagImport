package main

import (
	"encoding/xml"
	"errors"
	"fmt"

	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
	"github.com/hornbill/pb"
)

//cacheAssets  - caches asset records from instance
func cacheAssets() error {
	//Get Count
	var err error
	assetCount, err = getAssetCount()
	if err != nil {
		return err
	}

	if assetCount == 0 {
		return errors.New("No assets could be found on your Hornbill instance")
	}
	var i int
	hornbillHelpers.Logger(1, "Retrieving "+fmt.Sprint(assetCount)+" assets from Hornbill. Please wait...", true, logFileName)

	bar := pb.New(assetCount)
	bar.ShowPercent = false
	bar.ShowCounters = false
	bar.ShowTimeLeft = false
	bar.Start()
	for i = 0; i <= assetCount; i += xmlmcPageSize {
		blockAssets, err := getAssets(i, xmlmcPageSize)
		if err != nil {
			bar.Finish()
			return err
		}
		if len(blockAssets) > 0 {
			for _, v := range blockAssets {
				keyval := getKeyVal(&v)
				assets[keyval] = v
			}
		}
		bar.Add(xmlmcPageSize)
	}
	bar.Finish()
	hornbillHelpers.Logger(1, fmt.Sprint(len(assets))+" assets cached.", true, logFileName)
	return err
}

func getKeyVal(asset *assetDetailsStruct) string {
	switch importConf.AssetMatchColumn.Hornbill {
	case "PrimaryKey":
		return asset.AssetID
	case "Description":
		return asset.AssetDescription
	case "Name":
		return asset.AssetName
	case "Tag":
		return asset.AssetTag
	}
	return asset.AssetName
}

func getAssetCount() (int, error) {
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("table", "h_cmdb_assets")
	if configDryrun {
		hornbillHelpers.Logger(3, "[DRYRUN] [ASSETS] [COUNT] "+espXmlmc.GetParam(), false, logFileName)
	}
	xmlAssetCount, err := espXmlmc.Invoke("data", "getRecordCount")
	if err != nil {
		return 0, errors.New("getAssetCount:Invoke:" + err.Error())
	}

	var xmlResponse methodCallResult
	err = xml.Unmarshal([]byte(xmlAssetCount), &xmlResponse)
	if err != nil {
		return 0, errors.New("getAssetCount:Unmarshal:" + err.Error())
	}
	if xmlResponse.Status != "ok" {
		return 0, errors.New("getAssetCount:Xmlmc:" + xmlResponse.State.ErrorRet)
	}
	return xmlResponse.Params.Count, err
}

func getAssets(rowStart, limit int) ([]assetDetailsStruct, error) {
	var assets []assetDetailsStruct
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("queryName", "getAssetsList")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("rowstart", fmt.Sprint(rowStart))
	espXmlmc.SetParam("limit", fmt.Sprint(limit))
	espXmlmc.CloseElement("queryParams")
	if configDryrun {
		hornbillHelpers.Logger(3, "[DRYRUN] [ASSETS] [GET] "+espXmlmc.GetParam(), false, logFileName)
	}
	xmlAssets, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		return assets, errors.New("getAssets:Invoke:" + err.Error())
	}

	var xmlResponse methodCallResult
	err = xml.Unmarshal([]byte(xmlAssets), &xmlResponse)
	if err != nil {
		return assets, errors.New("getAssets:Unmarshal:" + err.Error())
	}
	if xmlResponse.Status != "ok" {
		return assets, errors.New("getAssets:Xmlmc:" + xmlResponse.State.ErrorRet)
	}
	return xmlResponse.Params.Assets, err
}

func updateAsset(assetID string, assetTag *assetTagStruct) error {
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("entity", "Asset")
	espXmlmc.SetParam("returnModifiedData", "true")
	espXmlmc.OpenElement("primaryEntityData")
	espXmlmc.OpenElement("record")
	espXmlmc.SetParam("h_pk_asset_id", assetID)
	for k, v := range importConf.ImportMapping {
		espXmlmc.SetParam(k, getKeyValMeridian(v, assetTag))
	}
	espXmlmc.CloseElement("record")
	espXmlmc.CloseElement("primaryEntityData")
	if configDryrun {
		hornbillHelpers.Logger(1, espXmlmc.GetParam(), false, logFileName)
		counters.assetsSkipped++
		espXmlmc.ClearParam()
		return nil
	}
	xmlAssets, err := espXmlmc.Invoke("data", "entityUpdateRecord")
	if err != nil {
		return errors.New("updateAsset:Invoke:" + err.Error())
	}

	var xmlResponse methodCallResult
	err = xml.Unmarshal([]byte(xmlAssets), &xmlResponse)
	if err != nil {
		return errors.New("updateAsset:Unmarshal:" + err.Error())
	}
	if xmlResponse.Status != "ok" {
		return errors.New("updateAsset:Xmlmc:" + xmlResponse.State.ErrorRet)
	}
	if len(xmlResponse.Params.UpdatedCols.ColList) > 0 {
		hornbillHelpers.Logger(1, "Asset Updated", false, logFileName)
		counters.assetsUpdated++
	} else {
		hornbillHelpers.Logger(1, "Nothing to update", false, logFileName)
		counters.assetsSkipped++
	}
	return nil
}
