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

	dec := json.NewDecoder(f)

	// Check for opening bracket
	if token, err := dec.Token(); err != nil || token != json.Delim('[') {
		fmt.Println("Expected array start")
		return
	}

	for dec.More() {
		var item struct {
			Name     string          `json:"Name"`
			Response json.RawMessage `json:"Response"`
			Error    interface{}     `json:"Error"`
		}
		if err := dec.Decode(&item); err != nil {
			fmt.Printf("Error decoding item: %v\n", err)
			return
		}

		fmt.Printf("Section: %s\n", item.Name)
		if item.Error != nil {
			fmt.Printf("  -> Error: %v\n", item.Error)
			continue
		}

		if item.Name == "metadata" {
			var metadata map[string]interface{}
			if err := json.Unmarshal(item.Response, &metadata); err == nil {
				if authOps, ok := metadata["AuthorizedOperations"]; ok {
					fmt.Printf("  -> Found 'AuthorizedOperations' (Top Level): %v\n", authOps)
				} else {
					fmt.Println("  -> 'AuthorizedOperations' not found at top level.")
				}
				
				// List other top level keys just in case
				fmt.Print("  -> All Metadata Keys: ")
				for k := range metadata {
					fmt.Printf("%s, ", k)
				}
				fmt.Println()
			}
		} else if item.Name == "topic_configs" {
			var configs []interface{}
			if err := json.Unmarshal(item.Response, &configs); err == nil {
				fmt.Printf("  -> topic_configs count: %d\n", len(configs))
				if len(configs) > 0 {
					fmt.Println("  -> Sample item 0 structure:")
					printStructure(configs[0], "    ")
				}
			}
		} else if item.Name == "broker_configs" {
			fmt.Println("  -> Found broker_configs section (currently unparsed by app)")
			// Since it errored in the user's bundle, we might not see response, but good to note.
		}
	}
}

func printStructure(v interface{}, indent string) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Map {
		for _, key := range val.MapKeys() {
			fmt.Printf("%sKey: %v, Type: %T\n", indent, key, val.MapIndex(key).Interface())
		}
	} else {
		fmt.Printf("%sValue: %v\n", indent, v)
	}
}
