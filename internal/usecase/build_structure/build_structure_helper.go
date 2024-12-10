package buildstructure

func mergeMaps(map1, map2 map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range map1 {
		if v2, ok := map2[k]; ok {
			switch vTyped := v.(type) {
			case map[string]any:
				if v2Typed, ok := v2.(map[string]any); ok {
					result[k] = mergeMaps(vTyped, v2Typed)
				} else {
					result[k] = v
				}
			default:
				result[k] = v
			}
		} else {
			result[k] = v
		}
	}

	for k, v := range map2 {
		if _, ok := map1[k]; !ok {
			result[k] = v
		}
	}
	return result
}

func generateSchemaForArray(data []any) map[string]any {
	if len(data) == 0 {
		return map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "null",
			},
		}
	}
	typeOfElement := inferType(data[0])
	if typeOfElement != "object" {
		return map[string]any{
			"type":  "array",
			"items": map[string]any{"type": typeOfElement},
		}
	} else {
		currentSchema := generateSchema(data[0].(map[string]any))
		for i := 1; i < len(data); i++ {
			currentSchema = mergeMaps(currentSchema, generateSchema(data[i].(map[string]any)))
		}
		return map[string]any{
			"type":  "array",
			"items": currentSchema,
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
		if typeOfField == "object" {
			properties[key] = generateSchema(value.(map[string]any))
		} else if typeOfField == "array" {
			// Assume array contains only objects
			if len(value.([]any)) == 0 {
				properties[key] = map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "null",
					},
				}
				continue
			}
			typeOfElement := inferType(value.([]any)[0])
			if typeOfElement != "object" {
				properties[key] = map[string]any{
					"type":  "array",
					"items": map[string]any{"type": typeOfElement},
				}
			} else {
				properties[key] = generateSchemaForArray(value.([]any))
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
