package main

import (
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Codeplug struct {
	Version  string `yaml:"version"`
	Settings struct {
		// IntroLine1 string `yaml:"introLine1"`
		// IntroLine2 string `yaml:"introLine2"`
		// MicLevel   int    `yaml:"micLevel"`
		// Speech     bool   `yaml:"speech"`
		// Power      string `yaml:"power"`
		// Squelch    int    `yaml:"squelch"`
		// Vox        int    `yaml:"vox"`
		// Tot        int    `yaml:"tot"`
		// Anytone    struct {
		// 	SubChannel             bool   `yaml:"subChannel"`
		// 	SelectedVFO            string `yaml:"selectedVFO"`
		// 	ModeA                  string `yaml:"modeA"`
		// 	ModeB                  string `yaml:"modeB"`
		// 	VfoScanType            string `yaml:"vfoScanType"`
		// 	MinVFOScanFrequencyUHF string `yaml:"minVFOScanFrequencyUHF"`
		// 	MaxVFOScanFrequencyUHF string `yaml:"maxVFOScanFrequencyUHF"`
		// 	MinVFOScanFrequencyVHF string `yaml:"minVFOScanFrequencyVHF"`
		// 	MaxVFOScanFrequencyVHF string `yaml:"maxVFOScanFrequencyVHF"`
		// 	KeepLastCaller         bool   `yaml:"keepLastCaller"`
		// 	VfoStep                string `yaml:"vfoStep"`
		// 	SteType                string `yaml:"steType"`
		// 	SteFrequency           int    `yaml:"steFrequency"`
		// 	SteDuration            string `yaml:"steDuration"`
		// 	TbstFrequency          string `yaml:"tbstFrequency"`
		// 	ProMode                bool   `yaml:"proMode"`
		// 	MaintainCallChannel    bool   `yaml:"maintainCallChannel"`
		// 	BootSettings           struct {
		// 		BootDisplay         string                 `yaml:"bootDisplay"`
		// 		BootPasswordEnabled bool                   `yaml:"bootPasswordEnabled"`
		// 		BootPassword        string                 `yaml:"bootPassword"`
		// 		DefaultChannel      bool                   `yaml:"defaultChannel"`
		// 		GpsCheck            bool                   `yaml:"gpsCheck"`
		// 		Reset               bool                   `yaml:"reset"`
		// 		Additional          map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"bootSettings"`
		// 	PowerSaveSettings struct {
		// 		AutoShutdown            int                    `yaml:"autoShutdown"`
		// 		ResetAutoShutdownOnCall bool                   `yaml:"resetAutoShutdownOnCall"`
		// 		PowerSave               string                 `yaml:"powerSave"`
		// 		Atpc                    bool                   `yaml:"atpc"`
		// 		Additional              map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"powerSaveSettings"`
		// 	KeySettings struct {
		// 		FuncKey1Short     string                 `yaml:"funcKey1Short"`
		// 		FuncKey1Long      string                 `yaml:"funcKey1Long"`
		// 		FuncKey2Short     string                 `yaml:"funcKey2Short"`
		// 		FuncKey2Long      string                 `yaml:"funcKey2Long"`
		// 		FuncKey3Short     string                 `yaml:"funcKey3Short"`
		// 		FuncKey3Long      string                 `yaml:"funcKey3Long"`
		// 		FuncKey4Short     string                 `yaml:"funcKey4Short"`
		// 		FuncKey4Long      string                 `yaml:"funcKey4Long"`
		// 		FuncKey5Short     string                 `yaml:"funcKey5Short"`
		// 		FuncKey5Long      string                 `yaml:"funcKey5Long"`
		// 		FuncKey6Short     string                 `yaml:"funcKey6Short"`
		// 		FuncKey6Long      string                 `yaml:"funcKey6Long"`
		// 		FuncKeyAShort     string                 `yaml:"funcKeyAShort"`
		// 		FuncKeyALong      string                 `yaml:"funcKeyALong"`
		// 		FuncKeyBShort     string                 `yaml:"funcKeyBShort"`
		// 		FuncKeyBLong      string                 `yaml:"funcKeyBLong"`
		// 		FuncKeyCShort     string                 `yaml:"funcKeyCShort"`
		// 		FuncKeyCLong      string                 `yaml:"funcKeyCLong"`
		// 		FuncKeyDShort     string                 `yaml:"funcKeyDShort"`
		// 		FuncKeyDLong      string                 `yaml:"funcKeyDLong"`
		// 		LongPressDuration string                 `yaml:"longPressDuration"`
		// 		AutoKeyLock       bool                   `yaml:"autoKeyLock"`
		// 		KnobLock          bool                   `yaml:"knobLock"`
		// 		KeypadLock        bool                   `yaml:"keypadLock"`
		// 		SideKeysLock      bool                   `yaml:"sideKeysLock"`
		// 		ForcedKeyLock     bool                   `yaml:"forcedKeyLock"`
		// 		Additional        map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"keySettings"`
		// 	ToneSettings struct {
		// 		KeyTone       bool `yaml:"keyTone"`
		// 		KeyToneLevel  int  `yaml:"keyToneLevel"`
		// 		SmsAlert      bool `yaml:"smsAlert"`
		// 		CallAlert     bool `yaml:"callAlert"`
		// 		DmrTalkPermit bool `yaml:"dmrTalkPermit"`
		// 		DmrReset      bool `yaml:"dmrReset"`
		// 		FmTalkPermit  bool `yaml:"fmTalkPermit"`
		// 		DmrIdle       bool `yaml:"dmrIdle"`
		// 		FmIdle        bool `yaml:"fmIdle"`
		// 		Startup       bool `yaml:"startup"`
		// 		Tot           bool `yaml:"tot"`
		// 		CallMelody    struct {
		// 			Bpm        int                    `yaml:"bpm"`
		// 			Melody     string                 `yaml:"melody"`
		// 			Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 		} `yaml:"callMelody"`
		// 		IdleMelody struct {
		// 			Bpm        int                    `yaml:"bpm"`
		// 			Melody     string                 `yaml:"melody"`
		// 			Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 		} `yaml:"idleMelody"`
		// 		ResetMelody struct {
		// 			Bpm        int                    `yaml:"bpm"`
		// 			Melody     string                 `yaml:"melody"`
		// 			Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 		} `yaml:"resetMelody"`
		// 		CallEndMelody struct {
		// 			Bpm        int                    `yaml:"bpm"`
		// 			Melody     string                 `yaml:"melody"`
		// 			Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 		} `yaml:"callEndMelody"`
		// 		Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"toneSettings"`
		// 	DisplaySettings struct {
		// 		DisplayFrequency        bool                   `yaml:"displayFrequency"`
		// 		Brightness              int                    `yaml:"brightness"`
		// 		BacklightDuration       int                    `yaml:"backlightDuration"`
		// 		BacklightDurationTX     int                    `yaml:"backlightDurationTX"`
		// 		BacklightDurationRX     int                    `yaml:"backlightDurationRX"`
		// 		CustomChannelBackground bool                   `yaml:"customChannelBackground"`
		// 		VolumeChangePrompt      bool                   `yaml:"volumeChangePrompt"`
		// 		CallEndPrompt           bool                   `yaml:"callEndPrompt"`
		// 		ShowClock               bool                   `yaml:"showClock"`
		// 		ShowCall                bool                   `yaml:"showCall"`
		// 		ShowContact             bool                   `yaml:"showContact"`
		// 		ShowChannelNumber       bool                   `yaml:"showChannelNumber"`
		// 		ShowColorCode           bool                   `yaml:"showColorCode"`
		// 		ShowTimeSlot            bool                   `yaml:"showTimeSlot"`
		// 		ShowChannelType         bool                   `yaml:"showChannelType"`
		// 		ShowLastHeard           bool                   `yaml:"showLastHeard"`
		// 		LastCallerDisplay       string                 `yaml:"lastCallerDisplay"`
		// 		CallColor               string                 `yaml:"callColor"`
		// 		StandbyTextColor        string                 `yaml:"standbyTextColor"`
		// 		StandbyBackgroundColor  string                 `yaml:"standbyBackgroundColor"`
		// 		ChannelNameColor        string                 `yaml:"channelNameColor"`
		// 		ChannelBNameColor       string                 `yaml:"channelBNameColor"`
		// 		ZoneNameColor           string                 `yaml:"zoneNameColor"`
		// 		ZoneBNameColor          string                 `yaml:"zoneBNameColor"`
		// 		Language                string                 `yaml:"language"`
		// 		DateFormat              string                 `yaml:"dateFormat"`
		// 		Additional              map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"displaySettings"`
		// 	AudioSettings struct {
		// 		VoxDelay           string                 `yaml:"voxDelay"`
		// 		VoxSource          string                 `yaml:"voxSource"`
		// 		Recording          bool                   `yaml:"recording"`
		// 		Enhance            bool                   `yaml:"enhance"`
		// 		MuteDelay          string                 `yaml:"muteDelay"`
		// 		MaxVolume          int                    `yaml:"maxVolume"`
		// 		MaxHeadPhoneVolume int                    `yaml:"maxHeadPhoneVolume"`
		// 		EnableFMMicGain    bool                   `yaml:"enableFMMicGain"`
		// 		FmMicGain          int                    `yaml:"fmMicGain"`
		// 		Additional         map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"audioSettings"`
		// 	MenuSettings struct {
		// 		Duration   string                 `yaml:"duration"`
		// 		Separator  bool                   `yaml:"separator"`
		// 		Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"menuSettings"`
		// 	AutoRepeaterSettings struct {
		// 		DirectionA string                 `yaml:"directionA"`
		// 		DirectionB string                 `yaml:"directionB"`
		// 		VhfMin     string                 `yaml:"vhfMin"`
		// 		VhfMax     string                 `yaml:"vhfMax"`
		// 		UhfMin     string                 `yaml:"uhfMin"`
		// 		UhfMax     string                 `yaml:"uhfMax"`
		// 		Vhf2Min    string                 `yaml:"vhf2Min"`
		// 		Vhf2Max    string                 `yaml:"vhf2Max"`
		// 		Uhf2Min    string                 `yaml:"uhf2Min"`
		// 		Uhf2Max    string                 `yaml:"uhf2Max"`
		// 		Offsets    []interface{}          `yaml:"offsets"`
		// 		Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"autoRepeaterSettings"`
		// 	DmrSettings struct {
		// 		GroupCallHangTime         string                 `yaml:"groupCallHangTime"`
		// 		ManualGroupCallHangTime   string                 `yaml:"manualGroupCallHangTime"`
		// 		PrivateCallHangTime       string                 `yaml:"privateCallHangTime"`
		// 		ManualPrivateCallHangTime string                 `yaml:"manualPrivateCallHangTime"`
		// 		PreWaveDelay              int                    `yaml:"preWaveDelay"`
		// 		WakeHeadPeriod            int                    `yaml:"wakeHeadPeriod"`
		// 		FilterOwnID               bool                   `yaml:"filterOwnID"`
		// 		MonitorSlotMatch          string                 `yaml:"monitorSlotMatch"`
		// 		MonitorColorCodeMatch     bool                   `yaml:"monitorColorCodeMatch"`
		// 		MonitorIDMatch            bool                   `yaml:"monitorIDMatch"`
		// 		MonitorTimeSlotHold       bool                   `yaml:"monitorTimeSlotHold"`
		// 		SmsFormat                 string                 `yaml:"smsFormat"`
		// 		SendTalkerAlias           bool                   `yaml:"sendTalkerAlias"`
		// 		TalkerAliasSource         string                 `yaml:"talkerAliasSource"`
		// 		TalkerAliasEncoding       string                 `yaml:"talkerAliasEncoding"`
		// 		Encryption                string                 `yaml:"encryption"`
		// 		Additional                map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"dmrSettings"`
		// 	GpsSettings struct {
		// 		Units          string                 `yaml:"units"`
		// 		TimeZone       string                 `yaml:"timeZone"`
		// 		ReportPosition bool                   `yaml:"reportPosition"`
		// 		UpdatePeriod   string                 `yaml:"updatePeriod"`
		// 		Mode           string                 `yaml:"mode"`
		// 		Additional     map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"gpsSettings"`
		// 	RoamingSettings struct {
		// 		AutoRoam          bool                   `yaml:"autoRoam"`
		// 		AutoRoamPeriod    string                 `yaml:"autoRoamPeriod"`
		// 		AutoRoamDelay     int                    `yaml:"autoRoamDelay"`
		// 		RoamStart         string                 `yaml:"roamStart"`
		// 		RoamReturn        string                 `yaml:"roamReturn"`
		// 		RangeCheck        bool                   `yaml:"rangeCheck"`
		// 		CheckInterval     string                 `yaml:"checkInterval"`
		// 		RetryCount        int                    `yaml:"retryCount"`
		// 		OutOfRangeAlert   string                 `yaml:"outOfRangeAlert"`
		// 		Notification      bool                   `yaml:"notification"`
		// 		NotificationCount int                    `yaml:"notificationCount"`
		// 		GpsRoaming        bool                   `yaml:"gpsRoaming"`
		// 		Additional        map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"roamingSettings"`
		// 	BluetoothSettings struct {
		// 		PttLatch      bool                   `yaml:"pttLatch"`
		// 		PttSleepTimer int                    `yaml:"pttSleepTimer"`
		// 		Additional    map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"bluetoothSettings"`
		// 	SimplexRepeaterSettings struct {
		// 		Enabled    bool                   `yaml:"enabled"`
		// 		Monitor    bool                   `yaml:"monitor"`
		// 		TimeSlot   string                 `yaml:"timeSlot"`
		// 		Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"simplexRepeaterSettings"`
		// 	Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// } `yaml:"anytone,omitempty"`
		// DefaultID  string                 `yaml:"defaultID,omitempty"`
		Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
	} `yaml:"settings"`
	RadioIDs []struct {
		Dmr struct {
			ID         string                 `yaml:"id"`
			Name       string                 `yaml:"name"`
			Number     int                    `yaml:"number"`
			Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		} `yaml:"dmr,flow"`
		Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
	} `yaml:"radioIDs"`
	Contacts    []*Contact   `yaml:"contacts"`
	GroupLists  []*GroupList `yaml:"groupLists"`
	Channels    []*Channel   `yaml:"channels"`
	Zones       []*Zone      `yaml:"zones"`
	Positioning []struct {
		// Aprs struct {
		// 	ID      string `yaml:"id"`
		// 	Name    string `yaml:"name"`
		// 	Period  int    `yaml:"period"`
		// 	Icon    string `yaml:"icon"`
		// 	Message string `yaml:"message"`
		// 	Anytone struct {
		// 		TxDelay        string `yaml:"txDelay"`
		// 		PreWaveDelay   string `yaml:"preWaveDelay"`
		// 		PassAll        bool   `yaml:"passAll"`
		// 		ReportPosition bool   `yaml:"reportPosition"`
		// 		ReportMicE     bool   `yaml:"reportMicE"`
		// 		ReportObject   bool   `yaml:"reportObject"`
		// 		ReportItem     bool   `yaml:"reportItem"`
		// 		ReportMessage  bool   `yaml:"reportMessage"`
		// 		ReportWeather  bool   `yaml:"reportWeather"`
		// 		ReportNMEA     bool   `yaml:"reportNMEA"`
		// 		ReportStatus   bool   `yaml:"reportStatus"`
		// 		ReportOther    bool   `yaml:"reportOther"`
		// 		Frequencies    []struct {
		// 			ID         string                 `yaml:"id"`
		// 			Name       string                 `yaml:"name"`
		// 			Frequency  string                 `yaml:"frequency"`
		// 			Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 		} `yaml:"frequencies"`
		// 		Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// 	} `yaml:"anytone"`
		// 	Destination string                 `yaml:"destination"`
		// 	Source      string                 `yaml:"source"`
		// 	Path        []string               `yaml:"path,flow"`
		// 	Additional  map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
		// } `yaml:"aprs"`
		Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
	} `yaml:"positioning,omitempty"`
	RoamingChannels []struct {
		ID          string                 `yaml:"id"`
		Name        string                 `yaml:"name"`
		RxFrequency string                 `yaml:"rxFrequency"`
		TxFrequency string                 `yaml:"txFrequency"`
		Additional  map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
	} `yaml:"roamingChannels,omitempty"`
	RoamingZones []struct {
		ID         string                 `yaml:"id"`
		Name       string                 `yaml:"name"`
		Channels   []string               `yaml:"channels"`
		Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
	} `yaml:"roamingZones,omitempty"`
	Commercial struct {
		EncryptionKeys []interface{}          `yaml:"encryptionKeys"`
		Additional     map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
	} `yaml:"commercial"`
	Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
}

type Channel struct {
	Digital    Digital                `yaml:"digital,omitempty"`
	Analog     Analog                 `yaml:"analog,omitempty"`
	Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
}

func (c Channel) GetID() string {
	if c.Analog.ID != "" {
		return c.Analog.ID
	}
	return c.Digital.ID
}

type Digital struct {
	ID          string         `yaml:"id"`
	Name        string         `yaml:"name"`
	RxFrequency string         `yaml:"rxFrequency"`
	TxFrequency string         `yaml:"txFrequency"`
	RxOnly      bool           `yaml:"rxOnly"`
	Admit       string         `yaml:"admit"`
	ColorCode   int            `yaml:"colorCode"`
	TimeSlot    string         `yaml:"timeSlot"`
	RadioID     DefaultableInt `yaml:"radioId"`
	GroupList   string         `yaml:"groupList"`
	Contact     string         `yaml:"contact"`
	Anytone     struct {
		// Talkaround          bool                   `yaml:"talkaround"`
		// FrequencyCorrection int                    `yaml:"frequencyCorrection"`
		// HandsFree           bool                   `yaml:"handsFree"`
		// CallConfirm         bool                   `yaml:"callConfirm"`
		// Sms                 bool                   `yaml:"sms"`
		// SmsConfirm          bool                   `yaml:"smsConfirm"`
		// DataACK             bool                   `yaml:"dataACK"`
		// SimplexTDMA         bool                   `yaml:"simplexTDMA"`
		// AdaptiveTDMA        bool                   `yaml:"adaptiveTDMA"`
		// LoneWorker          bool                   `yaml:"loneWorker"`
		// ThroughMode         bool                   `yaml:"throughMode"`
		Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
	} `yaml:"anytone"`
	Power      DefaultableString      `yaml:"power"`
	Timeout    DefaultableInt         `yaml:"timeout"`
	Vox        DefaultableInt         `yaml:"vox"`
	Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
}

type Analog struct {
	ID          string                 `yaml:"id"`
	Name        string                 `yaml:"name"`
	RxFrequency string                 `yaml:"rxFrequency"`
	TxFrequency string                 `yaml:"txFrequency"`
	RxOnly      bool                   `yaml:"rxOnly"`
	Admit       string                 `yaml:"admit"`
	Bandwidth   string                 `yaml:"bandwidth"`
	Power       DefaultableString      `yaml:"power"`
	Timeout     DefaultableInt         `yaml:"timeout"`
	Vox         DefaultableInt         `yaml:"vox"`
	RxTone      Tone                   `yaml:"rxTone,flow,omitempty"`
	TxTone      Tone                   `yaml:"txTone,flow,omitempty"`
	Squelch     DefaultableInt         `yaml:"squelch"`
	Additional  map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
}

type Zone struct {
	ID         string                 `yaml:"id"`
	Name       string                 `yaml:"name"`
	A          []string               `yaml:"A,flow"`
	B          []string               `yaml:"B,flow"`
	Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
}

func (z Zone) GetID() string {
	return z.ID
}

type Tone struct {
	CTCSS float64 `yaml:"ctcss,omitempty"`
	DCS   float64 `yaml:"dcs,omitempty"`
}

func (t *Tone) Set(val string) error {
	var err error
	if strings.HasPrefix(val, "D") {
		t.DCS, err = strconv.ParseFloat(val[1:], 64)
	} else if val != "" {
		t.CTCSS, err = strconv.ParseFloat(val, 64)
	}
	return err
}

type Contact struct {
	DMR  DMR  `yaml:"dmr,flow,omitempty"`
	DTMF DTMF `yaml:"dtmf,flow,omitempty"`
}

func (c Contact) GetID() string {
	if c.DTMF.ID != "" {
		return c.DTMF.ID
	}
	return c.DMR.ID
}

type DMR struct {
	ID         string                 `yaml:"id"`
	Name       string                 `yaml:"name"`
	Ring       bool                   `yaml:"ring"`
	Type       string                 `yaml:"type"`
	Number     int                    `yaml:"number"`
	Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
}

type DTMF struct {
	ID         string                 `yaml:"id"`
	Name       string                 `yaml:"name"`
	Ring       bool                   `yaml:"ring"`
	Number     int                    `yaml:"number"`
	Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
}

type GroupList struct {
	ID         string                 `yaml:"id"`
	Name       string                 `yaml:"name"`
	Contacts   []string               `yaml:"contacts"`
	Additional map[string]interface{} `yaml:",inline"` // Any new keys will show up here to be roundtripped
}

func (g GroupList) GetID() string {
	return g.ID
}

type DefaultableInt struct {
	Value    int
	HasValue bool
}

func (di *DefaultableInt) UnmarshalYAML(n *yaml.Node) error {
	var err error
	// logVeryVerbose("DefaultableInt, node %#v", n)
	if n.Tag == "!default" {
		di.HasValue = false
	} else {
		di.HasValue = true
		di.Value, err = strconv.Atoi(n.Value)
	}
	return err
}

func (di DefaultableInt) MarshalYAML() (interface{}, error) {
	// logVeryVerbose("DefaultableInt.MarshalYAML(): %#v", di)
	if !di.HasValue {
		return yaml.Node{Kind: 0x8, Style: 0x3, Tag: "!default"}, nil
	}
	return yaml.Node{Kind: 0x8, Style: 0x0, Tag: "!!int", Value: strconv.Itoa(di.Value)}, nil
}

type DefaultableString struct {
	Value    string
	HasValue bool
}

func (ds *DefaultableString) UnmarshalYAML(n *yaml.Node) error {
	// logVeryVerbose("DefaultableString, node %#v", n)
	if n.Tag == "!default" {
		ds.HasValue = false
	} else {
		ds.HasValue = true
		ds.Value = n.Value
	}
	return nil
}

func (ds DefaultableString) MarshalYAML() (interface{}, error) {
	if !ds.HasValue {
		return yaml.Node{Kind: yaml.ScalarNode, Style: 0x3, Tag: "!default"}, nil
	}
	return yaml.Node{Kind: yaml.ScalarNode, Style: 0x0, Tag: "!!str", Value: ds.Value}, nil
}
