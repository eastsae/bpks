// Package bpks implements a B+Tree Key Store that stores key/value pairs on an underlying io.ReadWriteSeeker device.
package bpks

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

// BPKS (B+Tree Key Store) is a key-value store based around a B+Tree
type BPKS struct {
	Device io.ReadWriteSeeker
	Root   *IndexBlock
}

// BPKSHeader is the byte array "BPKS" plus a major version (0x00, 0x01)
var BPKSHeader = []byte{0x42, 0x50, 0x4b, 0x53, 0x0, 0x1}

var firstFreeBlock uint64 = 2

// New returns a new BPKS attached to the specified io.ReadWriteSeeker
func New(device io.ReadWriteSeeker) *BPKS {
	return &BPKS{
		Device: device,
	}
}

// Mount mounts the BPKS keystore on the attached device. An error is returned if the
// device does not contain a formatted BPKS keystore.
func (me *BPKS) Mount() error {
	// Check Header
	_, err := me.Device.Seek(0, 0)
	if err != nil {
		return err
	}
	var buf = make([]byte, 6)
	_, err = me.Device.Read(buf)
	if err != nil {
		return err
	}
	if bytes.Compare(buf, BPKSHeader) != 0 {
		return errors.New("Not a BPKS device")
	}

	// Load Index Block
	root, err := me.LoadIndexBlock(2)
	if err != nil {
		return err
	}
	me.Root = root
	return nil
}

// Format initialises a new BPKS keystore on the attached ReadWriteSeeker. This
// will erase all keys and values from an existing keystore.
func (me *BPKS) Format() error {
	// Header
	_, err := me.Device.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = me.Device.Write(BPKSHeader)
	if err != nil {
		return err
	}

	// TODO: SpaceBPKS

	// Root Index Block
	me.Root = &IndexBlock{
		BPKS:         me,
		BlockAddress: 2,
	}
	return me.SaveIndexBlock(me.Root)
}

// Allocate gets the block address of the first free block on the device and marks it used.
func (me *BPKS) Allocate() uint64 {
	firstFreeBlock++
	return firstFreeBlock
}

// Deallocate frees the specified block address for reuse.
func (me *BPKS) Deallocate(blockAddress uint64) {
	panic("Not Implemented")
}

// Add writes the specified KeyPointer to the keystore
func (me *BPKS) Add(kp KeyPointer) {
	me.Root.Add(kp)
}

// Find finds the specified Key in the keystore, returning its KeyPointer if found, or
// an empty KeyPointer and false if not
func (me *BPKS) Find(key Key) (KeyPointer, bool, error) {
	return me.Root.Find(key)
}

func (me *BPKS) LoadIndexBlock(blockAddress uint64) (*IndexBlock, error) {
	fmt.Printf("Loading Index Block at address %d (offset %d)\n", blockAddress, blockAddress*4096)
	_, err := me.Device.Seek(int64(blockAddress*4096), 0)
	if err != nil {
		return nil, err
	}
	buffer := [4096]byte{}
	fmt.Printf("- Reading 4096 Bytes\n")
	c, err := me.Device.Read(buffer[:])
	if err != nil {
		return nil, err
	}
	fmt.Printf("- Init Index Block from buffer len %d\n", c)
	return NewIndexBlockFromBuffer(me, blockAddress, buffer[:]), nil
}

func (me *BPKS) SaveIndexBlock(block *IndexBlock) error {
	_, err := me.Device.Seek(int64(block.BlockAddress), 0)
	if err != nil {
		return err
	}
	return nil
}
