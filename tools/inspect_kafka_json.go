package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

func main() {
	filePath := "bundle/kafka.json"
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer f.Close()

	fmt.Printf("Analyzing %s...\n", filePath)

	// Use a raw decoder to avoid buffering everything if possible,
	// though for structure analysis we often need to decode generic interfaces.
	decoder := json.NewDecoder(f)

	// Peek at the first token
	t, err := decoder.Token()
	if err != nil {
		panic(err)
	}

	if delim, ok := t.(json.Delim); ok {
		if delim == '[' {
			fmt.Println("Root Structure: ARRAY")
			analyzeArray(decoder)
		} else if delim == '{' {
			fmt.Println("Root Structure: OBJECT")
			analyzeObject(decoder)
		} else {
			fmt.Printf("Unexpected delimiter: %v\n", delim)
		}
	} else {
		fmt.Printf("Root is a primitive value: %v\n", t)
	}
}

func analyzeArray(decoder *json.Decoder) {
	fmt.Println("Scanning all items in root array...")
	fmt.Printf("%-20s | %-10s | %s\n", "Name", "Error", "Response Keys (Top Level)")
	fmt.Println("--------------------------------------------------------------------------------")

	count := 0
	for decoder.More() {
		// partial decode just to get structure summary
		var item map[string]interface{}
		if err := decoder.Decode(&item); err != nil {
			fmt.Printf("Error decoding item %d: %v\n", count, err)
			return
		}

		name, _ := item["Name"].(string)
		errField := item["Error"]
		response, _ := item["Response"].(map[string]interface{})

		// Gather keys from response for context
		respKeys := ""
		if response != nil {
			i := 0
			for k := range response {
				if i > 0 {
					respKeys += ", "
				}
				if i > 4 {
					respKeys += "..."
					break
				}
				respKeys += k
				i++
			}
		} else if item["Response"] != nil {
            respKeys = fmt.Sprintf("(%T)", item["Response"])
        }

		fmt.Printf("%-20s | %-10v | %s\n", name, errField, respKeys)
		count++
	}
	fmt.Printf("\nTotal items: %d\n", count)
}

func analyzeObject(decoder *json.Decoder) {
	// We are inside an object (after '{')
	// We read key-value pairs
	fmt.Println("Top-level Keys:")
	for decoder.More() {
		t, err := decoder.Token()
		if err != nil {
			break
		}
		key := fmt.Sprintf("%v", t)

		var value interface{}
		if err := decoder.Decode(&value); err != nil {
			fmt.Printf("  %s: Error decoding value\n", key)
			continue
		}

		fmt.Printf("  %-25s : %s\n", key, summarizeValue(value))
	}
}

func summarizeValue(v interface{}) string {
	if v == nil {
		return "null"
	}
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Map:
		// It's a nested object, list keys
		keys := val.MapKeys()
		keyStr := ""
		for i, k := range keys {
			if i > 0 {
				keyStr += ", "
			}
			if i > 5 {
				keyStr += "..."
				break
			}
			keyStr += fmt.Sprintf("%v", k)
		}
		return fmt.Sprintf("Object {%s}", keyStr)
	case reflect.Slice, reflect.Array:
		return fmt.Sprintf("Array (len=%d)", val.Len())
	default:
		// Primitive
        str := fmt.Sprintf("%v", v)
        if len(str) > 50 {
            str = str[:47] + "..."
        }
		return fmt.Sprintf("%T (%v)", v, str)
	}
}

func printSummary(v interface{}, indent string) {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Map:
		keys := val.MapKeys()
		for _, k := range keys {
			subVal := val.MapIndex(k).Interface()
			fmt.Printf("%s%-20v: %s\n", indent, k, summarizeValue(subVal))
		}
	default:
		fmt.Printf("%s%v\n", indent, summarizeValue(v))
	}
}
