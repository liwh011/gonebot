package websocket

import (
	"sync"
)

type responseStore struct {
	seqNum     uint64 // 消息序号
	seqNumLock sync.Mutex
	seq2Chan   sync.Map // 消息序号到存取通道的映射
}

func (store *responseStore) Init() {
	store.seqNum = 0
	store.seqNumLock = sync.Mutex{}
	store.seq2Chan = sync.Map{}
}

func (store *responseStore) getSeqNum() uint64 {
	store.seqNumLock.Lock()
	defer store.seqNumLock.Unlock()
	store.seqNum++
	return store.seqNum
}

func (store *responseStore) store(resp response) {
	seq := resp.Echo
	ch, ok := store.seq2Chan.Load(seq)
	if ok {
		ch.(chan response) <- resp
		close(ch.(chan response))
	}
}

func (store *responseStore) get(seq uint64) chan response {
	ch := make(chan response)
	store.seq2Chan.LoadOrStore(seq, ch)
	return ch
}
