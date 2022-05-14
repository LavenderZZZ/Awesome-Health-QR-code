// Package blockchain provides an implementation of the consensus.BlockChain interface.
package blockchain

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/relab/hotstuff/consensus"
	"log"
	"sync"
	"time"
)

// blockChain stores a limited amount of blocks in a map.
// blocks are evicted in LRU order.
type blockChain struct {
	mods          *consensus.Modules
	mut           sync.Mutex
	pruneHeight   consensus.View
	blocks        map[consensus.Hash]*consensus.Block
	blockAtHeight map[consensus.View]*consensus.Block
	pendingFetch  map[consensus.Hash]context.CancelFunc // allows a pending fetch operation to be cancelled
}

// InitConsensusModule gives the module a reference to the Modules object.
// It also allows the module to set module options using the OptionsBuilder.
func (chain *blockChain) InitConsensusModule(mods *consensus.Modules, _ *consensus.OptionsBuilder) {
	chain.mods = mods

}

// New creates a new blockChain with a maximum size.
// Blocks are dropped in least recently used order.
func New() consensus.BlockChain {
	bc := &blockChain{
		blocks:        make(map[consensus.Hash]*consensus.Block),
		blockAtHeight: make(map[consensus.View]*consensus.Block),
		pendingFetch:  make(map[consensus.Hash]context.CancelFunc),
	}
	bc.Store(consensus.GetGenesis())
	return bc
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}
func DeserializeBlock(d []byte) *consensus.Block {
	var block consensus.Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		//fmt.Println("error")
	}
	return &block
}

func (i *BlockchainIterator) Next() *consensus.Block {
	var block *consensus.Block
	db, err := bolt.Open(dbFile, 0600, nil)
	i.db = db
	defer db.Close()
	err = i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	hash1 := block.Parent()
	hashp := hash1[:]
	i.currentHash = hashp

	return block
}

// Store stores a block in the blockchain
func (chain *blockChain) Store(block *consensus.Block) {

	chain.mut.Lock()
	defer chain.mut.Unlock()

	chain.blocks[block.Hash()] = block
	chain.blockAtHeight[block.View()] = block
	fmt.Print(chain.blocks)

	// cancel any pending fetch operations
	if cancel, ok := chain.pendingFetch[block.Hash()]; ok {
		cancel()
	}

	fmt.Printf(" the cmd of the storing block is %s\n", block.Command())
	chain.addBlockchain(block)

	//bc := NewBlockchain()
	//bci := bc.Iterator()
	//for {
	//	block := bci.Next()
	//	chain.mods.Logger().Infof("Parent Hash is %s\n Block hash is %s\n  block cmd is %s\n", block.Parent().String(), block.Hash().String(), block.Command())
	//
	//	if len(block.Parent()) == 0 {
	//		break
	//	}
	//
	//}
}

const dbFile = "blockchain_demo4.db"
const blocksBucket = "blocks"

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

func NewGenesisBlock() *consensus.Block {
	return consensus.NewBlock(consensus.Hash{}, consensus.QuorumCert{}, "genesis block", 0, 0)
}

func NewBlockchain() *Blockchain {
	var tip []byte
	//fmt.Printf("before open database\n")
	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库
	if err != nil {
		log.Panic(err)
		fmt.Printf("in open\n")
	}
	defer db.Close()
	//fmt.Printf("open database\n")

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		//fmt.Printf("get bucket\n")
		if b == nil { //新建存储bucket
			fmt.Println("No existing blockchain found.Creating a new one...")
			//fmt.Printf("before genesis\n")
			genesis := NewGenesisBlock()
			//fmt.Printf("after genesis\n")
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				fmt.Printf("in creat bucket\n")
				log.Panic(err)
			}

			//consensus.Modules.Logger().Infof()
			hash1 := genesis.Hash()
			var hashin = []byte{}
			hashin = hash1[:]
			g := genesis.Serialize()

			//var k=[]byte{12,32}
			//var v=[]byte{12,36}
			//err = b.Put(hashin, v)
			err = b.Put(hashin, g) //hash作为键值，指向区块链信息

			if err != nil {
				fmt.Printf("in put genesis\n")
				log.Panic(err)

			}
			//fmt.Printf("creat ok?\n")

			err = b.Put([]byte("l"), hashin)
			if err != nil {
				log.Panic(err)
				//fmt.Printf("in put2\n")
			}

			tip = hashin

		} else {
			tip = b.Get([]byte("l")) //得到genesis.Hash
		}

		return nil
	})

	bc := Blockchain{tip, db}

	return &bc
}

func (chain *blockChain) addBlockchain(block *consensus.Block) {

	var bc *Blockchain = nil

	bc = NewBlockchain()

	bc.db.Close()
	//fmt.Printf("get bc\n")
	bc.AddBlock(block)
	bc.db.Close()
	//fmt.Printf("adding block info: Parent Hash is %s\n Block hash is %s\n  block cmd is %s\n",block.Parent().String(),block.Hash().String(),block.Command())
}

func (bc *Blockchain) AddBlock(badd *consensus.Block) {
	fmt.Printf("adding block to database is :Parent Hash is %s\n Block hash is %s\n  block cmd is %s\n", badd.Parent().String(), badd.Hash().String(), badd.Command())
	var lastHash []byte
	//fmt.Printf("before add view\n")
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	defer db.Close()
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l")) //上一个hash
		_ = lastHash
		return nil
	})

	if err != nil {
		log.Panic(err)
		//fmt.Printf("in view\n")
	}
	//fmt.Printf("after add view\n")

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		//fmt.Printf("get bucket\n")
		bhash1 := badd.Hash()
		bhashin := bhash1[:]
		err := b.Put(bhashin, badd.Serialize())
		if err != nil {
			log.Panic(err)
			//fmt.Printf("in add1")
		}

		err = b.Put([]byte("l"), bhashin)
		if err != nil {
			log.Panic(err)
			//fmt.Printf("in add2")
		}

		bc.tip = bhashin

		return nil
	})
	if err != nil {
		log.Panic(err)
		//fmt.Printf("in add update")
	}
}

// Get retrieves a block given its hash. It will only try the local cache.
func (chain *blockChain) LocalGet(hash consensus.Hash) (*consensus.Block, bool) {
	chain.mut.Lock()
	defer chain.mut.Unlock()

	block, ok := chain.blocks[hash]
	if !ok {
		return nil, false
	}

	return block, true
}

// Get retrieves a block given its hash. Get will try to find the block locally.
// If it is not available locally, it will try to fetch the block.
func (chain *blockChain) Get(hash consensus.Hash) (block *consensus.Block, ok bool) {
	// need to declare vars early, or else we won't be able to use goto
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	chain.mut.Lock()
	block, ok = chain.blocks[hash]
	if ok {
		goto done
	}

	ctx, cancel = context.WithCancel(chain.mods.Synchronizer().ViewContext())
	chain.pendingFetch[hash] = cancel

	chain.mut.Unlock()
	chain.mods.Logger().Debugf("Attempting to fetch block: %.8s", hash)
	block, ok = chain.mods.Configuration().Fetch(ctx, hash)
	chain.mut.Lock()

	delete(chain.pendingFetch, hash)
	if !ok {
		// check again in case the block arrived while we we fetching
		block, ok = chain.blocks[hash]
		goto done
	}

	chain.mods.Logger().Debugf("Successfully fetched block: %.8s", hash)

	chain.blocks[hash] = block
	chain.blockAtHeight[block.View()] = block

done:
	defer chain.mut.Unlock()

	if !ok {
		return nil, false
	}

	return block, true
}

// Extends checks if the given block extends the branch of the target block.
func (chain *blockChain) Extends(block, target *consensus.Block) bool {
	current := block
	ok := true
	for ok && current.View() > target.View() {
		current, ok = chain.Get(current.Parent())
	}
	return ok && current.Hash() == target.Hash()
}

func (chain *blockChain) PruneToHeight(height consensus.View) (forkedBlocks []*consensus.Block) {
	chain.mut.Lock()
	defer chain.mut.Unlock()

	committedHeight := chain.mods.Consensus().CommittedBlock().View()
	committedViews := make(map[consensus.View]bool)
	committedViews[committedHeight] = true
	for h := committedHeight; h >= chain.pruneHeight; {
		block, ok := chain.blockAtHeight[h]
		if !ok {
			break
		}
		parent, ok := chain.blocks[block.Parent()]
		if !ok || parent.View() < chain.pruneHeight {
			break
		}
		h = parent.View()
		committedViews[h] = true
	}

	for h := height; h > chain.pruneHeight; h-- {
		if !committedViews[h] {
			block, ok := chain.blockAtHeight[h]
			if ok {
				chain.mods.Logger().Debugf("PruneToHeight: found forked block: %v", block)
				forkedBlocks = append(forkedBlocks, block)
			}
		}
		delete(chain.blockAtHeight, h)
	}
	chain.pruneHeight = height
	return forkedBlocks
}

var _ consensus.BlockChain = (*blockChain)(nil)
