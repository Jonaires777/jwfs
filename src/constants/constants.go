package constants

const (
	DiskSize        = 1 * 1024 * 1024 * 1024 // 1GB
	VirtualDisk     = "virtual_disk.img"
	InodeSize       = 64
	MaxInodes       = 1024
	BlockSize       = 4096
	NumBlocks       = DiskSize / BlockSize
	BitmapSize      = NumBlocks / 8
	SuperBlockStart = 0
	BitmapStart     = BlockSize
	InodeTableStart = BitmapStart + BitmapSize
	InodeTableSize  = MaxInodes * InodeSize
	DataStart       = InodeTableStart + InodeTableSize
	MaxFilenameLen  = 32
	HugePageSize    = 2 * 1024 * 1024
	PagingFile      = "paging_file.bin"
)
