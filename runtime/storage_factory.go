package runtime

import (
	"fmt"
	"sync"

	"oss.nandlabs.io/orcaloop/config"
)

var store Storage
var mutex sync.Mutex

func GetStorage(c *config.StorageConfig) (s Storage, err error) {

	if store == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if store == nil {
			switch c.Type {
			case config.InMemoryStorageType:
				store, err = NewInMemoryStorage(c), nil
			// TODO implement other storage types
			// case config.MongoStorageType:
			// 	return NewMongoStorage(config)
			// case config.PostGresStorageType:
			// 	return NewPostGresStorage(config)
			default:
				err = fmt.Errorf("unknown storage type: %s", c.Type)
			}
		}
	}
	s = store
	return
}
