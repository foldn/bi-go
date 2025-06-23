// internal/engine/fetcher.go
package engine

import "github.com/foldn/bi-go/internal/models"

// DataFetcher defines the interface that JoinStreamer needs to fetch data for its right-hand side.
// By depending on this interface, the engine package does not need to know about the service package.
type DataFetcher interface {
	FetchDataForEntity(jobUUID string, dataSourceID uint, entityName string, step []models.OperationStep) ([]map[string]interface{}, error)
}
