package xlsx_templator

import (
	"encoding/json"
	"fmt"
	"github.com/tealeg/xlsx/v3"
	"reflect"
	"testing"
)

func TestReverseSlice(t *testing.T) {
	// Arrange
	expectedSlice := []string{"t", "e", "m", "p", "l", "a", "t", "o", "r"}
	slice := []string{"r", "o", "t", "a", "l", "p", "m", "e", "t"}

	//Act
	reversedSlice := reverseSlice(slice)

	//Assert
	if !reflect.DeepEqual(expectedSlice, reversedSlice) {
		t.Error(fmt.Sprintf("Expected %s got %s", expectedSlice, reversedSlice))
	}
}

func TestFindPropertyInRow(t *testing.T) {
	var tests = []struct {
		cellValues       []string
		expectedProperty string
	}{
		{[]string{"", "{text}", "{{lv1}}"}, "lv1"},
		{[]string{"{{lv1}}"}, "lv1"},
		{[]string{"{{ lv1 }}"}, "lv1"},
		{[]string{"{{lv1.lv2.lv3}}"}, "lv1.lv2.lv3"},
	}

	for index, testData := range tests {
		testTitle := fmt.Sprintf("%d) Input: '%s', expectation: '%s'", index, testData.cellValues, testData.expectedProperty)
		t.Run(testTitle, func(t *testing.T) {
			// Arrange
			sheet := xlsx.Sheet{}
			row := sheet.AddRow()

			for _, cellValue := range testData.cellValues {
				row.AddCell().SetValue(cellValue)
			}

			//Act
			foundProperty, err := findPropertyInRow(row)

			//Assert
			if testData.expectedProperty != foundProperty {
				t.Error(fmt.Sprintf("Expected '%s' got '%s'", testData.expectedProperty, foundProperty))
			}
			if err != nil {
				t.Error("Found error: ", err)
			}
		})
	}
}

func TestFindSliceInPropertyPath(t *testing.T) {
	var tests = []struct {
		ctxJson              string
		property             string
		expectedPropertyPath []string
		expectedSlice        []interface{}
		expectedOk           bool
	}{
		{"{}", "", []string{}, []interface{}{}, false},
		{"{\"lv1\": {\"lv2\": [\"123\"]}}", "lv1", []string{}, []interface{}{}, false},
		{"{\"lv1\": {\"lv2\": [\"123\"]}}", "lv1.lv2", []string{"lv1", "lv2"}, []interface{}{"123"}, true},
		{"{\"lv1\": {\"lv2\": null}}", "lv1.lv2", []string{}, []interface{}{}, false},
		{"{\"lv1\": {\"other\": null}}", "lv1.lv2", []string{}, []interface{}{}, false},
		{"{\"lv1\": {\"lv2\": \"text\"}}", "lv1.lv2.lv3", []string{}, []interface{}{}, false},
	}

	for index, testData := range tests {
		testTitle := fmt.Sprintf("%d) Input: '%s', expectation: '%s'", index, testData.property, testData.expectedPropertyPath)
		t.Run(testTitle, func(t *testing.T) {
			// Arrange
			var ctx map[string]interface{}

			//goland:noinspection GoUnhandledErrorResult
			json.Unmarshal([]byte(testData.ctxJson), &ctx)

			//Act
			resultPropertyPath, resultSlice, resultOk := findSliceInPropertyPath(ctx, testData.property)

			//Assert
			if testData.expectedOk != resultOk {
				t.Error("Expected 'ok' value to be ", map[bool]string{true: "truthy", false: "falsy"}[testData.expectedOk])
			}
			if !reflect.DeepEqual(testData.expectedPropertyPath, resultPropertyPath) {
				t.Error(fmt.Sprintf("Expected '%s' got '%s'", testData.expectedPropertyPath, resultPropertyPath))
			}
			if !reflect.DeepEqual(testData.expectedSlice, resultSlice) {
				t.Error(fmt.Sprintf("Expected '%s' got '%s'", testData.expectedSlice, resultSlice))
			}
		})
	}
}

func TestRenderCell(t *testing.T) {
	var tests = []struct {
		ctxJson           string
		cellValue         string
		cellBgColor       string
		cellHMerge        int
		expectedValue     string
		expectedApplyFill bool
	}{
		{"{ \"lv1\": { \"lv2\": 12345 } }", "{{ lv1.lv2 }}", "", 0, "12345", false},
		{"{}", "", "#000000", 0, "", true},
		{"{}", "", "", 5, " ", false},
	}

	for index, testData := range tests {
		testTitle := fmt.Sprintf("%d) Input: '%s', expectation: '%s'", index, testData.cellValue, testData.expectedValue)
		t.Run(testTitle, func(t *testing.T) {
			// Arrange
			var ctx map[string]interface{}

			//goland:noinspection GoUnhandledErrorResult
			json.Unmarshal([]byte(testData.ctxJson), &ctx)

			cell := xlsx.Cell{
				Value:  testData.cellValue,
				HMerge: testData.cellHMerge,
			}
			cell.SetStyle(&xlsx.Style{Fill: xlsx.Fill{BgColor: testData.cellBgColor}})

			//Act
			err := renderCell(&cell, ctx)

			//Assert
			if cell.Value != testData.expectedValue {
				t.Error(fmt.Sprintf("Expected '%s' got '%s'", testData.expectedValue, cell.Value))
			}
			if cell.GetStyle().ApplyFill != testData.expectedApplyFill {
				t.Error("Expected 'ApplyFill' value to be ", map[bool]string{true: "truthy", false: "falsy"}[testData.expectedApplyFill])
			}
			if err != nil {
				t.Error("Found error: ", err)
			}
		})
	}
}
