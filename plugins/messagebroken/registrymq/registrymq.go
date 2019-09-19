package registrymq

// Register-Broken
type RegisterMQ struct {
	node int      // 节点id
	reg  Registry //
	mq   MQBroken //
}

type Registry interface {
	Register(uid int64, node int)
	DeRegister(uid int64, node int)
}

type MQBroken struct {
}
