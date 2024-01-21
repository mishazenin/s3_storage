package config

import (
	"net"
)

type SenderConfig struct {
	AvailableServers []net.IP `envconfig:"SENDER_SERVERS" default:""`
}
