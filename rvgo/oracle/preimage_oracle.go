package oracle

type PreImageOracle interface {
	ReadPreImagePart(key [32]byte, offset uint64) (dat [32]byte, datlen uint8, err error)
}

type PreImageReaderFn func(key [32]byte, offset uint64) (dat [32]byte, datlen uint8, err error)

func (fn PreImageReaderFn) ReadPreImagePart(key [32]byte, offset uint64) (dat [32]byte, datlen uint8, err error) {
	return fn(key, offset)
}
