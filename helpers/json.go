package main

import "encoding/json"
import "errors"
import "io/ioutil"
import "os"

// DecodeFile is a more generic JSON parser
func DecodeFile(fileName string, object interface{}) error {

	//Open the config file
	file, err := os.Open(fileName)

	if err != nil {
		return errors.New("Could not open file: " + err.Error())
	}

	jsonParser := json.NewDecoder(file)
	err = jsonParser.Decode(object)

	if err != nil {
		return errors.New("Could not parse file: " + err.Error())
	} else {
		return nil
	}

}

// EncodeFile is a more generic JSON encoder (to file)
func EncodeFile(fileName string, object interface{}) error {

	json, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fileName, json, 0644)
	if err != nil {
		return errors.New("Couldn't write JSON to disk: " + err.Error())
	}

	return nil

}
