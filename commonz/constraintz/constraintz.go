package constraintz

type SInts interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type UInts interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Ints interface {
	SInts | UInts
}

type Floats interface {
	~float32 | ~float64
}

type Numbers interface {
	Ints | Floats
}

type Data interface {
	string | []byte
}

type Basic interface {
	string | Numbers
}

type Void any
