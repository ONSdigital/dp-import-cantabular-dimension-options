package schema

import (
	"github.com/ONSdigital/dp-kafka/v2/avro"
)

// Define here until finalised and added do dp-import/events
var categoryDimensionImport = `{
  "type": "record",
  "name": "cantabular-dataset-category-dimension-import",
  "fields": [
    {"name": "dimension_id",   "type": "string"},
    {"name": "job_id", "type": "string"},
    {"name": "instance_id", "type": "string"},
    {"name": "cantabular_blob", "type": "string"}
  ]
}`

var instanceComplete = `{
  "type": "record",
  "name": "cantabular-dataset-instance-complete",
  "fields": [
    {"name": "instance_id",     "type": "string"},
    {"name": "cantabular_blob", "type": "string"}
  ]
}`

var CategoryDimensionImport = &avro.Schema{
	Definition: categoryDimensionImport,
}

var InstanceComplete = &avro.Schema{
	Definition: instanceComplete,
}
