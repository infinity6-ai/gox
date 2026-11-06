package bitmapz

type Bit uint64
type Bitmap uint64

func Contains[M Bitmap, B Bit](m M, b B) bool {
	return uint64(m)&uint64(b) != 0
}
