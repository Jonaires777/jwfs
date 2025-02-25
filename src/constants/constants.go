package constants

const (
	Disksize = 1 * 1024 * 1024 * 1024
	VirtualDisk = "virtual_disk.img"
	InodeTableStart = 4096
	InodeSize       = 64
	MaxInodes       = 1024
	DataStart       = 4096 + 64*1024
	BlockSize       = 4096
	NumBlocks       = Disksize / BlockSize
	BitmapSize      = NumBlocks / 8
	BitmapStart     = 4096
	SuperBlockStart = 0
)