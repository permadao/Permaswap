package router

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/permadao/permaswap/router/schema"
	"github.com/everVision/everpay-kits/utils"
	"gopkg.in/h2non/gentleman.v2"
)

type NFT struct {
	ContractAddr string `json:"contractAddr"`
	TokenId      string `json:"tokenId"`
	Collection   string `json:"collection"`
	Name         string `json:"name"`
	Owner        string `json:"owner"`
}

func (n *NFT) ID() string {
	return fmt.Sprintf("%s/%s", n.ContractAddr, n.TokenId)
}

type NFTOwnerChangeMsg struct {
	NFTID string
	From  string
	To    string
}

type NFTInfo struct {
	ApiURL string

	NFTs               map[string]*NFT
	NFTToOwner         map[string]string
	OwnerToNFTs        map[string][]string
	Whitelist          []string
	NFTOwnerChangeChan chan *NFTOwnerChangeMsg //notify the change of nft owner

	lock sync.RWMutex
}

func NewNFTInfo(apiURL string, whitelist []string, c chan *NFTOwnerChangeMsg) *NFTInfo {
	n := NFTInfo{
		ApiURL:             apiURL,
		NFTs:               make(map[string]*NFT),
		NFTToOwner:         make(map[string]string),
		OwnerToNFTs:        make(map[string][]string),
		Whitelist:          whitelist,
		NFTOwnerChangeChan: c,
	}

	// init nft info
	for {
		err := n.updateNFTInfo()
		if err == nil {
			break
		}
		log.Error("Failed to get nft info data", "err", err)
		time.Sleep(5 * time.Second)
	}

	return &n
}

func (n *NFTInfo) Run() {
	go func() {
		for {
			err := n.updateNFTInfo()
			if err != nil {
				log.Error("Failed to get nft info data", "err", err)
				time.Sleep(5 * time.Second)
			}

			time.Sleep(5 * time.Minute)
		}
	}()
}

func (n *NFTInfo) GetOwners() map[string]int {
	n.lock.RLock()
	defer n.lock.RUnlock()

	owners := make(map[string]int)
	for o, nfts := range n.OwnerToNFTs {
		if len(nfts) > 0 {
			owners[o] = len(nfts)
		}
	}
	return owners
}

func (n *NFTInfo) GetNFTInfo() (map[string]string, map[string][]string, []string) {
	n.lock.RLock()
	defer n.lock.RUnlock()

	nftToOwner := make(map[string]string)
	for nft, owner := range n.NFTToOwner {
		nftToOwner[nft] = owner
	}

	ownerToNFTs := make(map[string][]string)
	for o, nfts := range n.OwnerToNFTs {
		if len(nfts) > 0 {
			nftsCopy := make([]string, len(nfts))
			copy(nftsCopy, nfts)
			ownerToNFTs[o] = nftsCopy
		}
	}
	return nftToOwner, ownerToNFTs, n.Whitelist
}

func (n *NFTInfo) isOwner(account string) bool {
	n.lock.RLock()
	defer n.lock.RUnlock()

	nfts, ok := n.OwnerToNFTs[account]
	if !ok {
		return false
	}
	if len(nfts) == 0 {
		return false
	}
	return true
}

func (n *NFTInfo) Passed(account string) bool {
	n.lock.RLock()
	defer n.lock.RUnlock()

	if len(n.Whitelist) > 0 {
		for _, addr := range n.Whitelist {
			if addr == account {
				return true
			}
		}
	}

	nfts, ok := n.OwnerToNFTs[account]
	if !ok {
		return false
	}
	if len(nfts) == 0 {
		return false
	}
	return true
}

func (n *NFTInfo) updateNFTInfo() error {
	httpCli := gentleman.New().URL(n.ApiURL)
	req := httpCli.Request()

	req.Path("/info")
	res, err := req.Send()
	if err != nil {
		return err
	}
	defer res.Close()
	info := []schema.ResNFT{}
	err = json.Unmarshal(res.Bytes(), &info)
	if err != nil {
		return err
	}

	req = httpCli.Request()
	req.Path("/ar/info")
	res, err = req.Send()
	if err != nil {
		return err
	}
	defer res.Close()
	arNFTInfo := []schema.ArNFT{}
	err = json.Unmarshal(res.Bytes(), &arNFTInfo)
	if err != nil {
		return err
	}

	nfts := make(map[string]*NFT)
	owners := make(map[string]int)
	nftToOwner := make(map[string]string)
	ownerToNFTs := make(map[string][]string)

	for _, i := range info {

		//in case nft api data is invalid
		if i.ContractAddr == "" || i.Name == "" || i.CollectionName == "" || i.Owner == "" {
			log.Error("Invalid nft data.", "nft", i)
			// todo
			continue
			//return WsErrInvalidNFTData
		}

		_, accid, err := utils.IDCheck(i.Owner)
		if err != nil {
			log.Error("Invalid nft owner.", "n.Owner", i.Owner)
			return WsErrInvalidNFTData
		}
		nft := NFT{i.ContractAddr, i.TokenId, i.CollectionName, i.Name, i.Owner}
		nfts[nft.ID()] = &nft
		owners[accid] += 1

		nftToOwner[nft.ID()] = accid
		ownerToNFTs[accid] = append(ownerToNFTs[accid], nft.ID())
	}

	for _, i := range arNFTInfo {

		//in case nft api data is invalid
		if i.ContractAddr == "" || i.Name == "" || i.CollectionName == "" || i.Owner == "" {
			log.Error("Invalid ar nft data.", "nft", i)
			// todo
			continue
			//return WsErrInvalidNFTData
		}

		_, accid, err := utils.IDCheck(i.Owner)
		if err != nil {
			log.Error("Invalid nft owner.", "n.Owner", i.Owner)
			return WsErrInvalidNFTData
		}
		nft := NFT{i.ContractAddr, i.TokenId, i.CollectionName, i.Name, i.Owner}
		nfts[nft.ID()] = &nft
		owners[accid] += 1

		nftToOwner[nft.ID()] = accid
		ownerToNFTs[accid] = append(ownerToNFTs[accid], nft.ID())
	}

	if n.NFTOwnerChangeChan != nil {
		previousNFTToOwner, _, _ := n.GetNFTInfo()
		for id, previousOwner := range previousNFTToOwner {
			if owner := nftToOwner[id]; owner != previousOwner {
				n.NFTOwnerChangeChan <- &NFTOwnerChangeMsg{
					NFTID: id,
					From:  previousOwner,
					To:    owner,
				}
			}

		}
	}

	n.lock.Lock()
	defer n.lock.Unlock()

	n.NFTs = nfts
	n.NFTToOwner = nftToOwner
	n.OwnerToNFTs = ownerToNFTs
	return nil
}

func (r *Router) CheckNFTOrNot() bool {
	if !SetNFTWhiteList(r.chainID) {
		return false
	}
	if r.NFTInfo == nil || len(r.NFTInfo.GetOwners()) == 0 {
		return false
	}
	return true
}

func (r *Router) nftOwnerChangeProc(msg *NFTOwnerChangeMsg) {
	if i, ok := r.lpAddrToID[msg.From]; ok {
		log.Info("close ws con because this address is not a nft owner", "address", msg.From, "ws session id", i)
		r.lpHub.CloseSession(i)
	}
}

func (n *NFTInfo) SetWhitelist(whitelist []string) {
	n.lock.Lock()
	defer n.lock.Unlock()

	n.Whitelist = whitelist

}
