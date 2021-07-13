package event

// TODO: use dp-import struct when available
type CategoryDimensionImport struct {
	JobID          string `avro:"job_id"`
	InstanceID     string `avro:"instance_id"`
	DimensionID    string `avro:"dimension_id"`
	CantabularBlob string `avro:"cantabular_blob"`
}
