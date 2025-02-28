package filemanager

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"os"
	"sort"
	"syscall"
	"time"

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
		DiskSize:        constants.DiskSize,
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

	err = file.Truncate(constants.DiskSize)
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

	if int64(size)*4 > constants.DiskSize-(startBlock-constants.DataStart) {
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

	data := make([]byte, size*4)
	for i := 0; i < size; i++ {
		binary.LittleEndian.PutUint32(data[i*4:], uint32(rand.Intn(100000)))
	}

	_, err = disk.WriteAt(data, startBlock)
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
			if string(inode.Filename[:bytes.IndexByte(inode.Filename[:], 0)]) == filename {
				err = UpdateBitmap(disk, (inode.StartBlock-constants.DataStart)/constants.BlockSize, false)
				if err != nil {
					return err
				}

				emptyInode := Inode{}
				_, err = disk.WriteAt(SerializeInode(emptyInode), constants.InodeTableStart+i*constants.InodeSize)
				if err != nil {
					return err
				}

				return nil
			}
		}
	}

	return errors.New("arquivo não encontrado")
}

func ReadFile(filename string, startIdx, endIdx int64) ([]int32, error) {
	disk, err := os.Open(constants.VirtualDisk)
	if err != nil {
		return nil, err
	}
	defer disk.Close()

	if startIdx < 0 || endIdx < 0 || startIdx > endIdx {
		return nil, errors.New("índices inválidos")
	}

	inodes, _, _ := ListFiles()
	for _, inode := range inodes {
		if string(inode.Filename[:bytes.IndexByte(inode.Filename[:], 0)]) == filename {

			if int64(endIdx) > inode.Size {
				return nil, errors.New("índice final maior que o tamanho do arquivo")
			}

			var numbers []int32
			for i := startIdx; i < endIdx; i++ {
				offset := inode.StartBlock + i*4
				data := make([]byte, 4)
				_, err := disk.ReadAt(data, offset)
				if err != nil {
					return nil, err
				}
				numbers = append(numbers, int32(binary.LittleEndian.Uint32(data)))
			}
			return numbers, nil
		}
	}
	return nil, errors.New("arquivo não encontrado")
}

func allocateHugePage() ([]byte, error) {
	// Abrir um arquivo especial para Huge Pages (necessário para mapeamento)
	file, err := os.OpenFile("/dev/zero", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Mapear 2MB de memória usando Huge Pages
	data, err := syscall.Mmap(int(file.Fd()), 0, constants.HugePageSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_PRIVATE|syscall.MAP_HUGETLB)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func createPagingFile() error {
	file, err := os.Create(constants.PagingFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Garantir que o arquivo tenha um tamanho inicial de 2MB
	err = file.Truncate(constants.HugePageSize)
	if err != nil {
		return err
	}

	return nil
}

func OrderFile(filename string) (time.Duration, error) {
	hugePage, err := allocateHugePage()
	if err != nil {
		return 0, err
	}
	defer syscall.Munmap(hugePage)

	err = createPagingFile()
	if err != nil {
		return 0, err
	}

	pagingFile, err := os.OpenFile(constants.PagingFile, os.O_RDWR, 0666)
	if err != nil {
		return 0, err
	}
	defer pagingFile.Close()

	disk, err := os.OpenFile(constants.VirtualDisk, os.O_RDWR, 0666)
	if err != nil {
		return 0, err
	}
	defer disk.Close()

	buffer := make([]byte, constants.InodeSize)
	var inode *Inode

	// Buscar o inode do arquivo
	for i := int64(0); i < constants.MaxInodes; i++ {
		offset := constants.InodeTableStart + i*constants.InodeSize
		n, err := disk.ReadAt(buffer, offset)
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n == 0 {
			break
		}

		in := DeserializeInode(buffer)
		if in.Size > 0 && string(in.Filename[:bytes.IndexByte(in.Filename[:], 0)]) == filename {
			inode = &in
			break
		}
	}
	if inode == nil {
		return 0, errors.New("arquivo não encontrado")
	}

	startTime := time.Now()

	// Ordenação Externa (Lendo blocos de 2MB por vez)
	for i := int64(0); i < inode.Size; i += constants.HugePageSize {
		offset := inode.StartBlock + i

		bytesToRead := int64(constants.HugePageSize)
		if inode.Size-i < bytesToRead {
			bytesToRead = inode.Size - i
		}

		n, err := disk.ReadAt(hugePage[:bytesToRead], offset)
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n == 0 {
			break
		}

		// Ordenar os números dentro da Huge Page
		numbers := make([]int, n/4)
		for j := 0; j < n; j += 4 {
			numbers[j/4] = int(binary.LittleEndian.Uint32(hugePage[j:]))
		}

		sort.Ints(numbers)

		pagingFile.Seek(0, 0)
		for _, num := range numbers {
			err := binary.Write(pagingFile, binary.LittleEndian, uint32(num))
			if err != nil {
				return 0, err
			}
		}
	}

	// Escrever de volta no arquivo original
	pagingFile.Seek(0, 0)
	for i := int64(0); i < inode.Size; i += constants.HugePageSize {
		offset := inode.StartBlock + i

		n, err := pagingFile.Read(hugePage)
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n == 0 {
			break
		}

		disk.WriteAt(hugePage[:n], offset)
	}

	elapsedTime := time.Since(startTime)
	return elapsedTime, nil
}

func ConcatFiles(filename1, filename2, newFilename string) error {
	disk, err := os.OpenFile(constants.VirtualDisk, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer disk.Close()

	buffer := make([]byte, constants.InodeSize)
	var inode1, inode2 Inode
	for i := int64(0); i < constants.MaxInodes; i++ {
		offset := constants.InodeTableStart + i*constants.InodeSize
		_, err := disk.ReadAt(buffer, offset)
		if err != nil {
			break
		}

		inode := DeserializeInode(buffer)
		if inode.Size > 0 {
			if string(inode.Filename[:bytes.IndexByte(inode.Filename[:], 0)]) == filename1 {
				inode1 = inode
				err = UpdateBitmap(disk, (inode1.StartBlock-constants.DataStart)/constants.BlockSize, false)

				if err != nil {
					return err
				}

				err = RemoveFile(filename1)
				if err != nil {
					return err
				}

			} else if string(inode.Filename[:bytes.IndexByte(inode.Filename[:], 0)]) == filename2 {
				inode2 = inode

				err = UpdateBitmap(disk, (inode2.StartBlock-constants.DataStart)/constants.BlockSize, false)
				if err != nil {
					return err
				}

				err = RemoveFile(filename2)
				if err != nil {
					return err
				}
			}
		}
	}

	if inode1.Size == 0 || inode2.Size == 0 {
		return errors.New("arquivo não encontrado")
	}

	newSize := inode1.Size + inode2.Size
	if newSize*4 > constants.DiskSize-(inode1.StartBlock-constants.DataStart) {
		return errors.New("espaço insuficiente no disco")
	}

	newInodeOffset, err := findFreeInode(disk)
	if err != nil {
		return err
	}

	newStartBlock, err := findFreeBlock(disk)
	if err != nil {
		return err
	}

	err = UpdateBitmap(disk, (newStartBlock-constants.DataStart)/constants.BlockSize, true)
	if err != nil {
		return err
	}

	var newInode Inode
	copy(newInode.Filename[:], []byte(newFilename))
	newInode.Size = newSize
	newInode.StartBlock = newStartBlock

	_, err = disk.WriteAt(SerializeInode(newInode), newInodeOffset)
	if err != nil {
		return err
	}

	data1 := make([]byte, inode1.Size*4)
	for i := int64(0); i < inode1.Size; i++ {
		offset := inode1.StartBlock + i*4
		data := make([]byte, 4)
		_, err := disk.ReadAt(data, offset)
		if err != nil {
			return err
		}
		copy(data1[i*4:], data)
	}

	data2 := make([]byte, inode2.Size*4)
	for i := int64(0); i < inode2.Size; i++ {
		offset := inode2.StartBlock + i*4
		data := make([]byte, 4)
		_, err := disk.ReadAt(data, offset)
		if err != nil {
			return err
		}
		copy(data2[i*4:], data)
	}

	newData := append(data1, data2...)
	_, err = disk.WriteAt(newData, newStartBlock)
	if err != nil {
		return err
	}

	return nil
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
