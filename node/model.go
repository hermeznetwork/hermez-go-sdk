package node

import "time"

type NodeState struct {
	Node              Node              `json:"node"`
	Network           Network           `json:"network"`
	Metrics           Metrics           `json:"metrics"`
	Rollup            Rollup            `json:"rollup"`
	Auction           Auction           `json:"auction"`
	WithdrawalDelayer WithdrawalDelayer `json:"withdrawalDelayer"`
	RecommendedFee    RecommendedFee    `json:"recommendedFee"`
}

type Node struct {
	ForgeDelay int `json:"forgeDelay"`
	PoolLoad   int `json:"poolLoad"`
}

type (
	CollectedFees struct{}
	LastBatch     struct {
		ItemID                        int           `json:"itemId"`
		BatchNum                      int           `json:"batchNum"`
		EthereumBlockNum              int           `json:"ethereumBlockNum"`
		EthereumBlockHash             string        `json:"ethereumBlockHash"`
		Timestamp                     time.Time     `json:"timestamp"`
		ForgerAddr                    string        `json:"forgerAddr"`
		CollectedFees                 CollectedFees `json:"collectedFees"`
		HistoricTotalCollectedFeesUsd int           `json:"historicTotalCollectedFeesUSD"`
		StateRoot                     string        `json:"stateRoot"`
		NumAccounts                   int           `json:"numAccounts"`
		ExitRoot                      string        `json:"exitRoot"`
		ForgeL1TransactionsNum        int           `json:"forgeL1TransactionsNum"`
		SlotNum                       int           `json:"slotNum"`
		ForgedTransactions            int           `json:"forgedTransactions"`
	}
)

type Coordinator struct {
	ItemID        int    `json:"itemId"`
	BidderAddr    string `json:"bidderAddr"`
	ForgerAddr    string `json:"forgerAddr"`
	EthereumBlock int    `json:"ethereumBlock"`
	URL           string `json:"URL"`
}

type Period struct {
	SlotNum       int    `json:"slotNum"`
	FromBlock     int    `json:"fromBlock"`
	ToBlock       int    `json:"toBlock"`
	FromTimestamp string `json:"fromTimestamp"`
	ToTimestamp   string `json:"toTimestamp"`
}

type NextForgers struct {
	Coordinator Coordinator `json:"coordinator"`
	Period      Period      `json:"period"`
}

type Network struct {
	LastEthereumBlock     int           `json:"lastEthereumBlock"`
	LastSynchedBlock      int           `json:"lastSynchedBlock"`
	LastBatch             LastBatch     `json:"lastBatch"`
	CurrentSlot           int           `json:"currentSlot"`
	NextForgers           []NextForgers `json:"nextForgers"`
	PendingL1Transactions int           `json:"pendingL1Transactions"`
}

type Metrics struct {
	TransactionsPerBatch   float64 `json:"transactionsPerBatch"`
	BatchFrequency         float64 `json:"batchFrequency"`
	TransactionsPerSecond  float64 `json:"transactionsPerSecond"`
	TokenAccounts          int     `json:"tokenAccounts"`
	Wallets                int     `json:"wallets"`
	AvgTransactionFee      float64 `json:"avgTransactionFee"`
	EstimatedTimeToForgeL1 int     `json:"estimatedTimeToForgeL1"`
}

type Buckets struct {
	CeilUsd         string `json:"ceilUSD"`
	BlockStamp      string `json:"blockStamp"`
	Withdrawals     string `json:"withdrawals"`
	RateBlocks      string `json:"rateBlocks"`
	RateWithdrawals string `json:"rateWithdrawals"`
	MaxWithdrawals  string `json:"maxWithdrawals"`
}

type Rollup struct {
	EthereumBlockNum      int       `json:"ethereumBlockNum"`
	FeeAddToken           string    `json:"feeAddToken"`
	ForgeL1L2BatchTimeout int       `json:"forgeL1L2BatchTimeout"`
	WithdrawalDelay       int       `json:"withdrawalDelay"`
	Buckets               []Buckets `json:"buckets"`
	SafeMode              bool      `json:"safeMode"`
}

type Auction struct {
	EthereumBlockNum         int      `json:"ethereumBlockNum"`
	DonationAddress          string   `json:"donationAddress"`
	BootCoordinator          string   `json:"bootCoordinator"`
	BootCoordinatorURL       string   `json:"bootCoordinatorUrl"`
	DefaultSlotSetBid        []string `json:"defaultSlotSetBid"`
	DefaultSlotsetBidsLotNum int      `json:"defaultSlotSetBidSlotNum"`
	ClosedAuctionSlots       int      `json:"closedAuctionSlots"`
	OpenAuctionSlots         int      `json:"openAuctionSlots"`
	AllocationRatio          []int    `json:"allocationRatio"`
	OutBidding               int      `json:"outbidding"`
	SlotDeadline             int      `json:"slotDeadline"`
}

type WithdrawalDelayer struct {
	Ethereumblocknum           int    `json:"ethereumBlockNum"`
	Hermezgovernanceaddress    string `json:"hermezGovernanceAddress"`
	Emergencycounciladdress    string `json:"emergencyCouncilAddress"`
	Withdrawaldelay            int    `json:"withdrawalDelay"`
	Emergencymodestartingblock int    `json:"emergencyModeStartingBlock"`
	Emergencymode              bool   `json:"emergencyMode"`
}

type RecommendedFee struct {
	Existingaccount       float64 `json:"existingAccount"`
	Createaccount         float64 `json:"createAccount"`
	Createaccountinternal float64 `json:"createAccountInternal"`
}
