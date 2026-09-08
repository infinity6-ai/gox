package storezoption

type TransactionOption interface {
	nop()
}

type TransactionOptionMaxAttempts struct {
	MaxAttempts int
}

func (t *TransactionOptionMaxAttempts) nop() {
	panic("nop")
}
