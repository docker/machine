package linodego

import (
	"encoding/json"
	_ "fmt"
)

type Response struct {
	Errors  []Error         `json:"ERRORARRAY"`
	RawData json.RawMessage `json:"DATA"`
	Action  string          `json:"ACTION"`
}

type Error struct {
	ErrorCode    int    `json:"ERRORCODE"`
	ErrorMessage string `json:"ERRORMESSAGE"`
}

type DataCenter struct {
	DataCenterId int    `json:"DATACENTERID"`
	Location     string `json:"LOCATION"`
	Abbr         string `json:"ABBR"`
}

type Distribution struct {
	Is64Bit             int          `json:"IS64BIT"`
	Label               CustomString `json:"LABEL"`
	MinImageSize        int          `json:"MINIMAGESIZE"`
	DistributionId      int          `json:"DISTRIBUTIONID"`
	CreatedDt           CustomTime   `json:"CREATE_DT"`
	RequiresPVOPSKernel int          `json:"REQUIRESPVOPSKERNEL"`
}

type Kernel struct {
	Label    CustomString `json:"LABEL"`
	IsXen    int          `json:"ISXEN"`
	IsPVOPS  int          `json:"ISPVOPS"`
	KernelId int          `json:"KERNELID"`
}

type LinodePlan struct {
	Cores  int            `json:"CORES"`
	Price  float32        `json:"PRICE"`
	RAM    int            `json:"RAM"`
	Xfer   int            `json:"Xfer"`
	PlanId int            `json:"PLANID"`
	Label  CustomString   `json:"LABEL"`
	Avail  map[string]int `json:"AVAIL"`
	Disk   int            `json:"DISK"`
	Hourly float32        `json:"HOURLY"`
}

type NodeBalancer struct {
	Hourly      float32 `json:"HOURLY"`
	Monthly     float32 `json:"MONTHLY"`
	Connections int     `json:"CONNECTIONS"`
}

type StackScript struct {
	//Script string `json:"SCRIPT"`
	//Description string `json:"DESCRIPTION"`
	//DistributionidList int `json:"DISTRIBUTIONIDLIST"`
	Label             CustomString `json:"LABEL"`
	DeploymentsTotal  int          `json:"DEPLOYMENTSTOTAL"`
	LatestRev         int          `json:"LATESTREV"`
	CreatedDt         CustomTime   `json:"CREATE_DT"`
	DeploymentsActive int          `json:"DEPLOYMENTSACTIVE"`
	StackScriptId     int          `json:"STACKSCRIPTID"`
	RevNote           int          `json:"REV_NOTE"`
	RevDt             int          `json:"REV_DT"`
	IsPublic          int          `json:"ISPUBLIC"`
	UserId            int          `json:"USERID"`
}

type EstimateInvoice struct {
	InvoiceTo CustomShortTime `json:"INVOICE_TO"`
	Amount    float32         `json:"AMOUNT"`
}

type AccountInfo struct {
	AccountSince     CustomTime `json:"ACTIVE_SINCE"`
	TransferPool     int        `json:"TRANSFER_POOL"`
	TransferUsed     int        `json:"TRANSFER_USED"`
	TransferBillable int        `json:"TRANSFER_BILLABLE"`
	BillingMethod    string     `json:"BILLING_METHOD"`
	Managed          bool       `json:"MANAGED"`
	Balance          float32    `json:"BALANCE"`
}

type Image struct {
	CreateDt    CustomTime   `json:"CREATE_DT"`
	Creator     string       `json:"CREATOR"`
	Description string       `json:"DESCRIPTION"`
	FsType      string       `json:"FS_TYPE"`
	ImageId     int          `json:"IMAGEID"`
	IsPublic    int          `json:"ISPUBLIC"`
	Label       CustomString `json:"LABEL"`
	LastUsedDt  CustomTime   `json:"LAST_USED_DT"`
	MinSize     int          `json:"MINSIZE"`
	Status      string       `json:"STATUS"`
	Type        string       `json:"TYPE"`
}

type Linode struct {
	TotalXFer             int          `json:"TOTALXFER"`
	BackupsEnabled        int          `json:"BACKUPSENABLED"`
	WatchDog              int          `json:"WATCHDOG"`
	LpmDisplayGroup       string       `json:"LPM_DISPLAYGROUP"`
	AlertBwQuotaEnabled   int          `json:"ALERT_BWQUOTA_ENABLED"`
	Status                int          `json:"STATUS"`
	TotalRAM              int          `json:"TOTALRAM"`
	AlertDiskIOThreshold  int          `json:"ALERT_DISKIO_THRESHOLD"`
	BackupWindow          int          `json:"BACKUPWINDOW"`
	AlertBwOutEnabled     int          `json:"ALERT_BWOUT_ENABLED"`
	AlertBwOutThreshold   int          `json:"ALERT_BWOUT_THRESHOLD"`
	Label                 CustomString `json:"LABEL"`
	AlertCPUEnabled       int          `json:"ALERT_CPU_ENABLED"`
	AlertBwQuotaThreshold int          `json:"ALERT_BWQUOTA_THRESHOLD"`
	AlertBwInThreshold    int          `json:"ALERT_BWIN_THRESHOLD"`
	BackupWeeklyDay       int          `json:"BACKUPWEEKLYDAY"`
	DataCenterId          int          `json:"DATACENTERID"`
	AlertCPUThreshold     int          `json:"ALERT_CPU_THRESHOLD"`
	TotalHD               int          `json:"TOTALHD"`
	AlertDiskIOEnabled    int          `json:"ALERT_DISKIO_ENABLED"`
	AlertBwInEnabled      int          `json:"ALERT_BWIN_ENABLED"`
	LinodeId              int          `json:"LINODEID"`
	CreateDt              CustomTime   `json:"CREATE_DT"`
	PlanId                int          `json:"PLANID"`
	DistributionVendor    string       `json:"DISTRIBUTIONVENDOR"`
}

type LinodeId struct {
	LinodeId int `json:"LinodeID"`
}

type Job struct {
	EnteredDt    CustomTime   `json:"ENTERED_DT"`
	Action       string       `json:"ACTION"`
	Label        string       `json:"LABEL"`
	HostStartDt  CustomTime   `json:"HOST_START_DT"`
	LinodeId     int          `json:"LINODEID"`
	HostFinishDt CustomTime   `json:"HOST_FINISH_DT"`
	HostMessage  string       `json:"HOST_MESSAGE"`
	JobId        int          `json:"JOBID"`
	HostSuccess  CustomString `json:"HOST_SUCCESS"` // Linode API returns empty string if HostSuccess is false. 1 otherwise.
}

type LinodeConfig struct {
	HelperDisableUpdateDB int          `json:"helper_disableUpdateDB"`
	RootDeviceRO          bool         `json:"RootDeviceRO"`
	RootDeviceCustom      string       `json:"RootDeviceCustom"`
	Label                 CustomString `json:"Label"`
	DiskList              string       `json:"DiskList"`
	LinodeId              int          `json:"LinodeID"`
	Comments              string       `json:"Comments"`
	ConfigId              string       `json:"ConfigID"`
	HelperXen             int          `json:"helper_xen"`
	RunLevel              string       `json:"RunLevel"`
	HelperDepmod          string       `json:"helper_depmod"`
	KernelId              int          `json:"KernelID"`
	RootDeviceNum         int          `json:"RootDeviceNum"`
	HelperLibtls          bool         `json:"helper_libtls"`
	RAMLimit              int          `json:"RAMLimit"`
}

type LinodeConfigId struct {
	LinodeConfigId int `json:"ConfigID"`
}

type JobId struct {
	JobId int `json:"JobID"`
}

type DiskJob struct {
	JobId  int `json:"JobID"`
	DiskId int `json:"DiskID"`
}

type Disk struct {
	UpdateDt   CustomTime   `json:"UPDATE_DT"`
	DiskId     int          `json:"DISKID"`
	Label      CustomString `json:"LABEL"`
	Type       string       `json:"TYPE"`
	LinodeId   int          `json:"LINODEID"`
	IsReadOnly int          `json:"ISREADONLY"`
	Status     int          `json:"STATUS"`
	CreateDt   CustomTime   `json:"CREATE_DT"`
	Size       int          `json:"SIZE"`
}

type IPAddress struct {
	IPAddress   string `json:"IPAddress"`
	IPAddressId int    `json:"IPAddressID"`
}

type FullIPAddress struct {
	LinodeId    int    `json:"LINODEID"`
	IsPublic    int    `json:"ISPUBLIC"`
	RDNSName    string `json:"RDNS_NAME"`
	IPAddress   string `json:"IPADDRESS"`
	IPAddressId int    `json:"IPADDRESSID"`
}

type RDNSIPAddress struct {
	HostName    string `json:"HOSTNAME"`
	IPAddress   string `json:"IPADDRESS"`
	IPAddressId int    `json:"IPADDRESSID"`
}

type LinodeIPAddress struct {
	LinodeId    int    `json:"LINODEID"`
	IPAddress   string `json:"IPADDRESS"`
	IPAddressId int    `json:"IPADDRESSID"`
}
