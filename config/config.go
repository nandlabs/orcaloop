package config

import (
	"oss.nandlabs.io/golly/rest"
)

const (
	// InMemoryStorageType represents the local storage type.
	InMemoryStorageType = "in-memory"
	// MongoStorageType represents the MongoDB storage type.
	MongoStorageType = "mongo"
	// PostgresStorageType represents the SQL storage type.
	PostgresStorageType = "postgres"
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
	PostgreSQL *PostgresStorage `json:"sql" yaml:"sql"`
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

// PostgresStorage represents the configuration for SQL storage.
// It contains the necessary details to connect to a SQL database.
//
// Fields:
//
//	ConnectionString: The connection string used to connect to the SQL database.
//	Database: The name of the database to be used.
type PostgresStorage struct {
	// The name of the storage
	// if connection string is present, ignore the rest otherwise use the rest
	ConnectionString string `json:"connectionString,omitempty" yaml:"connectionString,omitempty"`
	// The name of the database
	Database string `json:"database,omitempty" yaml:"database,omitempty"`
	Host     string `json:"host,omitempty" yaml:"host,omitempty"`
	Port     int    `json:"port,omitempty" yaml:"port,omitempty"`
	Schema   string `json:"schema,omitempty" yaml:"schema,omitempty"`
	User     string `json:"user,omitempty" yaml:"user,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	SSLMode  string `json:"sslMode,omitempty" yaml:"sslMode,omitempty"`
	//Limits for the connection pool
	MaxLifetimeMs     int   `json:"maxLifetimeMs" yaml:"maxLifetimeMs"`
	MaxIdleTimeMs     int   `json:"maxIdleTimeMs" yaml:"maxIdleTimeMs"`
	MaxOpenConns      int   `json:"maxOpenConns" yaml:"maxOpenConns"`
	MaxIdleConns      int   `json:"maxIdleConns" yaml:"maxIdleConns"`
	WaitCount         int64 `json:"waitCount" yaml:"waitCount"`
	MaxIdleClosed     int64 `json:"maxIdleClosed" yaml:"maxIdleClosed"`
	MaxIdleTimeClosed int64 `json:"maxIdleTimeClosed" yaml:"maxIdleTimeClosed"`
	MaxLifetimeClosed int64 `json:"maxLifetimeClosed" yaml:"maxLifetimeClosed"`
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
	// return &Orcaloop{
	// 	Name: "Orcaloop",
	// 	StorageConfig: &StorageConfig{
	// 		Type:     InMemoryStorageType,
	// 		Provider: &Provider{Local: &LocalStorage{PurgeTimeout: 60}},
	// 	},
	// 	ApiSrvConfig: restOptions,
	// }
	return &Orcaloop{
		Name: "Orcaloop",
		StorageConfig: &StorageConfig{
			Type: PostgresStorageType,
			Provider: &Provider{
				PostgreSQL: &PostgresStorage{
					Host:     "localhost",
					Port:     5432,
					Database: "orcaloop-dev",
					User:     "pgadmin_user",
					Password: "pgadmin_password",
					SSLMode:  "disable",
				},
			},
		},
		ApiSrvConfig: restOptions,
	}
}
