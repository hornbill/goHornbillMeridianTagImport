package main

import (
	"strconv"

	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
	"github.com/hornbill/pb"
)

func processTags() {

	hornbillHelpers.Logger(1, "Processing "+strconv.Itoa(len(assetTags))+" Meridian asset tags...", true, logFileName)
	bar := pb.New(len(assetTags))
	bar.ShowPercent = false
	bar.ShowCounters = true
	bar.ShowTimeLeft = false
	bar.Start()

	for _, tag := range assetTags {
		bar.Increment()
		tag.Source = "Meridian"
		assetName := getKeyValMeridian(importConf.AssetMatchColumn.Meridian, &tag)
		hornbillAssetID := getAssetID(assetName)
		if hornbillAssetID == "" {
			hornbillHelpers.Logger(5, "Could not find asset: ["+assetName+"]", false, logFileName)
			continue
		}

		hornbillHelpers.Logger(1, "Processing "+assetName+" ["+hornbillAssetID+"]", false, logFileName)
		err := updateAsset(hornbillAssetID, &tag)
		if err != nil {
			counters.updateFailed++
			hornbillHelpers.Logger(4, err.Error(), false, logFileName)
		}
	}
	bar.Finish()
}

//getAssetID -- Check if asset exists
func getAssetID(assetIdentifier string) string {
	assetRecord, ok := assets[assetIdentifier]
	if ok {
		return assetRecord.AssetID
	}
	return ""
}

func getKeyValMeridian(column string, assetTag *assetTagStruct) string {
	value := ""
	switch column {
	case "Location":
		value = assetTag.Location
	case "ExternalID":
		value = assetTag.ExternalID
	case "ID":
		value = assetTag.ID
	case "Name":
		value = assetTag.Name
	case "Mac":
		value = assetTag.Mac
	case "Source":
		value = "Meridian"
	}
	return value
}
