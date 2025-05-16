package models

type Common struct {
	ID        uint `json:"-" form:"id"`
	CreatedAt uint `json:"-" form:"created_at"`
	UpdatedAt uint `json:"-" form:"updated_at"`
}

type Device struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Factory      string `json:"factory"`
	Phase        string `json:"phase"`
	Datacenter   string `json:"datacenter"`
	Room         string `json:"room"`
	Rack         string `json:"rack"`
	Side         string `json:"side"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Protocol     string `json:"protocol"`
	IP           string `json:"ip"`
	Common
	// 關聯到 DeviceLink 表
	DeviceLinks []DeviceLink `gorm:"foreignKey:ParentDeviceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"device_links"`
	// 關聯到 DeviceRule 表
	DeviceRules []DeviceRule `gorm:"foreignKey:Type;references:Type;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"device_rules"`
}

type DeviceLink struct {
	ID               uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentDeviceID   uint   `gorm:"index;foreignKey:ParentDeviceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"parent_device_id"`
	ChildDeviceID    uint   `gorm:"index;foreignKey:ChildDeviceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"child_device_id"`
	ParentDevicePort int    `json:"parent_device_port"`
	ParentDeviceType string `json:"parent_device_type"`
	Common
	// 關聯到父設備
	ParentDevice Device `gorm:"foreignKey:ParentDeviceID;references:ID" json:"parent_device"`
	// 關聯到子設備
	ChildDevice Device `gorm:"foreignKey:ChildDeviceID;references:ID" json:"child_device"`
}

type DeviceRule struct {
	ID            uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Enabled       bool    `json:"enabled"`
	Type          string  `gorm:"type:index" json:"type"` // 用於標識設備類型
	CheckPeriod   int     `json:"check_period"`
	Name          string  `json:"name"`
	Code          string  `json:"code"`
	Operator      string  `json:"operator"`
	InfoThreshold float64 `json:"info_threshold"`
	WarnThreshold float64 `json:"warn_threshold"`
	CritThreshold float64 `json:"crit_threshold"`
	Message       string  `json:"message"`
	Mode          string  `json:"mode"` // 區分運行環境（如 factory 或 dc）
	Common
}

type Group struct {
	Name string `gorm:"primaryKey" json:"name"`
	Common
	GroupZones []GroupZone `gorm:"foreignKey:GroupName;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"group_zones"`
}

type GroupZone struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	GroupName  string `json:"group_name"`
	Factory    string `json:"factory"`
	Phase      string `json:"phase"`
	Datacenter string `json:"datacenter"`
	Common
	Group Group `gorm:"foreignKey:GroupName;references:Name" json:"group"`
}

type Env struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Job struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Log struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
}
