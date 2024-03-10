package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/rivo/tview"
)

type PathAndPathItem struct {
	Path     string
	Method   string
	PathItem *openapi3.PathItem
}

func (p PathAndPathItem) AsFuzzyEntry() string {
	return fmt.Sprintf("%-8s%s", p.Method, p.Path)
}

func (p PathAndPathItem) GetRequestBody() *openapi3.RequestBodyRef {
	if p.PathItem.Get != nil {
		return p.PathItem.Get.RequestBody
	} else if p.PathItem.Post != nil {
		return p.PathItem.Post.RequestBody
	} else if p.PathItem.Put != nil {
		return p.PathItem.Put.RequestBody
	} else if p.PathItem.Delete != nil {
		return p.PathItem.Delete.RequestBody
	} else {
		return nil
	}
}

func bailOnError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: ./main <HOST|path-to-openapi.json>")
		os.Exit(1)
	}
	specPath := os.Args[1]

	var remoteHostOrigin string = ""

	// Load the OpenAPI spec
	loader := openapi3.NewLoader()
	var doc *openapi3.T
	var err error

	// detect if specPath is a URL
	if strings.HasPrefix(specPath, "http") {
		// if specPath doesn't end in .json, append openapi.json to it
		var loadUrl string
		if !strings.HasSuffix(specPath, ".json") {
			loadUrl = fmt.Sprintf("%s/openapi.json", specPath)
		} else {
			loadUrl = specPath
		}

		parsedURL, err := url.Parse(loadUrl)
		if err != nil {
			fmt.Println("Error parsing URL:", err)
			return
		}
		remoteHostOrigin = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

		doc, err = loader.LoadFromURI(parsedURL)
		bailOnError(err)
		if err != nil {
			fmt.Printf("Error loading OpenAPI spec: %v\n", err)
			os.Exit(1)
		}
	} else {
		doc, err = loader.LoadFromFile(specPath)
		bailOnError(err)
	}

	// Generate options for the fuzzy finder
	var options []PathAndPathItem
	for path, pathItem := range doc.Paths.Map() {
		if pathItem == nil {
			continue
		}

		// Iterate over operations defined in the pathItem
		for _, operation := range pathItem.Operations() {
			if operation != nil {
				// description := operation.Summary
				// option := fmt.Sprintf("%s\t%s\t%s", strings.ToUpper(method), path, description)

				var method string
				if pathItem.Get != nil {
					method = "GET"
				} else if pathItem.Post != nil {
					method = "POST"
				} else if pathItem.Put != nil {
					method = "PUT"
				} else if pathItem.Delete != nil {
					method = "DELETE"
				} else {
					continue
				}

				pathAndPathItem := PathAndPathItem{
					Path:     path,
					Method:   method,
					PathItem: pathItem,
				}
				options = append(options, pathAndPathItem)
			}
		}
	}

	selectedIndex := fuzzyFind(options)
	pathAndPathItem := options[selectedIndex]
	// WARN: this really only works for POST due to the curl construction
	if pathAndPathItem.GetRequestBody() != nil {
		getRequestBody(remoteHostOrigin, &pathAndPathItem)
	}
}

func fuzzyFind(pathAndPathItem []PathAndPathItem) int {
	idx, err := fuzzyfinder.Find(
		pathAndPathItem,
		func(i int) string {
			return pathAndPathItem[i].AsFuzzyEntry()
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("PATH: %s", pathAndPathItem[i].Path)
		}))
	if err != nil {
		log.Fatal(err)
	}
	return idx
}

func parseJson(jsonString string) interface{} {
	var jsonObject interface{}
	err := json.Unmarshal([]byte(jsonString), &jsonObject)
	if err != nil {
		fmt.Println("Error unmarshalling JSON", err)
		return nil
	}
	return jsonObject
}

func asJsonString(jsonObject *interface{}) string {
	jsonString, err := json.MarshalIndent(jsonObject, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON", err)
		return ""
	}
	return string(jsonString)
}

func reflowJsonString(jsonString string) string {
	jsonObject := parseJson(jsonString)
	return asJsonString(&jsonObject)
}

func getRequestBody(remoteHostOrigin string, pathAndPathItem *PathAndPathItem) {
	// Assuming 'application/json' content type for simplicity
	requestBody := pathAndPathItem.GetRequestBody()
	if requestBody == nil {
		fmt.Println("The selected operation does not have a request body.")
		return
	}

	mediaType := requestBody.Value.Content.Get("application/json")
	if mediaType == nil {
		fmt.Println("The selected operation does not support 'application/json'.")
		return
	}

	schema := mediaType.Schema.Value
	properties := schema.Properties

	app := tview.NewApplication()
	form := tview.NewForm()

	// Map to store form input values
	requestBodyData := make(map[string]interface{})

	// iterate over the properties and collect them into fields
	for propName, propSchemaRef := range properties {
		propSchema := propSchemaRef.Value

		if propSchema.Type == "string" || propSchema.Type == "integer" || propSchema.Type == "number" || propSchema.Type == "boolean" {

			var acceptanceFunc func(text string, ch rune) bool
			switch propSchema.Type {
			case "string":
				acceptanceFunc = nil
			case "boolean":
				acceptanceFunc = func(text string, lastChar rune) bool {
					return text == "true" || text == "false"
				}
			case "integer":
				acceptanceFunc = tview.InputFieldInteger
			case "number":
				acceptanceFunc = tview.InputFieldFloat
			}
			// primitive type
			form.AddInputField(
				fmt.Sprintf("%s (%s)\n", propName, propSchema.Type),
				"",
				0,
				acceptanceFunc,
				(func(propName string) func(string) {
					return func(text string) {
						var err error
						if propSchema.Type == "boolean" {
							requestBodyData[propName] = text == "true"
						} else if propSchema.Type == "integer" {
							requestBodyData[propName], err = strconv.ParseInt(text, 10, 64)
							bailOnError(err)
						} else if propSchema.Type == "number" {
							requestBodyData[propName], err = strconv.ParseFloat(text, 64)
							bailOnError(err)
						} else {
							requestBodyData[propName] = text
						}
					}
				})(propName),
			)
		} else {
			// compound type
			jsonBytes, err := propSchema.MarshalJSON()
			bailOnError(err)
			reflowedJsonString := reflowJsonString(string(jsonBytes))
			numLines := strings.Count(reflowedJsonString, "\n") + 1
			form.AddTextView(
				fmt.Sprintf("%s (%s)\n", propName, propSchema.Type),
				reflowedJsonString,
				0, numLines, true, true)
			form.AddTextArea(
				"",
				"",
				0,
				0,
				0,
				(func(propName string) func(text string) {
					return func(text string) {
						requestBodyData[propName] = text
					}
				})(propName),
			)
		}

	}

	form.AddButton("Save", func() {
		app.Stop()

		// postprocessing: coerce compound values
		for propName, propSchemaRef := range properties {
			propSchema := propSchemaRef.Value
			if propSchema.Type == "object" || propSchema.Type == "array" {
				requestBodyData[propName] = parseJson(requestBodyData[propName].(string))
			}
		}

		var curlHost string
		if remoteHostOrigin == "" {
			curlHost = "$TARGET"
		} else {
			curlHost = remoteHostOrigin
		}

		jsonString, err := json.Marshal(requestBodyData)
		if err != nil {
			fmt.Println("Error marshalling JSON")
			return
		}
		curlCommand := fmt.Sprintf(
			"curl -v -X %s %s%s -H 'Content-Type: application/json' -d '%s'",
			pathAndPathItem.Method,
			curlHost,
			pathAndPathItem.Path,
			jsonString,
		)
		fmt.Println(color.CyanString(curlCommand))

		if remoteHostOrigin != "" {
			cmd := exec.Command("sh", "-c", curlCommand)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				fmt.Println("Error running curl command:", err)
			}
		}

	})
	form.AddButton("Cancel", func() {
		app.Stop()
	})

	form.SetBorder(true).SetTitle(
		pathAndPathItem.AsFuzzyEntry(),
	).SetTitleAlign(tview.AlignLeft)

	if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}