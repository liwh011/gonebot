package gonebot

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	bolt "go.etcd.io/bbolt"
)

func init() {}

type StorageFactory interface {
	CreateStorage(dbName string) (Storage, error)
}

type Storage interface {
	GetBucket(key string) Bucket
	SetAutoCreateBucket(autoCreate bool)
	Batch(actionFunc func(pBatch StorageBatch)) error
	Set(bucketKey, key string, value any) error
	Get(bucketKey, key string, outValuePointer any) error
	Delete(bucketKey, key string) error
	DeleteBucket(bucketKey string) error
}

type Bucket interface {
	Set(key string, value any) error
	Get(key string, outValuePointer any) error
	Delete(key string) error
	DeleteBucket() error
	Batch(actionFunc func(pBatch BucketBatch)) error
}

type StorageBatch interface {
	Set(bucketKey, key string, value any)
	Get(bucketKey, key string, outValuePointer any)
	Delete(bucketKey, key string)
	DeleteBucket(bucketKey string)
}

type BucketBatch interface {
	Set(key string, value any)
	Get(key string, outValuePointer any)
	Delete(key string)
	DeleteBucket()
}

var storageFactory StorageFactory = &boltDBFactory{}

func SetStorageFactory(factory StorageFactory) {
	storageFactory = factory
}

func NewStorage(dbName string) (Storage, error) {
	return storageFactory.CreateStorage(dbName)
}

type boltDBFactory struct{}

type boltStorage struct {
	dbName           string
	dbConnection     *bolt.DB
	autoCreateBucket bool
}

type boltBucket struct {
	storage *boltStorage
	key     string
}

type boltStorageBatch struct {
	storage      *boltStorage
	funcs        []func(tx *bolt.Tx) error
	bucketsCache map[string]*bolt.Bucket
}

type boltBucketBatch struct {
	bucket *boltBucket
	sb     *boltStorageBatch
}

func (f *boltDBFactory) CreateStorage(dbName string) (Storage, error) {
	if !strings.HasSuffix(dbName, ".db") {
		dbName += ".db"
	}
	connection, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("打开%s数据库文件发生错误：%v", dbName, err)
	}
	return &boltStorage{
		dbName:           dbName,
		dbConnection:     connection,
		autoCreateBucket: true,
	}, nil
}

func (s *boltStorage) GetBucket(key string) Bucket {
	return &boltBucket{
		storage: s,
		key:     key,
	}
}

// 访问到的bucket不存在时是否自动创建
func (s *boltStorage) SetAutoCreateBucket(autoCreate bool) {
	s.autoCreateBucket = autoCreate
}

// 获取DB中的bucket
func (s *boltStorage) getDBBucket(tx *bolt.Tx, key string) *bolt.Bucket {
	bucket := tx.Bucket([]byte(key))
	if bucket != nil {
		return bucket
	}

	if s.autoCreateBucket {
		if tx.Writable() {
			bucket, _ = tx.CreateBucketIfNotExists([]byte(key))
		}
	}

	return bucket
}

func (s *boltStorage) Set(bucketKey, key string, value any) error {
	return s.dbConnection.Update(func(tx *bolt.Tx) error {
		targetBucket := s.getDBBucket(tx, bucketKey)
		if targetBucket == nil {
			return fmt.Errorf("bucket %s 不存在", bucketKey)
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(value)
		if err != nil {
			return err
		}
		return targetBucket.Put([]byte(key), buf.Bytes())
	})
}

func (s *boltStorage) Get(bucketKey, key string, outValuePointer any) error {
	return s.dbConnection.View(func(tx *bolt.Tx) error {
		targetBucket := s.getDBBucket(tx, bucketKey)
		if targetBucket == nil {
			return fmt.Errorf("bucket %s 不存在", bucketKey)
		}

		data := targetBucket.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("key %s 不存在", key)
		}

		dec := gob.NewDecoder(bytes.NewReader(data))
		err := dec.Decode(outValuePointer)
		return err
	})
}

func (s *boltStorage) Delete(bucketKey, key string) error {
	return s.dbConnection.Update(func(tx *bolt.Tx) error {
		targetBucket := s.getDBBucket(tx, bucketKey)
		if targetBucket == nil {
			return fmt.Errorf("bucket %s 不存在", bucketKey)
		}

		return targetBucket.Delete([]byte(key))
	})
}

func (s *boltStorage) DeleteBucket(bucketKey string) error {
	return s.dbConnection.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(bucketKey))
	})
}

func (s *boltStorage) Batch(actionFunc func(batch StorageBatch)) error {
	batch := &boltStorageBatch{
		storage: s,
		funcs:   []func(tx *bolt.Tx) error{},
	}
	actionFunc(batch)
	return s.dbConnection.Update(func(tx *bolt.Tx) error {
		for _, f := range batch.funcs {
			err := f(tx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (batch *boltStorageBatch) getBucket(tx *bolt.Tx, bucketKey string) *bolt.Bucket {
	if bucket, ok := batch.bucketsCache[bucketKey]; ok {
		return bucket
	}
	targetBucket := batch.storage.getDBBucket(tx, bucketKey)
	batch.bucketsCache[bucketKey] = targetBucket
	return targetBucket
}

func (batch *boltStorageBatch) Set(bucketKey, key string, value any) {
	batch.funcs = append(batch.funcs, func(tx *bolt.Tx) error {
		targetBucket := batch.getBucket(tx, bucketKey)
		if targetBucket == nil {
			return fmt.Errorf("bucket %s 不存在", bucketKey)
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(value)
		if err != nil {
			return err
		}
		return targetBucket.Put([]byte(key), buf.Bytes())
	})
}

func (batch *boltStorageBatch) Get(bucketKey, key string, outValuePointer any) {
	batch.funcs = append(batch.funcs, func(tx *bolt.Tx) error {
		targetBucket := batch.getBucket(tx, bucketKey)
		if targetBucket == nil {
			return fmt.Errorf("bucket %s 不存在", bucketKey)
		}

		data := targetBucket.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("key %s 不存在", key)
		}

		dec := gob.NewDecoder(bytes.NewReader(data))
		return dec.Decode(outValuePointer)
	})
}

func (batch *boltStorageBatch) Delete(bucketKey, key string) {
	batch.funcs = append(batch.funcs, func(tx *bolt.Tx) error {
		targetBucket := batch.getBucket(tx, bucketKey)
		if targetBucket == nil {
			return fmt.Errorf("bucket %s 不存在", bucketKey)
		}

		return targetBucket.Delete([]byte(key))
	})
}

func (batch *boltStorageBatch) DeleteBucket(bucketKey string) {
	batch.funcs = append(batch.funcs, func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(bucketKey))
	})
}

func (b *boltBucket) Set(key string, value any) error {
	return b.storage.Set(b.key, key, value)
}

func (b *boltBucket) Get(key string, outValuePointer any) error {
	return b.storage.Get(b.key, key, outValuePointer)
}

func (b *boltBucket) Delete(key string) error {
	return b.storage.Delete(b.key, key)
}

func (b *boltBucket) DeleteBucket() error {
	return b.storage.DeleteBucket(b.key)
}

func (b *boltBucket) Batch(actionFunc func(batch BucketBatch)) error {
	return b.storage.Batch(func(batch StorageBatch) {
		actionFunc(&boltBucketBatch{
			sb:     batch.(*boltStorageBatch),
			bucket: b,
		})
	})
}

func (batch *boltBucketBatch) Set(key string, value any) {
	batch.sb.Set(batch.bucket.key, key, value)
}

func (batch *boltBucketBatch) Get(key string, outValuePointer any) {
	batch.sb.Get(batch.bucket.key, key, outValuePointer)
}

func (batch *boltBucketBatch) Delete(key string) {
	batch.sb.Delete(batch.bucket.key, key)
}

func (batch *boltBucketBatch) DeleteBucket() {
	batch.sb.DeleteBucket(batch.bucket.key)
}
