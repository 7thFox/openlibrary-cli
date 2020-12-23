package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type BookInfo struct {
	// Flat fields
	Weight        string `json:"weight"`
	Title         string `json:"title"`
	URL           string `json:"url"`
	NumberOfPages int    `json:"number_of_pages"`
	PublishDate   string `json:"publish_date"`

	// Flat-ish fields
	Identifiers struct {
		Google           []string `json:"google"`
		LCCN             []string `json:"lccn"`
		ISBN13           []string `json:"isbn_13"`
		Amazon           []string `json:"amazon"`
		ISBN10           []string `json:"isbn_10"`
		OCLC             []string `json:"oclc"`
		LibraryThing     []string `json:"librarything"`
		ProjectGutenberg []string `json:"project_gutenberg"`
		Goodreads        []string `json:"goodreads"`
	} `json:"identifiers"`

	Classifications struct {
		DeweyDecimalClass []string `json:"dewey_decimal_class"`
		LCClassifications []string `json:"lc_classifications"`
	} `json:"classifications"`

	Cover struct {
		Small  string `json:"small"`
		Large  string `json:"large"`
		Medium string `json:"medium"`
	} `json:"cover"`

	// Array-type fields
	Publishers []struct {
		Name string `json:"name"`
	} `json:"publishers"`

	Links []struct {
		URL   string `json:"url"`
		Title string `json:"title"`
	} `json:"links"`

	Subjects []struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"subjects"`

	Authors []struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"authors"`

	Excerpts []struct {
		Comment string `json:"comment"`
		Text    string `json:"text"`
	} `json:"excerpts"`

	PublishPlaces []struct {
		Name string `json:"name"`
	} `json:"publish_places"`
}

var (
	fieldMap map[string]reflect.StructField
)

func getField(v reflect.Value, prefix string, parts []string) string {
	if fieldMap == nil {
		fieldMap = make(map[string]reflect.StructField)
		appendToMap(v.Type(), "")
	}
	if len(parts) == 0 {
		return fmt.Sprint(v.Interface())
	}
	head, tail := prefix+parts[0], parts[1:]
	if f, ok := fieldMap[head]; ok {
		k := f.Type.Kind()
		_ = k
		if f.Type.Kind() == reflect.Slice {
			return getArrayField(
				v.FieldByIndex(f.Index),
				head+".",
				tail)
		}
		return getField(
			v.FieldByIndex(f.Index),
			head+".",
			tail)
	}
	return ""
}

func getArrayField(arr reflect.Value, prefix string, parts []string) string {
	if arr.Len() == 0 {
		return ""
	}
	arrayHandle, tail := parts[0], parts[1:]

	i, err := strconv.Atoi(arrayHandle)
	if err == nil {
		if i < 0 {
			i = arr.Len() + i
		}
		if i >= 0 && i < arr.Len() {
			return getField(arr.Index(i), prefix, tail)
		}
		return ""
	}

	switch strings.ToLower(arrayHandle) {
	case "fst", "first", "head":
		return getField(arr.Index(0), prefix, tail)
	case "lst", "last", "tail":
		return getField(arr.Index(arr.Len()-1), prefix, tail)
	case "csv", "*", "all":
		var sb strings.Builder
		for i := 0; i < arr.Len(); i++ {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(getField(arr.Index(i), prefix, tail))
		}
		return sb.String()
	}

	colorPrintln(colorYellow, "WARN: Unknown array handle; defaulting to first.")
	return getField(arr.Index(0), prefix, tail)
}

func appendToMap(t reflect.Type, prefix string) {
	if t.Kind() != reflect.Struct {
		return
	}
	nf := t.NumField()
	for i := 0; i < nf; i++ {
		f := t.Field(i)
		if jsonName, ok := f.Tag.Lookup("json"); ok {
			fieldMap[prefix+jsonName] = f
			if f.Type.Kind() == reflect.Slice {
				appendToMap(f.Type.Elem(), prefix+jsonName+".")
			} else if f.Type.Kind() == reflect.Struct {
				appendToMap(f.Type, prefix+jsonName+".")
			}
		}
	}
}
