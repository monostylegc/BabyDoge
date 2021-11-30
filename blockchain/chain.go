package blockchain

import (
	"sync"

	"github.com/monostylegc/BabyDoge/db"
	"github.com/monostylegc/BabyDoge/utils"
)

type blockchain struct {
	NewestHash string `json:"newestHash"`
	Height     int    `json:"height"`
}

var b *blockchain

var once sync.Once

func (b *blockchain) restore(data []byte) {
	utils.FromBytes(b, data)
}

func (b *blockchain) persist() {
	db.SaveCheckpoint(utils.ToBytes(b))
}

func (b *blockchain) AddBlock(data string) {
	block := createBlock(data, b.NewestHash, b.Height+1)
	b.NewestHash = block.Hash
	b.Height = block.Height
	b.persist()
}

func (b *blockchain) Blocks() []*Block {
	hashCursor := b.NewestHash
	var blocks []*Block

	for {
		block, _ := FindBlock(hashCursor)
		blocks = append(blocks, block)
		if block.PrevHash != "" {
			hashCursor = block.PrevHash
		} else {
			break
		}
	}
	return blocks
}

func Blockchain() *blockchain {
	if b == nil {
		once.Do(func() {
			b = &blockchain{"", 0}
			//우선 checkpoint를 찾아본다음 blockchain을 db에서 불러온다 db.Blockchain은 data or nil을 return
			checkpoint := db.Checkpoint()

			if checkpoint == nil {
				//아무것도 없으면 생성
				b.AddBlock("Genesis")

			} else {
				//restore from byte
				b.restore(checkpoint)
			}
		})
	}
	return b
}
