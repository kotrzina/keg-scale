package scale

import (
	"fmt"
	"time"

	"github.com/hako/durafmt"
	"github.com/kotrzina/keg-scale/pkg/utils"
)

type WarehouseItem struct {
	Keg    int `json:"keg"`
	Amount int `json:"amount"`
}

type PubOutput struct {
	IsOpen   bool   `json:"is_open"`
	OpenedAt string `json:"opened_at"`
	ClosedAt string `json:"closed_at"`
}

type BtDevice struct {
	Name            string `json:"name"`
	IdentityAddress string `json:"identity_address"`
	RSSI            int    `json:"rssi"`
	Known           bool   `json:"known"`
	LastSeen        string `json:"last_seen"`
}

type FullOutput struct {
	IsOk               bool            `json:"is_ok"`
	BeersLeft          int             `json:"beers_left"`
	BeersTotal         int             `json:"beers_total"`
	LastWeight         float64         `json:"last_weight"`
	LastWeightFormated string          `json:"last_weight_formated"`
	LastAt             string          `json:"last_at"`
	LastAtDuration     string          `json:"last_at_duration"`
	Rssi               float64         `json:"rssi"`
	LastUpdate         string          `json:"last_update"`
	LastUpdateDuration string          `json:"last_update_duration"`
	Pub                PubOutput       `json:"pub"`
	ActiveKeg          int             `json:"active_keg"`
	ActiveKegAt        time.Time       `json:"active_keg_at"`
	IsLow              bool            `json:"is_low"`
	Warehouse          []WarehouseItem `json:"warehouse"`
	WarehouseBeerLeft  int             `json:"warehouse_beer_left"`

	BankBalance      BalanceOutput       `json:"bank_balance"`
	BankTransactions []TransactionOutput `json:"bank_transactions"`

	BtDevicesLastOk time.Time  `json:"bt_devices_last_ok"`
	BtDevices       []BtDevice `json:"bt_devices"`
}

func (s *Scale) GetScale() FullOutput {
	s.mux.RLock()
	defer s.mux.RUnlock()

	warehouse := []WarehouseItem{
		{Keg: 10, Amount: s.warehouse[0]},
		{Keg: 15, Amount: s.warehouse[1]},
		{Keg: 20, Amount: s.warehouse[2]},
		{Keg: 30, Amount: s.warehouse[3]},
		{Keg: 50, Amount: s.warehouse[4]},
	}

	// Copy the transactions
	bt := make([]TransactionOutput, len(s.bank.transactions))
	copy(bt, s.bank.transactions)

	btDevices := make([]BtDevice, len(s.attendance.active))
	i := 0
	for _, device := range s.attendance.active {
		name := device.IdentityAddress
		knownName, f := s.attendance.known[device.IdentityAddress]
		if f && knownName != "" {
			name = knownName
		}

		btDevices[i] = BtDevice{
			Name:            name,
			IdentityAddress: device.IdentityAddress,
			RSSI:            device.RSSI,
			Known:           f,
			LastSeen:        utils.FormatDate(device.LastSeen),
		}
		i++
	}

	output := FullOutput{
		IsOk:               s.isOk(),
		BeersLeft:          s.beersLeft,
		BeersTotal:         s.getBeersTotal(),
		LastWeight:         s.weight,
		LastWeightFormated: fmt.Sprintf("%.2f", s.weight/1000),
		LastAt:             utils.FormatDate(s.weightAt),
		LastAtDuration:     durafmt.Parse(time.Since(s.weightAt).Round(time.Second)).LimitFirstN(2).Format(s.fmtUnits),
		Rssi:               s.rssi,
		LastUpdate:         utils.FormatDate(s.lastOk),
		LastUpdateDuration: durafmt.Parse(time.Since(s.lastOk).Round(time.Second)).LimitFirstN(2).Format(s.fmtUnits),
		Pub: PubOutput{
			IsOpen:   s.pub.isOpen,
			OpenedAt: utils.FormatDate(s.pub.openedAt),
			ClosedAt: utils.FormatDate(s.pub.closedAt),
		},
		ActiveKeg:         s.activeKeg,
		ActiveKegAt:       s.activeKegAt,
		IsLow:             s.isLow,
		Warehouse:         warehouse,
		WarehouseBeerLeft: GetWarehouseBeersLeft(s.warehouse),
		BankBalance:       s.bank.balance,
		BankTransactions:  bt,

		BtDevicesLastOk: s.attendance.lastOk,
		BtDevices:       btDevices,
	}

	return output
}
