package blockchain

import (
	"sync"

	"github.com/monostylegc/BabyDoge/db"
	"github.com/monostylegc/BabyDoge/utils"
)

const (
	defaultDifficulty  int = 2
	difficultyInterval int = 5
	blockInterval      int = 2
	allowedRange       int = 2
)

type blockchain struct {
	NewestHash        string `json:"newestHash"`
	Height            int    `json:"height"`
	CurrentDifficulty int    `json:"currentDifficulty"`
}

var b *blockchain

var once sync.Once

//byte에서 blockchain restore
func (b *blockchain) restore(data []byte) {
	utils.FromBytes(b, data)
}

//db에 blockchain 저장
func persistBlockchain(b *blockchain) {
	db.SaveCheckpoint(utils.ToBytes(b))
}

//difficulty를 정함
func getDifficulty(b *blockchain) int {
	if b.Height == 0 {
		return defaultDifficulty
	} else if b.Height%difficultyInterval == 0 {
		return recalculateDifficulty(b)
	} else {
		return b.CurrentDifficulty
	}
}

//난이도를 재설정 너무빠르거나 느리게 채굴되지 않도록
func recalculateDifficulty(b *blockchain) int {
	//모든 블록을 불러옴
	allBlocks := Blocks(b)
	//가장 새로운 블럭
	newestBlock := allBlocks[0]
	//마지막 난이도가 계산된 블럭 interval - 1
	lastRecalculatedBlock := allBlocks[difficultyInterval-1]
	//interval 만큼(현재는 5개) 블럭이 실제로 생성된 시간
	actualTime := (newestBlock.Timestamp / 60) - (lastRecalculatedBlock.Timestamp / 60)
	//예상시간 10분
	expectedTime := difficultyInterval * blockInterval

	//실제시간이 짧으면 난이도를 올린다.
	if actualTime <= (expectedTime - allowedRange) {
		return b.CurrentDifficulty + 1
		//실제시간이 길면 난이도를 줄인다.
	} else if actualTime >= (expectedTime + allowedRange) {
		return b.CurrentDifficulty - 1
	}
	//난이도를 return
	return b.CurrentDifficulty
}

//Block 추가
func (b *blockchain) AddBlock() {
	block := createBlock(b.NewestHash, getDifficulty(b), b.Height+1)
	b.NewestHash = block.Hash
	b.CurrentDifficulty = block.Difficulty
	b.Height = block.Height
	persistBlockchain(b)
}

//block들을 불러온다(전부)
func Blocks(b *blockchain) []*Block {
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

//address의 잔고를 확인해준다.
func BalanceByAddress(address string, b *blockchain) int {
	txOuts := UTxOutsByAddress(address, b)
	var amount int
	for _, txOut := range txOuts {
		amount += txOut.Amount
	}
	return amount
}

func UTxOutsByAddress(address string, b *blockchain) []*UTxOut {
	var uTxOuts []*UTxOut
	creatorTxs := make(map[string]bool)

	for _, block := range Blocks(b) {
		for _, tx := range block.Transactions {
			for _, input := range tx.TxIns {
				if input.Owner == address {
					creatorTxs[input.TxID] = true
				}
			}
			for index, output := range tx.TxOuts {
				if output.Owner == address {
					if _, ok := creatorTxs[tx.ID]; !ok {
						uTxOut := &UTxOut{tx.ID, index, output.Amount}
						if !isOnMempool(uTxOut) {
							uTxOuts = append(uTxOuts, uTxOut)
						}
					}
				}
			}
		}
	}
	return uTxOuts
}

//initial함수
func Blockchain() *blockchain {

	once.Do(func() {
		b = &blockchain{
			Height: 0,
		}
		//우선 checkpoint를 찾아본다음 blockchain을 db에서 불러온다 db.Blockchain은 data or nil을 return
		checkpoint := db.Checkpoint()

		if checkpoint == nil {
			//아무것도 없으면 생성
			b.AddBlock()

		} else {
			//restore from byte
			b.restore(checkpoint)
		}
	})
	return b
}
