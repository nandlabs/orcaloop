package config

import "oss.nandlabs.io/golly/rest"

const (
	// InMemoryStorageType represents the local storage type.
	InMemoryStorageType = "in-memory"
	// MongoStorageType represents the MongoDB storage type.
	MongoStorageType = "mongo"
	// PostGresStorageType represents the SQL storage type.
	PostGresStorageType = "postgres"
)

// Orcaloop represents the configuration for the Orcaloop service.
//
// Fields:
//
//	Name: The name of the service.
//	Storage: The storage configuration.
//	Listener: The listener configuration.
type Orcaloop struct {
	// The name of the service
	Name string `json:"name" yaml:"name"`
	// StorageConfig configuration
	StorageConfig *StorageConfig `json:"storage" yaml:"storage"`
	// ApiSrvConfig configuration
	ApiSrvConfig *rest.SrvOptions `json:"api_server" yaml:"api_server"`
}

// StorageConfig represents the configuration for a storage system.
// It includes the type of storage and the provider-specific configuration.
//
// Fields:
//   - Type: The name of the storage.
//   - Provider: The configuration of the storage provider.
type StorageConfig struct {
	// The name of the storage
	Type string `json:"name" yaml:"name"`
	// The configuration of the storage
	Provider *Provider `json:"provider" yaml:"provider"`
}

// Provider represents the configuration for different storage providers.
// It includes configurations for local storage, MongoDB storage, and SQL storage.
//
// Fields:
//   - Local: Local storage configuration (pointer to LocalStorage).
//   - Mongo: MongoDB storage configuration (pointer to MongoStorage).
//   - SQL: SQL storage configuration (pointer to SQLStorage).
type Provider struct {
	// Local storage configuration
	Local *LocalStorage `json:"local" yaml:"local"`
	// MongoDB storage configuration
	Mongo *MongoStorage `json:"mongo" yaml:"mongo"`
	// SQL storage configuration
	SQL *SQLStorage `json:"sql" yaml:"sql"`
}

// LocalStorage represents the configuration for local storage.
// Fields:
//
//	PurgeTimeout: The timeout duration (in seconds) after which the storage will be purged.
//	              This is represented as an integer and can be configured via JSON or YAML
//	              using the key "purgeTime".
type LocalStorage struct {
	// The name of the storage
	PurgeTimeout int `json:"purgeTime" yaml:"purgeTime"`
}

// MongoStorage represents the configuration for MongoDB storage.
// It includes the connection string and the database name.
//
// Fields:
//
//	ConnectionString: The connection string used to connect to the MongoDB instance.
//	Database: The name of the MongoDB database.
type MongoStorage struct {
	// The name of the storage
	ConnectionString string `json:"connectionString" yaml:"connectionString"`
	// The name of the database
	Database string `json:"database" yaml:"database"`
}

// SQLStorage represents the configuration for SQL storage.
// It contains the necessary details to connect to a SQL database.
//
// Fields:
//
//	ConnectionString: The connection string used to connect to the SQL database.
//	Database: The name of the database to be used.
type SQLStorage struct {
	// The name of the storage
	ConnectionString string `json:"connectionString" yaml:"connectionString"`
	// The name of the database
	Database string `json:"database" yaml:"database"`
}

// FirestoreStorage represents the configuration for Firestore storage.
//
// Fields:
//
//	ConnectionString: The connection string for the Firestore storage.
//	Database: The name of the Firestore database.
type FirestoreStorage struct {
	// The name of the storage
	ConnectionString string `json:"connectionString" yaml:"connectionString"`
	// The name of the database
	Database string `json:"database" yaml:"database"`
}

// DefaultConfig returns the default configuration for the Orcaloop service.
func DefaultConfig() *Orcaloop {
	restOptions := rest.DefaultSrvOptions()
	restOptions.Id = "orcaloop-api-server"
	restOptions.PathPrefix = "/api/v1"
	return &Orcaloop{
		Name: "Orcaloop",
		StorageConfig: &StorageConfig{
			Type:     InMemoryStorageType,
			Provider: &Provider{Local: &LocalStorage{PurgeTimeout: 60}},
		},
		ApiSrvConfig: restOptions,
	}
}
