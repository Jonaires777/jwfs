package filemanager

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/Jonaires777/src/constants"
)

func PrintBitmap() error {
	disk, err := os.Open(constants.VirtualDisk)
	if err != nil {
		return err
	}
	bitmap := make([]byte, constants.BitmapSize)
	_, err = disk.ReadAt(bitmap, constants.BitmapStart)
	if err != nil {
		return err
	}

	fmt.Println("Blocos ocupados:")

	for i := int64(0); i < constants.BitmapSize; i++ {
		for j := 0; j < 8; j++ {
			if (bitmap[i] & (1 << j)) != 0 {
				blockIndex := i*8 + int64(j)
				fmt.Printf(" %d", blockIndex)
			}
		}
	}

	fmt.Println() // Nova linha no final para melhor formatação
	return nil
}

func PrintSuperblock() error {
	disk, err := os.Open(constants.VirtualDisk)
	if err != nil {
		return err
	}
	defer disk.Close()

	data := make([]byte, 40) // Superbloco tem 40 bytes
	_, err = disk.ReadAt(data, constants.SuperBlockStart)
	if err != nil {
		return err
	}

	superblock := SuperBlock{
		DiskSize:        int64(binary.LittleEndian.Uint64(data[0:8])),
		MaxInodes:       int64(binary.LittleEndian.Uint64(data[8:16])),
		NumBlocks:       int64(binary.LittleEndian.Uint64(data[16:24])),
		InodeTableStart: int64(binary.LittleEndian.Uint64(data[24:32])),
		DataStart:       int64(binary.LittleEndian.Uint64(data[32:40])),
	}

	fmt.Println("Superblock:")
	fmt.Printf("  Disk Size: %d bytes (%.2f MB)\n", superblock.DiskSize, float64(superblock.DiskSize)/1024/1024)
	fmt.Printf("  Max Inodes: %d\n", superblock.MaxInodes)
	fmt.Printf("  Number of Blocks: %d\n", superblock.NumBlocks)
	fmt.Printf("  Inode Table Start: %d\n", superblock.InodeTableStart)
	fmt.Printf("  Data Start: %d\n", superblock.DataStart)
	fmt.Println()
	return nil
}

func PrintInodeTable() error {
	disk, err := os.Open(constants.VirtualDisk)
	if err != nil {
		return err
	}
	defer disk.Close()

	fmt.Println("Inode Table:")
	buffer := make([]byte, constants.InodeSize)
	for i := int64(0); i < constants.MaxInodes; i++ {
		offset := constants.InodeTableStart + i*constants.InodeSize
		_, err := disk.ReadAt(buffer, offset)
		if err != nil {
			return err
		}

		inode := DeserializeInode(buffer)
		if inode.Size > 0 { // Mostra apenas inodes ocupados
			filename := string(bytes.Trim(inode.Filename[:], "\x00")) // Remove bytes vazios
			fmt.Printf("  Inode %d -> File: %s | Size: %d bytes | Start Block: %d\n",
				i, filename, inode.Size, inode.StartBlock)
		}
	}
	fmt.Println()
	return nil
}

