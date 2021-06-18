package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

// Define here until finalised and added do dp-import/events
var categoryDimensionImport = `{
  "type": "record",
  "name": "cantabular-dataset-instance-started",
  "fields": [
    {"name": "dimension_id",   "type": "string"},
    {"name": "job_id", "type": "string"},
    {"name": "instance_id", "type": "string"}
  ]
}`

var CategoryDimensionImport = &avro.Schema{
	Definition: categoryDimensionImport,
}
