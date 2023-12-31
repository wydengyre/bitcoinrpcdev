package gensite

import (
	"bytes"
	"html/template"
	"reflect"
	"time"
)

type btcTemplate template.Template

func mustBtcTemplate(name string, content string) *btcTemplate {
	return (*btcTemplate)(
		mustAddFooter(
			mustAddNav(
				template.Must(template.New(name).Parse(content)))))
}

const charsetTag = `<meta charset="utf-8" />`
const viewportTag = `<meta name="viewport" content="width=device-width, initial-scale=1">`
const iconTags = `<link rel="icon" href="/favicon.ico" sizes="32x32">
<link rel="apple-touch-icon" href="/apple-icon.png">`

var headTags = template.HTML(charsetTag + viewportTag + iconTags)

func (t *btcTemplate) render(d interface{}) ([]byte, error) {
	m, ok := d.(map[string]interface{})
	if !ok {
		m = structToMap(d)
	}
	m["headTags"] = headTags
	addFooterData(m)
	var buf bytes.Buffer
	err := (*template.Template)(t).Execute(&buf, m)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func structToMap(item interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	itemValue := reflect.ValueOf(item)

	// If the item is a pointer, get the element it points to
	if itemValue.Kind() == reflect.Ptr {
		itemValue = itemValue.Elem()
	}

	// Iterate over the fields of the struct
	for i := 0; i < itemValue.NumField(); i++ {
		// Get the field name
		fieldName := itemValue.Type().Field(i).Name
		// Get the field value
		fieldValue := itemValue.Field(i).Interface()
		if fieldValue == nil {
			// this will only happen if the field type was nil
			continue
		}
		// Add the field to the map
		result[fieldName] = fieldValue
	}

	return result
}

func nowStr() string {
	return time.Now().UTC().Format(time.RFC3339)
}
