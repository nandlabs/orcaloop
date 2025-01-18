package models

import "oss.nandlabs.io/golly/rest/server"

type Orcaloop struct {
	// The name of the service
	Name string `json:"name" yaml:"name"`
	// Storage configuration
	Storage *StorageConfig `json:"storage" yaml:"storage"`
	// Listener configuration
	Listener *server.Options `json:"listener" yaml:"listener"`
}

type StorageConfig struct {
	// The name of the storage
	Type string `json:"name" yaml:"name"`
	// The configuration of the storage
	Provider *Provider `json:"provider" yaml:"provider"`
}

type Provider struct {
	// Local storage configuration
	Local *LocalStorage `json:"local" yaml:"local"`
	// MongoDB storage configuration
	Mongo *MongoStorage `json:"mongo" yaml:"mongo"`
	// SQL storage configuration
	SQL *SQLStorage `json:"sql" yaml:"sql"`
}

type LocalStorage struct {
	// The name of the storage
	PurgeTimeout int `json:"purgeTime" yaml:"purgeTime"`
}

type MongoStorage struct {
	// The name of the storage
	ConnectionString string `json:"connectionString" yaml:"connectionString"`
	// The name of the database
	Database string `json:"database" yaml:"database"`
}

type SQLStorage struct {
	// The name of the storage
	ConnectionString string `json:"connectionString" yaml:"connectionString"`
	// The name of the database
	Database string `json:"database" yaml:"database"`
}

type FirestoreStorage struct {
	// The name of the storage
	ConnectionString string `json:"connectionString" yaml:"connectionString"`
	// The name of the database
	Database string `json:"database" yaml:"database"`
}
