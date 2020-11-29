package xlsxtemplator

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/tealeg/xlsx/v3"
)

var (
	templatePropertyRegex = regexp.MustCompile(`{{\s*([\w.]+)+(\.+[\w]+)*\s*}}`)
)

// RenderTemplateWithData Renders result file and saves it on a disk
func RenderTemplateWithData(data map[string]interface{}, templatePath string, resultPath string) error {
	file, err := openTemplate(templatePath)
	if err != nil {
		return err
	}

	err = iterateSheetsAndRows(file.Sheets, data)
	if err != nil {
		return err
	}

	return save(file, resultPath)
}

func openTemplate(path string) (file *xlsx.File, err error) {
	return xlsx.OpenFile(path)
}

func save(file *xlsx.File, path string) error {
	return file.Save(path)
}

func iterateSheetsAndRows(sheets []*xlsx.Sheet, data map[string]interface{}) error {
	for _, sheet := range sheets {
		err := sheet.ForEachRow(func(row *xlsx.Row) error {
			return renderRow(row, data)
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func renderRow(row *xlsx.Row, ctx map[string]interface{}) error {
	property, err := findPropertyInRow(row)
	if err != nil {
		return err
	}

	if propertyParts, slice, ok := findSliceInPropertyPath(ctx, property); ok {
		return renderSliceRow(row, propertyParts, slice)
	}

	return renderCellsFromRow(row, ctx)
}

func renderSliceRow(row *xlsx.Row, slicePropertyParts []string, slice []interface{}) error {
	partsReversed := reverseSlice(slicePropertyParts)

	for j, item := range slice {
		ctx := make(map[string]interface{})

		ctx[partsReversed[0]] = item

		for _, part := range partsReversed[1:] {
			ctx = map[string]interface{}{part: ctx}
		}

		if j == len(slice)-1 {
			return renderCellsFromRow(row, ctx)
		}

		newRow, err := row.Sheet.AddRowAtIndex(row.GetCoordinate())
		if err != nil {
			return err
		}

		err = copyCellsBetweenRows(row, newRow)
		if err != nil {
			return err
		}

		err = renderCellsFromRow(newRow, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func copyCellsBetweenRows(row *xlsx.Row, newRow *xlsx.Row) error {
	newRow.SetHeight(row.GetHeight())

	return row.ForEachCell(func(cell *xlsx.Cell) error {
		newCell := newRow.AddCell()
		newCell.SetStyle(cell.GetStyle())
		newCell.HMerge = cell.HMerge
		newCell.VMerge = cell.VMerge
		newCell.Hidden = cell.Hidden
		newCell.NumFmt = cell.NumFmt
		newCell.Value = cell.Value
		return nil
	})
}

func renderCellsFromRow(row *xlsx.Row, ctx map[string]interface{}) error {
	return row.ForEachCell(func(cell *xlsx.Cell) error {
		err := renderCell(cell, ctx)
		if err != nil {
			return err
		}
		return nil
	})
}

func renderCell(cell *xlsx.Cell, ctx map[string]interface{}) error {
	cell.Value = strings.Replace(cell.Value, "{{", "{{{", -1)
	cell.Value = strings.Replace(cell.Value, "}}", "}}}", -1)

	template, err := raymond.Parse(cell.Value)
	if err != nil {
		return err
	}

	cell.Value, err = template.Exec(ctx)
	if err != nil {
		return err
	}

	if cell.GetStyle().Fill.BgColor != "" {
		cell.GetStyle().ApplyFill = true
	}

	if cell.HMerge > 0 && cell.Value == "" {
		cell.Value = " "
	}

	return nil
}

func findSliceInPropertyPath(ctx map[string]interface{}, property string) ([]string, []interface{}, bool) {
	handledPropertyParts := make([]string, 0)

	if strings.Contains(property, ".") {
		parts := strings.Split(property, ".")

		data := ctx

		for _, part := range parts {
			handledPropertyParts = append(handledPropertyParts, part)

			valueByPropertyPart, ok := data[part]

			if !ok || valueByPropertyPart == nil {
				break
			}

			if reflect.TypeOf(valueByPropertyPart).Kind() == reflect.Slice {
				return handledPropertyParts, valueByPropertyPart.([]interface{}), true
			}

			data, _ = valueByPropertyPart.(map[string]interface{})
		}
	}

	return make([]string, 0), make([]interface{}, 0), false
}

func findPropertyInRow(row *xlsx.Row) (string, error) {
	foundProperty := ""
	err := row.ForEachCell(func(cell *xlsx.Cell) error {
		if cell.Value == "" {
			return nil
		}

		if match := templatePropertyRegex.FindAllStringSubmatch(cell.Value, -1); match != nil {
			foundProperty = match[0][1]
		}

		return nil
	})
	return foundProperty, err
}

func reverseSlice(items []string) []string {
	reversed := make([]string, 0, len(items))

	for i := len(items) - 1; i >= 0; i-- {
		reversed = append(reversed, items[i])
	}

	return reversed
}
