package blockchain

import (
	"errors"
	"strings"
	"time"

	"github.com/monostylegc/BabyDoge/db"
	"github.com/monostylegc/BabyDoge/utils"
)

type Block struct {
	Data       string `json:"data"`
	Hash       string `json:"hash"`
	PrevHash   string `json:"prevHash,omitempty"`
	Height     int    `json:"height"`
	Difficulty int    `json:"difficulty"`
	Nonce      int    `json:"nonce"`
	Timestamp  int    `json:"timestamp"`
}

func (b *Block) persist() {
	db.SaveBlock(b.Hash, utils.ToBytes(b))
}

var ErrNotFound = errors.New("Block not found")

func (b *Block) restore(data []byte) {
	utils.FromBytes(b, data)
}

func FindBlock(hash string) (*Block, error) {
	blockByte := db.Block(hash)
	if blockByte == nil {
		return nil, ErrNotFound
	}
	block := &Block{}
	block.restore(blockByte)
	return block, nil
}

func (b *Block) mine() {
	target := strings.Repeat("0", b.Difficulty)

	for {
		//for loop를 시작할때마다 timestamp를 기록한다.
		b.Timestamp = int(time.Now().Unix())

		hash := utils.Hash(b)

		if strings.HasPrefix(hash, target) {
			b.Hash = hash
			return
		} else {
			b.Nonce++
		}
	}
}

func createBlock(data string, prevHash string, height int) *Block {
	block := &Block{
		Data:       data,
		Hash:       "",
		PrevHash:   prevHash,
		Height:     height,
		Difficulty: Blockchain().difficulty(),
		Nonce:      0,
	}
	block.mine()
	block.persist()
	return block
}
