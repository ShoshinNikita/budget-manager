package bufpool

func (b *Buffer) reset(pos int) {
	b.ResetBytes(b.buf[:pos])
}

func (b *Buffer) ResetBytes(buf []byte) {
	b.buf = buf
	b.off = 0
	b.lastRead = opInvalid
}
