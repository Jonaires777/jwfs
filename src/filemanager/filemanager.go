package filemanager

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/rand"
	"os"

	"github.com/Jonaires777/src/constants"
)

type Inode struct {
	Filename   [32]byte
	Size       int64
	StartBlock int64
}

type SuperBlock struct {
	DiskSize        int64
	MaxInodes       int64
	NumBlocks       int64
	InodeTableStart int64
	DataStart       int64
}

func InitializeBitmap(disk *os.File) error {
	bitmap := make([]byte, constants.BitmapSize)

	for i := int64(0); i < (constants.DataStart / constants.BlockSize); i++ {
		byteIndex := i / 8
		bitOffset := i % 8
		bitmap[byteIndex] |= (1 << bitOffset)
	}

	_, err := disk.WriteAt(bitmap, constants.BitmapStart)
	return err
}

func InitializeSuperblock(disk *os.File) error {
	superblock := SuperBlock{
		DiskSize:        constants.Disksize,
		MaxInodes:       constants.MaxInodes,
		NumBlocks:       constants.NumBlocks,
		InodeTableStart: constants.InodeTableStart,
		DataStart:       constants.DataStart,
	}

	data := make([]byte, 40)
	binary.LittleEndian.PutUint64(data[0:8], uint64(superblock.DiskSize))
	binary.LittleEndian.PutUint64(data[8:16], uint64(superblock.MaxInodes))
	binary.LittleEndian.PutUint64(data[16:24], uint64(superblock.NumBlocks))
	binary.LittleEndian.PutUint64(data[24:32], uint64(superblock.InodeTableStart))
	binary.LittleEndian.PutUint64(data[32:40], uint64(superblock.DataStart))

	_, err := disk.WriteAt(data, constants.SuperBlockStart)
	return err
}

func InitializeInodeTable(disk *os.File) error {
	emptyInode := make([]byte, constants.InodeSize)
	for i := int64(0); i < constants.MaxInodes; i++ {
		offset := constants.InodeTableStart + i*constants.InodeSize
		_, err := disk.WriteAt(emptyInode, offset)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateBitmap(disk *os.File, blockIndex int64, allocated bool) error {
	if blockIndex >= constants.NumBlocks {
		return errors.New("índice do bloco fora do intervalo")
	}

	byteIndex := blockIndex / 8
	bitOffset := blockIndex % 8

	var singleByte [1]byte
	_, err := disk.ReadAt(singleByte[:], constants.BitmapStart+byteIndex)
	if err != nil {
		return err
	}

	if allocated {
		singleByte[0] |= (1 << bitOffset)
	} else {
		singleByte[0] &^= (1 << bitOffset)
	}

	_, err = disk.WriteAt(singleByte[:], constants.BitmapStart+byteIndex)
	return err
}

func SerializeInode(inode Inode) []byte {
	data := make([]byte, constants.InodeSize)
	copy(data[:32], inode.Filename[:])
	binary.LittleEndian.PutUint64(data[32:40], uint64(inode.Size))
	binary.LittleEndian.PutUint64(data[40:48], uint64(inode.StartBlock))
	return data
}

func DeserializeInode(data []byte) Inode {
	var inode Inode
	copy(inode.Filename[:], data[:32])
	inode.Size = int64(binary.LittleEndian.Uint64(data[32:40]))
	inode.StartBlock = int64(binary.LittleEndian.Uint64(data[40:48]))
	return inode
}

func CreateVirtualDisk(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = file.Truncate(constants.Disksize)
	if err != nil {
		return err
	}

	err = InitializeSuperblock(file)
	if err != nil {
		return err
	}

	err = InitializeBitmap(file)
	if err != nil {
		return err
	}

	err = InitializeInodeTable(file)
	if err != nil {
		return err
	}

	return nil
}

func CheckFileExistence(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func CreateFile(filename string, size int) error {
	disk, err := os.OpenFile(constants.VirtualDisk, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer disk.Close()

	inodes, _, _ := ListFiles()
	for _, inode := range inodes {
		if string(inode.Filename[:bytes.IndexByte(inode.Filename[:], 0)]) == filename {
			return errors.New("arquivo já existe")
		}
	}

	inodeOffset, err := findFreeInode(disk)
	if err != nil {
		return err
	}

	startBlock, err := findFreeBlock(disk)
	if err != nil {
		return err
	}

	if int64(size)*4 > constants.Disksize-(startBlock-constants.DataStart) {
		return errors.New("espaço insuficiente no disco")
	}

	err = UpdateBitmap(disk, (startBlock-constants.DataStart)/constants.BlockSize, true)
	if err != nil {
		return err
	}

	var inode Inode
	copy(inode.Filename[:], []byte(filename))
	inode.Size = int64(size)
	inode.StartBlock = startBlock

	_, err = disk.WriteAt(SerializeInode(inode), inodeOffset)
	if err != nil {
		return err
	}

	disk.Seek(startBlock, 0)
	data := make([]byte, size*4)
	for i := 0; i < size; i++ {
		binary.LittleEndian.PutUint32(data[i*4:], uint32(rand.Intn(100000)))
	}

	_, err = disk.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func ListFiles() ([]Inode, int64, error) {
	var inodes []Inode
	var totalUsed int64

	disk, err := os.Open(constants.VirtualDisk)
	if err != nil {
		return nil, 0, err
	}
	defer disk.Close()

	buffer := make([]byte, constants.InodeSize)
	for i := int64(0); i < constants.MaxInodes; i++ {
		offset := constants.InodeTableStart + i*constants.InodeSize
		_, err := disk.ReadAt(buffer, offset)
		if err != nil {
			break
		}

		inode := DeserializeInode(buffer)
		if inode.Size > 0 {
			inodes = append(inodes, inode)
			totalUsed += inode.Size
		}
	}

	return inodes, totalUsed, nil
}

func RemoveFile(filename string) error {
	disk, err := os.OpenFile(constants.VirtualDisk, os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	inodes, _, _ := ListFiles()
	for _, inode := range inodes {
		if string(inode.Filename[:bytes.IndexByte(inode.Filename[:], 0)]) == filename {
			err = UpdateBitmap(disk, (inode.StartBlock-constants.DataStart)/constants.BlockSize, false)
			if err != nil {
				return err
			}

			var emptyInode Inode
			_, err = disk.WriteAt(SerializeInode(emptyInode), inode.StartBlock)
			if err != nil {
				return err
			}

			return nil
		}
	}
	return errors.New("arquivo não encontrado")
}

func findFreeInode(disk *os.File) (int64, error) {
	buffer := make([]byte, constants.InodeSize)
	for i := int64(0); i < constants.MaxInodes; i++ {
		offset := constants.InodeTableStart + i*constants.InodeSize
		_, err := disk.ReadAt(buffer, offset)
		if err != nil {
			return -1, err
		}

		inode := DeserializeInode(buffer)
		if inode.Size == 0 {
			return offset, nil
		}
	}
	return -1, errors.New("sem espaço para novos arquivos")
}

func findFreeBlock(disk *os.File) (int64, error) {
	bitmap := make([]byte, constants.BitmapSize)
	_, err := disk.ReadAt(bitmap, constants.BitmapStart)
	if err != nil {
		return -1, err
	}

	for i := int64(0); i < constants.BitmapSize; i++ {
		for j := 0; j < 8; j++ {
			if (bitmap[i] & (1 << j)) == 0 {
				blockIndex := i*8 + int64(j)
				if blockIndex >= constants.NumBlocks {
					return -1, errors.New("sem blocos disponíveis")
				}
				return constants.DataStart + blockIndex*constants.BlockSize, nil
			}
		}
	}
	return -1, errors.New("sem blocos disponíveis")
}