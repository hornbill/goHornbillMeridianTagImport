package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	apiLib "github.com/hornbill/goApiLib"
	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
)

func main() {
	//-- Grab Flags
	flag.StringVar(&configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.BoolVar(&configVersion, "version", false, "Return version and end")
	flag.BoolVar(&configDryrun, "dryrun", false, "Outputs the expected API calls to the log, without actually performing the API calls")
	flag.Parse()

	//-- If configVersion just output version number and die
	if configVersion {
		fmt.Printf("%v \n", version)
		return
	}

	//Load Config
	importConf = loadConfig()

	//Create global espxmlmc session
	espXmlmc = apiLib.NewXmlmcInstance(importConf.InstanceID)
	espXmlmc.SetAPIKey(importConf.APIKey)

	//-- Output
	hornbillHelpers.Logger(2, "---- XMLMC Meridian Asset Tag Import Utility V"+version+" ----", true, logFileName)
	hornbillHelpers.Logger(2, "Flag - Config File "+configFileName, true, logFileName)

	//Cache Service Manager Asset Records
	err := cacheAssets()
	if err != nil {
		hornbillHelpers.Logger(4, "Error when caching assets from Hornbill: "+err.Error(), true, logFileName)
		os.Exit(1)
	}

	//Get Asset Tag records from Meridian
	tagCount, err := getTags()
	if err != nil {
		hornbillHelpers.Logger(4, err.Error(), true, logFileName)
		os.Exit(1)
	}
	if tagCount == 0 {
		hornbillHelpers.Logger(5, "No asset tags found in your Meridian locations!", true, logFileName)
		os.Exit(1)
	}
	processTags()

	//Output
	hornbillHelpers.Logger(2, "Processing Complete!", true, logFileName)
	hornbillHelpers.Logger(2, "* Tags Found: "+strconv.Itoa(len(assetTags)), true, logFileName)
	hornbillHelpers.Logger(2, "* Assets Updated: "+strconv.Itoa(counters.assetsUpdated), true, logFileName)
	hornbillHelpers.Logger(2, "* Assets Skipped: "+strconv.Itoa(counters.assetsSkipped), true, logFileName)
	logPrefix := 2
	if counters.updateFailed > 0 {
		logPrefix = 4
	}
	hornbillHelpers.Logger(logPrefix, "* Asset Updates Failed: "+strconv.Itoa(counters.updateFailed), true, logFileName)
	os.Exit(0)
}

//loadConfig -- Function to Load Configruation File
func loadConfig() sqlImportConfStruct {
	//-- Check Config File File Exists
	cwd, _ := os.Getwd()
	configurationFilePath := cwd + "/" + configFileName
	hornbillHelpers.Logger(1, "Loading Config File: "+configurationFilePath, false, logFileName)
	if _, fileCheckErr := os.Stat(configurationFilePath); os.IsNotExist(fileCheckErr) {
		hornbillHelpers.Logger(4, "No Configuration File", true, logFileName)
		os.Exit(102)
	}
	//-- Load Config File
	file, fileError := os.Open(configurationFilePath)
	//-- Check For Error Reading File
	if fileError != nil {
		hornbillHelpers.Logger(4, "Error Opening Configuration File: "+fmt.Sprintf("%v", fileError), true, logFileName)
	}

	//-- New Decoder
	decoder := json.NewDecoder(file)
	//-- New Var based on SQLimportConf
	esqlConf := sqlImportConfStruct{}
	//-- Decode JSON
	err := decoder.Decode(&esqlConf)
	//-- Error Checking
	if err != nil {
		hornbillHelpers.Logger(4, "Error Decoding Configuration File: "+fmt.Sprintf("%v", err), true, logFileName)
	}
	//-- Return New Congfig
	return esqlConf
}
