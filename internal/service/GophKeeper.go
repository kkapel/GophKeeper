package service

import (
	"github.com/kkapel/gophkeeper/internal/db/connections"
)

type GophKeeper struct {
	db *connections.DB
}
