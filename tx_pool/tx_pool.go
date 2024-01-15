package main

import (
	"bufio"
	"dumbo_fabric/network"
	pb "dumbo_fabric/struct"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"
	
	"github.com/golang/protobuf/proto"

	"gopkg.in/yaml.v2"
)

//var blk_size int = 800

type Tx_pool struct {
	Tx_pool       string `yaml:"Tx_pool"`
	BC_l_m        string `yaml:"BC_l_m"`
	Node_num      int    `yaml:"Node_num"`
	BatchSize     int    `yaml:"BatchSize"`
	BCIPPath      string `yaml:"BCIPPath"`
	TpIPPath      string `yaml:"TpIPPath"`
	Min_BatchSize int    `yaml:"Min_BatchSize"`
	ID            int
	K             int `yaml:"K"`

	clientMsgCH    chan []byte
	bcConn         []net.Conn
	conReady       chan bool
	msgOut         chan []byte
	IsControlSpeed bool `yaml:"IsControlSpeed"`
	TxpoolNetSpeed int  `yaml:"TxpoolNetSpeed"`
	Interval       int  `yaml:"Interval"`
	net            network.Network
}

func NewTx_Pool(id int) *Tx_pool {
	newTx_pool := &Tx_pool{}
	gopath := os.Getenv("GOPATH")
	tx_poolBytes, err := ioutil.ReadFile(gopath + "/src/dumbo_fabric/config/node.yaml")
	if err != nil {
		panic(err)
	}
	err1 := yaml.Unmarshal(tx_poolBytes, newTx_pool)
	if err1 != nil {
		panic(err1)
	}

	newTx_pool.ID = id
	newTx_pool.clientMsgCH = make(chan []byte, 10000)
	newTx_pool.conReady = make(chan bool, 1)
	newTx_pool.bcConn = make([]net.Conn, newTx_pool.K)
	newTx_pool.msgOut = make(chan []byte, 1000)
	newTx_pool.net = network.New(newTx_pool.IsControlSpeed, newTx_pool.TxpoolNetSpeed, false, 0)
	newTx_pool.net.Init()
	return newTx_pool
}

func (tp *Tx_pool) Init() {
	//read ip
	gopath := os.Getenv("GOPATH")
	TpIPPath := fmt.Sprintf("%s%sip.txt", gopath, tp.TpIPPath)
	fi, err := os.Open(TpIPPath)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	br := bufio.NewReader(fi)
	var tpIP string
	a, _, c := br.ReadLine()
	if c == io.EOF {
		panic("missing txpool ip")
	}
	tpIP = string(a)

	fi.Close()

	bcmips := make([]string, tp.K)

	ipmPath := fmt.Sprintf("%s%sipm.txt", gopath, tp.BCIPPath)
	fi, err = os.Open(ipmPath)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	br = bufio.NewReader(fi)
	for j := 0; j < tp.K; j++ {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			panic("missing broadcast ip")
		}
		bcmips[j] = string(a)

	}

	fi.Close()

	go tp.handle_clientMsg()
	go tp.handle_msgOutCH()

	go func() {
		time.Sleep(time.Duration(20) * time.Second)
		for i := 0; i < tp.K; i++ {
			go func(i int) {
				var err error

				tp.bcConn[i], err = network.Dial(bcmips[i])
				fmt.Println("dial:", bcmips[i])
				if err != nil {
					panic(err)
				}
				tp.conReady <- true
			}(i)
		}

	}()

	network.Listen_msg(tpIP, tp.clientMsgCH, true)
}

func (tp *Tx_pool) handle_clientMsg() {
	txCount := 0
	var txBlk [][]byte
	for i := 0; i < tp.K; i++ {
		<-tp.conReady
	}

	var ticker time.Ticker
	firsttime := true
	forcesend := false
	hassent := false
	for {
		tx := <-tp.clientMsgCH
		if firsttime {
			ticker = *time.NewTicker(time.Duration(tp.Interval) * time.Millisecond)
			firsttime = false
		}
		select {
		case <-ticker.C:
			if !hassent {
				forcesend = true
			}
			hassent = false

		default:
		}

		//fmt.Println("get a tx of length:", len(tx))
		/*if len(tx) == 1 {
			countnow := len(txBlk)
			if countnow < tp.Min_BatchSize {
				for i := 0; i < tp.Min_BatchSize-countnow; i++ {
					tx := <-tp.clientMsgCH
					txBlk = append(txBlk, tx[:])
				}
			}

			txs := pb.TXs{
				Txs: txBlk,
			}
			txsBytes, err := proto.Marshal(&txs)
			if err != nil {
				panic(err)
			}
			txCount = 0
			bp := pb.TxPool{
				Payloads: txsBytes,
			}
			bpBytes, err := proto.Marshal(&bp)
			if err != nil {
				panic(err)
			}
			//fmt.Println("generate a txblk of len:", len(bpBytes))
			tp.msgOut <- bpBytes
			if err != nil {
				panic(err)
			}
			txBlk = txBlk[0:0]
		} else {*/
		txBlk = append(txBlk, tx[:])
		txCount++
		if forcesend {
			//send directly
			countnow := len(txBlk)
			if countnow < tp.Min_BatchSize {
				for i := 0; i < tp.Min_BatchSize-countnow; i++ {
					tx := <-tp.clientMsgCH
					txBlk = append(txBlk, tx[:])
				}
			}
			txs := pb.TXs{
				Txs: txBlk,
			}
			txsBytes, err := proto.Marshal(&txs)
			if err != nil {
				panic(err)
			}
			txCount = 0
			bp := pb.TxPool{
				Payloads: txsBytes,
			}
			bpBytes, err := proto.Marshal(&bp)
			if err != nil {
				panic(err)
			}
			fmt.Println("generate a txblk of len:", len(bpBytes)/250)
			tp.msgOut <- bpBytes
			if err != nil {
				panic(err)
			}
			txBlk = txBlk[0:0]
			forcesend = false
		} else {
			if txCount >= tp.BatchSize {
				txs := pb.TXs{
					Txs: txBlk,
				}
				txsBytes, err := proto.Marshal(&txs)
				if err != nil {
					panic(err)
				}
				txCount = 0
				bp := pb.TxPool{
					Payloads: txsBytes,
				}
				bpBytes, err := proto.Marshal(&bp)
				if err != nil {
					panic(err)
				}
				fmt.Println("generate a txblk of len:", len(bpBytes)/250)
				tp.msgOut <- bpBytes
				if err != nil {
					panic(err)
				}
				txBlk = txBlk[0:0]
				hassent = true
			}
		}

	}
}

func (tp *Tx_pool) handle_msgOutCH() {
	i := 0
	for {
		i = i % tp.K
		msg := <-tp.msgOut
		tp.net.Send(tp.bcConn[i], msg)

		//fmt.Println("done send a txblk of len:", len(msg))
		i++
	}
}

func main() {
	var idf = flag.Int("id", 0, "-id")
	flag.Parse()
	id := *idf
	newTxPool := NewTx_Pool(id)
	newTxPool.Init()

}

