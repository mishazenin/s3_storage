package dataService

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/klauspost/reedsolomon"

	"karma_s3/internal/config"
	"karma_s3/internal/entity"
	"karma_s3/pkg/logger"
)

// objShard file len(available servers)
// send to particular shards
// guarantee replication factor
// add shards live probe, if it fails -> re-balance

// add saga

var ErrServerNotAlive = errors.New("shards is not alive")

// записать чанк на шард - вернуть айди шарда и хэш куска который записал и сохранить в метаданные
type Sender struct {
	sync.Mutex
	log              *logger.Logger
	availableServers []shard
	// Assuming your availableServers is sorted by shards IP
}

// func (s *Sender) chooseShard(key []byte) (shards, error) {
// 	s.Lock()
// 	defer s.Unlock()
// 	var aliveShard []shards
// 	for _, srv := range s.availableServers {
// 		if srv.isAlive.Load() {
// 			aliveShard = append(aliveShard, srv)
// 		}
// 	}
// 	if len(aliveShard) == 0 {
// 		return shards{}, ErrServerNotAlive
// 	}
// 	hashValue := hash(key)
// 	serverIndex := sort.Search(len(aliveShard), func(i int) bool {
// 		return bytes.Compare(aliveShard[i].ip, hashValue) >= 0
// 	})
// 	if serverIndex == len(aliveShard) {
// 		serverIndex = 0 // Wrap around if the hash value is greater than the highest shards IP.
// 	}
// 	return aliveShard[serverIndex], nil
// }
//
// // Hash the input key using FNV-1a algorithm.
// func hash(key []byte) net.IP {
// 	h := fnv.New32a()
// 	h.Write(key)
// 	hashValue := h.Sum32()
// 	// Convert the hash value to a 4-byte representation (IPv4)
// 	hashBytes := make([]byte, 4)
// 	binary.BigEndian.PutUint32(hashBytes, hashValue)
// 	return hashBytes
// }

func NewSender(
	log *logger.Logger,
	conf config.SenderConfig,
) (*Sender, error) {
	// TODO
	conf.AvailableServers = []net.IP{{}, {}, {}, {}, {}, {}}
	if len(conf.AvailableServers) < 1 {
		return nil, errors.New("servers are not set")
	}
	servers := make([]shard, len(conf.AvailableServers))
	for i := range conf.AvailableServers {
		s := shard{
			ip:      conf.AvailableServers[i],
			isAlive: atomic.Bool{},
		}
		s.isAlive.Store(true)
		servers[i] = s
	}
	return &Sender{
		log:              log,
		availableServers: servers,
	}, nil
}

type File struct {
	ID     string
	Chunks []objShard
}

func (s *Sender) GetObject(ctx context.Context, metadata entity.Metadata) (io.Reader, error) {
	return nil, nil
}
func (s *Sender) shardObject(ctx context.Context, object []byte, numberOfShards uint8) ([][]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	partSize := len(object) / int(numberOfShards)
	remainingBytes := len(object) % int(numberOfShards)
	parts := make([][]byte, numberOfShards)
	startIndex := 0
	for i := 0; i < int(numberOfShards); i++ {
		size := partSize
		if remainingBytes > 0 {
			size++
			remainingBytes--
		}
		endIndex := startIndex + size
		chunkPayload := object[startIndex:endIndex]
		if endIndex > len(object) {
			return nil, fmt.Errorf("invalid endIndex")
		}
		parts[i] = chunkPayload
		startIndex = endIndex
	}
	return parts, nil
}

// func (s *Sender) shardObject(ctx context.Context, file []byte, objHash string, numberOfChunks uint8) ([]objShard, error) {
// 	if err := ctx.Err(); err != nil {
// 		return nil, err
// 	}
// 	partSize := len(file) / int(numberOfChunks)
// 	remainingBytes := len(file) % int(numberOfChunks)
// 	parts := make([]objShard, numberOfChunks)
// 	startIndex := 0
// 	for i := 0; i < int(numberOfChunks); i++ {
// 		size := partSize
// 		if remainingBytes > 0 {
// 			size++
// 			remainingBytes--
// 		}
// 		endIndex := startIndex + size
// 		chunkPayload := file[startIndex:endIndex]
// 		h := sha256.New()
// 		h.Write(chunkPayload)
// 		if endIndex > len(file) {
// 			return nil, fmt.Errorf("invalid endIndex")
// 		}
// 		parts[i] = objShard{
// 			payload:    chunkPayload,
// 			hash:       string(h.Sum(nil)),
// 			parentHash: objHash,
// 			sequence:   uint8(i),
// 			size:       uint(len(chunkPayload)),
// 		}
// 		startIndex = endIndex
// 	}
// 	return parts, nil
// }

func (s *Sender) send(ctx context.Context, srv *shard, c objShard) error {
	if !srv.isAlive.Load() {
		return ErrServerNotAlive
	}
	cl := http.DefaultClient
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%d/api/v1/get_chunk", srv.ip), bytes.NewReader(c.payload))
	if err != nil {
		return err
	}
	resp, err := cl.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response code is: %d", resp.StatusCode)
	}
	return nil
}

const parityCoeff = 2

func (s *Sender) UploadObject(ctx context.Context, obj []byte, objHash string) (entity.Metadata, error) {
	ctx = context.Background()
	shards, err := s.shardObject(ctx, obj, uint8(len(s.availableServers)))
	if err != nil {
		return entity.Metadata{}, fmt.Errorf("objShard obj: %w", err)
	}
	shards, encoder, err := s.encodeWithParity(shards, parityCoeff)
	if err != nil {
		return entity.Metadata{}, fmt.Errorf("apply parity: %w", err)
	}
	objShardsMetadata, err := s.uploadObjShards(ctx, shards)
	if err != nil {
		return entity.Metadata{}, fmt.Errorf("send shards: %w", err)
	}
	return entity.Metadata{
		ObjectName: objHash,
		Encoder:    encoder,
		Size:       uint(len(obj)),
		ObjShards:  objShardsMetadata,
	}, nil
}

// implementing Erasure Coding https://en.wikipedia.org/wiki/Erasure_code
// // "reedsolomon" library is employed to simplify a significant amount of implementation code and hours of debugging.
// Therefore, I believe utilizing an encoder in metadata is not that bad.
func (s *Sender) encodeWithParity(shards [][]byte, parity uint) ([][]byte, reedsolomon.Encoder, error) {
	dataShards := len(shards)
	data := make([][]byte, dataShards)
	for i, sh := range shards {
		data[i] = make([]byte, len(sh))
		copy(data[i], sh)
	}
	enc, err := reedsolomon.New(4, int(parity))
	if err != nil {
		return nil, nil, fmt.Errorf("creating encoder: %v", err)
	}
	err = enc.Encode(data)
	if err != nil {
		return nil, nil, fmt.Errorf("encoding data: %v", err)
	}

	return data, enc, nil
}

func (s *Sender) decodeData(data [][]byte, enc reedsolomon.Encoder) ([][]byte, error) {
	// Simulate data loss by setting some shards to nil
	// for i := 0; i < 2; i++ {
	// 	data[i] = nil
	// }
	if err := enc.Reconstruct(data); err != nil {
		return nil, fmt.Errorf("decoding data: %w", err)
	}
	return data, nil
}

func getHash(obj []byte) string {
	hasher := sha256.New()
	hasher.Write(obj)
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}

func (s *Sender) uploadObjShards(ctx context.Context, data [][]byte) ([]entity.ObjectShard, error) {
	objShardMetadata := make([]entity.ObjectShard, 0, len(data))
	// TODO concurrently
	for i, shardData := range data {
		shardIndex := uint(i % len(s.availableServers)) // usштп the remainder of division for uniform distribution
		targetServerIP := s.availableServers[shardIndex].ip
		shardObjHash := getHash(shardData)
		if err := s.upload(ctx, shardData, shardObjHash, targetServerIP); err != nil {
			// A VERY basic implementation of the SAGA pattern: iterates over sent shards to execute compensate transactions
			for j := 0; j <= i; j++ {
				if err = s.removeObj(ctx, shardObjHash, targetServerIP); err != nil {
					// Technically, this data should be sent to a queue (a channel monitored by a daemon with a cron job for retries)
					//  to ensure that servers do not persistently store invalid data.
					// but I'll ignore it since it's just тестовое задание
					return nil, fmt.Errorf("failed to remove object: %s from server: %s error: %w", shardObjHash, targetServerIP.String(), err)
				}
			}
			return nil, err
		}
		objShardMetadata = append(objShardMetadata, entity.ObjectShard{
			Sequence:     uint8(i),
			ServerID:     targetServerIP.String(),
			ObjShardHash: shardObjHash,
			Size:         uint(len(shardData)),
		})
	}
	return objShardMetadata, nil
}

// add compensate func for saga

const mqxRetires = 2

var store = make(map[string][]byte)

// set GRPC
func (s *Sender) upload(ctx context.Context, data []byte, shardObjHash string, serverIP net.IP) error {
	store[shardObjHash] = data
	return nil
	// return backoff.Retry(func() error {
	// 	// there should be sent to message queue, instead of REST
	// 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/api/v1/upload/%s", serverIP.String(), shardObjHash), bytes.NewReader(data))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	cl := http.DefaultClient
	// 	resp, err := cl.Do(req)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if resp.StatusCode != http.StatusOK {
	// 		return fmt.Errorf("status code is: %d", resp.StatusCode)
	// 	}
	// 	return nil
	// }, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), mqxRetires))
}

func (s *Sender) removeObj(ctx context.Context, shardObjHash string, serverIP net.IP) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("http://%s/api/v1/upload/%s", serverIP.String(), shardObjHash), nil)
	if err != nil {
		return err
	}
	cl := http.DefaultClient
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code is: %d", resp.StatusCode)
	}
	return nil
}

func (s *Sender) AddNewServer(ctx context.Context, ip net.IP) error {
	s.Lock()
	defer s.Unlock()
	if !s.pingServer(ip) {
		return ErrServerNotAlive
	}
	isAlive := atomic.Bool{}
	isAlive.Store(true)
	s.availableServers = append(s.availableServers, shard{
		ip:      ip,
		isAlive: isAlive,
	})
	return nil
}

func (s *Sender) pingServer(srv net.IP) bool {
	return true
}

func (s *shard) ServersHeartBeat(ctx context.Context) error {
	return nil
}

// TODO
// parityShards - уточнить требования
// func (s *Sender) Sharding() {
// 	dataShards := 8
// 	parityShards := 4
// 	// Create a new encoder
// 	enc, err := reedsolomon.New(dataShards, parityShards)
// 	if err != nil {
// 		fmt.Println("Error creating encoder:", err)
// 		return
// 	}
// 	// Create some example data
// 	data := make([][]byte, dataShards)
// 	for i := range data {
// 		data[i] = make([]byte, 1024)
// 		for j := range data[i] {
// 			data[i][j] = byte(i + j)
// 		}
// 	}
// 	err = enc.Encode(data)
// 	if err != nil {
// 		fmt.Println("Error encoding data:", err)
// 		return
// 	}
// 	// Simulate data loss by setting some shards to nil
// 	for i := 0; i < 2; i++ {
// 		data[i] = nil
// 	}
//
// 	err = enc.Reconstruct(data)
// 	if err != nil {
// 		fmt.Println("Error decoding data:", err)
// 		return
// 	}
// 	for i := range data {
// 		fmt.Printf("Shard %d: %v\n", i, data[i])
// 	}
// }
