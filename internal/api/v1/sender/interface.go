package sender

import (
	"context"
	"net"
)

type Sender interface {
	AddNewServer(ctx context.Context, ip net.IP) error
}
