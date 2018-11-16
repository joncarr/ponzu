// Package db contains all interfaces to the databases used by Ponzu, including
// exported functions to easily manage addons, users, indices, search, content,
// and configuration.
package db

import (
	"github.com/jinzhu/gorm"
	"database/sql"
	"log"

	"github.com/ponzu-cms/ponzu/system/item"
	"github.com/ponzu-cms/ponzu/system/search"

	"github.com/boltdb/bolt"
	"github.com/nilslice/jwt"
)

var (
	db *gorm.DB

	tables = []string{
		"__config", "__users",
		"__addons", "__uploads",
		"__contentIndex",
	}

	tablesToAdd []string
)

// DB provides access to the underlying *gorm.DB database
func DB() *gorm.DB {
	return db
}

// Close exports the abillity to close our db. Should be called with defer
// after call to Init() from the same place.
func Close() {
	err := db.Close()
	if err != nil {
		log.Println(err)
	}
}

// Init creates a db connection, initializes db with required info, sets secrets
func Init() {
	if store != nil {
		return
	}

	// TODO: Build connection string based on .env file or environment vars
	var err error
	db, err = gorm.Open("postgres", connStr)
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Update(func(db *gorm.DB) error {
		// initialize db with all content type buckets & sorted bucket for type
		tx := db.Begin()
		for t := range item.Types {
			_, err := tx.CreateBucketIfNotExists([]byte(t))
			if err != nil {
				return err
			}

			_, err = tx.CreateBucketIfNotExists([]byte(t + "__sorted"))
			if err != nil {
				return err
			}
		}

		// init db with other tables as needed
		tables = append(tables, tablesToAdd...)

		for _, name := range tables {
			_, err := tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Fatalln("Coudn't initialize db with buckets.", err)
	}

	err = LoadCacheConfig()
	if err != nil {
		log.Fatalln("Failed to load config cache.", err)
	}

	clientSecret := ConfigCache("client_secret").(string)

	if clientSecret != "" {
		jwt.Secret([]byte(clientSecret))
	}

	// invalidate cache on system start
	err = InvalidateCache()
	if err != nil {
		log.Fatalln("Failed to invalidate cache.", err)
	}
}

// AddBucket adds a bucket to be created if it doesn't already exist
func AddBucket(name string) {
	bucketsToAdd = append(bucketsToAdd, name)
}

// InitSearchIndex initializes Search Index for search to be functional
// This was moved out of db.Init and put to main(), because addon checker was initializing db together with
// search indexing initialisation in time when there were no item.Types defined so search index was always
// empty when using addons. We still have no guarentee whatsoever that item.Types is defined
// Should be called from a goroutine after SetContent is successful (SortContent requirement)
func InitSearchIndex() {
	for t := range item.Types {
		err := search.MapIndex(t)
		if err != nil {
			log.Fatalln(err)
			return
		}
		SortContent(t)
	}
}

// SystemInitComplete checks if there is at least 1 admin user in the db which
// would indicate that the system has been configured to the minimum required.
func SystemInitComplete() bool {
	complete := false

	err := store.View(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte("__users"))
		if users == nil {
			return bolt.ErrBucketNotFound
		}

		err := users.ForEach(func(k, v []byte) error {
			complete = true
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		complete = false
		log.Fatalln(err)
	}

	return complete
}
