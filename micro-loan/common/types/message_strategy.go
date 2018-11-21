package types

type MessageStrategy int64

const (
	MessageStrategyStart        MessageStrategy = 0x1
	RegisterNoOrderAccountAm    MessageStrategy = MessageStrategyStart
	RegisterNoOrderAccountPm    MessageStrategy = MessageStrategyStart << 1
	RegisterOrderNoKtpAccountAm MessageStrategy = MessageStrategyStart << 2
	RegisterOrderNoKtpAccountPm MessageStrategy = MessageStrategyStart << 3
)
