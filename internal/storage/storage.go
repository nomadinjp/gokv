package storage

import (
	"fmt"
	"os"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

// Storage handles the interaction with the underlying BadgerDB.
type Storage struct {
	db *badger.DB
}

// NewStorage initializes and opens the BadgerDB connection.
// It reads the database path from the DB_PATH environment variable,
// defaulting to "./data" if not set.
func NewStorage() (*Storage, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data"
	}

	opts := badger.DefaultOptions(dbPath)
	// Set up logging to be quiet unless there's an issue
	opts.Logger = nil 

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB at %s: %w", dbPath, err)
	}

	return &Storage{db: db}, nil
}

// Close gracefully closes the BadgerDB connection.
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Set creates or updates a key-value pair.
// The internal key is constructed as "bucket:key".
func (s *Storage) Set(bucket, key string, value []byte) error {
	internalKey := []byte(fmt.Sprintf("%s:%s", bucket, key))

	err := s.db.Update(func(txn *badger.Txn) error {
		// Set the value without a TTL (Time-To-Live)
		return txn.Set(internalKey, value)
	})

	if err != nil {
		return fmt.Errorf("failed to set key %s:%s: %w", bucket, key, err)
	}
	return nil
}

// Get retrieves the value associated with the given bucket and key.
// Returns badger.ErrKeyNotFound if the key does not exist.
func (s *Storage) Get(bucket, key string) ([]byte, error) {
	internalKey := []byte(fmt.Sprintf("%s:%s", bucket, key))
	var value []byte

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(internalKey)
		if err != nil {
			return err // Will return badger.ErrKeyNotFound if not found
		}

		// Copy the value from the item to a local slice
		value, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

// Delete removes the key-value pair associated with the given bucket and key.
// The internal key is constructed as "bucket:key".
func (s *Storage) Delete(bucket, key string) error {
	internalKey := []byte(fmt.Sprintf("%s:%s", bucket, key))

	err := s.db.Update(func(txn *badger.Txn) error {
		// Delete the key. BadgerDB's Delete operation is idempotent, 
		// meaning it succeeds even if the key does not exist.
		return txn.Delete(internalKey)
	})

	if err != nil {
		return fmt.Errorf("failed to delete key %s:%s: %w", bucket, key, err)
	}
	return nil
}

// ListBuckets retrieves a list of all unique bucket names by iterating over all keys.
func (s *Storage) ListBuckets() ([]string, error) {
	buckets := make(map[string]struct{})

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())
			
			// Keys are stored as "bucket:key". Find the first colon.
			if idx := strings.Index(key, ":"); idx != -1 {
				bucketName := key[:idx]
				buckets[bucketName] = struct{}{}
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	// Convert map keys to a slice
	result := make([]string, 0, len(buckets))
	for bucket := range buckets {
		result = append(result, bucket)
	}

	return result, nil
}

// ListKeys retrieves a list of all keys within a specific bucket.
func (s *Storage) ListKeys(bucket string) ([]string, error) {
	keys := make([]string, 0)
	prefix := []byte(fmt.Sprintf("%s:", bucket))

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix // Use prefix option to only iterate over keys in this bucket
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := string(item.Key())
			
			// Keys are stored as "bucket:key". Extract the key part.
			// The length of the prefix is len(bucket) + 1 (for the colon).
			keyName := key[len(prefix):]
			keys = append(keys, keyName)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list keys for bucket %s: %w", bucket, err)
	}

	return keys, nil
}
