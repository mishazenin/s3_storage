package dataService

import (
	"net"
	"sync/atomic"
)

type Object struct {
	hash   string
	shards []objShard
}

type objShard struct {
	parentHash string
	payload    []byte
	hash       string
	sequence   uint8
	size       uint
}

type shard struct {
	ip       net.IP
	isAlive  atomic.Bool
	id       string
	location string
	// ...
}
