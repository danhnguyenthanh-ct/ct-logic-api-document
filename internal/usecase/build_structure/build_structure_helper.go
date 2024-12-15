package buildstructure

func mergeObject(map1, map2 map[string]any) map[string]any {
	merged := make(map[string]interface{})
	merged["type"] = "object"

	props1, _ := map1["properties"].(map[string]interface{})
	props2, _ := map2["properties"].(map[string]interface{})

	mergedProps := make(map[string]interface{})
	for k, v := range props1 {
		vTyped := inferType(v)
		if vTyped == "object" {
			if v2, ok := props2[k]; ok {
				if v2Object, ok := v2.(map[string]interface{}); ok {
					vObject := v.(map[string]interface{})
					if len(vObject) > 0 && len(v2Object) > 0 {
						mergedProps[k] = mergeObject(vObject, v2Object)
						continue
					}
				}
			}
		} else if vTyped == "null" {
			continue
		}
		mergedProps[k] = v
	}
	for k, v := range props2 {
		if vTyped := inferType(v); vTyped == "object" {
			vObject := v.(map[string]interface{})
			if len(vObject) > 0 {
				mergedProps[k] = v
			}
			continue
		}
		mergedProps[k] = v
	}
	merged["properties"] = mergedProps
	return merged
}

func generateSchemaForArray(data []any) map[string]any {
	if len(data) == 0 {
		return nil
	}
	typeOfElement := inferType(data[0])
	if typeOfElement == "null" {
		return nil
	} else if typeOfElement == "object" {
		currentSchema := generateSchema(data[0].(map[string]any))
		for i := 1; i < len(data); i++ {
			currentSchema = mergeObject(currentSchema, generateSchema(data[i].(map[string]any)))
		}
		return map[string]any{
			"type":  "array",
			"items": currentSchema,
		}
	} else {
		return map[string]any{
			"type":  "array",
			"items": map[string]any{"type": typeOfElement},
		}
	}
}

func generateSchema(data map[string]any) map[string]any {
	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	properties := schema["properties"].(map[string]any)
	for key, value := range data {
		typeOfField := inferType(value)
		if typeOfField == "null" {
			continue
		}
		if typeOfField == "object" {
			if schemaForObject := generateSchema(value.(map[string]any)); schemaForObject != nil {
				properties[key] = schemaForObject
			}
		} else if typeOfField == "array" {
			// Assume array contains only one type of element
			if len(value.([]any)) == 0 {
				continue
			}
			typeOfElement := inferType(value.([]any)[0])
			if typeOfElement == "object" {
				if schemaForObjectsArray := generateSchemaForArray(value.([]any)); schemaForObjectsArray != nil {
					properties[key] = schemaForObjectsArray
				}
			} else if typeOfElement == "null" {
				continue
			} else {
				properties[key] = map[string]any{
					"type":  "array",
					"items": map[string]any{"type": typeOfElement},
				}
			}
		} else {
			properties[key] = map[string]any{
				"type": typeOfField,
			}
		}
	}
	return schema
}

func inferType(value any) string {
	switch value.(type) {
	case string:
		return "string"
	case float64: // JSON numbers are float64 by default
		return "number"
	case int:
		return "integer"
	case bool:
		return "boolean"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	default:
		return "null"
	}
}
