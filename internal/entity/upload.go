package entity

import "github.com/klauspost/reedsolomon"

type Metadata struct {
	ObjectName string
	Bucket     string

	UserID uint
	Prefix uint

	Encoder reedsolomon.Encoder

	sequence  uint8
	Size      uint
	ObjShards []ObjectShard

	// calculate
	// content type
	// Size
	// uuid
}

type ObjectShard struct {
	Sequence     uint8
	ServerID     string
	ObjShardHash string
	Size         uint
}
