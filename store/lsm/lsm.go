package lsm

import "godis/store/inmem"

type lsm struct {
	memtable inmem.Inmem
}
