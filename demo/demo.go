package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/itearo/golang-xlsx-templator"
)

func main() {
	jsonFile, err := os.Open("./report_data.json")
	if err != nil {
		panic(err)
	}

	//goland:noinspection GoUnhandledErrorResult
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var jsonData map[string]interface{}
	err = json.Unmarshal(byteValue, &jsonData)
	if err != nil {
		panic(err)
	}

	err = xlsxtemplator.RenderTemplateWithData(jsonData, "./report_template.xlsx", "./report_result.xlsx")
	if err != nil {
		panic(err)
	}
}
