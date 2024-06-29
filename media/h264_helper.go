package media

import (
	"encoding/base64"
	"fmt"
	"github.com/deepch/vdk/codec/h264parser"
	"strconv"
)

type H264Profile int

const (
	UnknownProfile H264Profile = 0
	CBaseLine      H264Profile = 1
	BaseLine       H264Profile = 2
	Main           H264Profile = 3
	Extended       H264Profile = 4
	High           H264Profile = 5
	High10         H264Profile = 6
	High42         H264Profile = 7
	High44         H264Profile = 8
	ProfileEnd     H264Profile = 9
)

func (p H264Profile) Name() string {
	return gProfileMaps[p].name
}

type H264ProfileLevel int

const (
	Lvl0   H264ProfileLevel = 0
	Lvl1   H264ProfileLevel = 1  // 176x144@15
	Lvl1b  H264ProfileLevel = 2  // 176x144@60
	Lvl1_1 H264ProfileLevel = 3  // 352x288@7.5
	Lvl1_2 H264ProfileLevel = 4  // 352x288@15
	Lvl1_3 H264ProfileLevel = 5  // 352x288@30
	Lvl2   H264ProfileLevel = 6  // 352x288@30
	Lvl2_1 H264ProfileLevel = 7  // 352x576@25
	Lvl2_2 H264ProfileLevel = 8  // 720x576@12.5
	Lvl3   H264ProfileLevel = 9  // 720x576@25
	Lvl3_1 H264ProfileLevel = 10 // 1280x720@30
	Lvl3_2 H264ProfileLevel = 11 // 1280x720@60
	Lvl4   H264ProfileLevel = 12 // 2048x1024@30
	Lvl4_1 H264ProfileLevel = 13 // 2048x1024@30
	Lvl4_2 H264ProfileLevel = 14 // 2048x1080@60
	Lvl5   H264ProfileLevel = 15 // 3680x1536@25
	Lvl5_1 H264ProfileLevel = 16 // 4096x2048@30
	Lvl5_2 H264ProfileLevel = 17 // 4096x2160@60
	LvlEnd H264ProfileLevel = 18
)

func (l H264ProfileLevel) Name() string {
	return gProfileLevelMaps[l].name
}

func GetProfileByIdcIop(idc, srcIop uint8, strict bool) H264Profile {
	// search for idc
	for p := CBaseLine; p <= ProfileEnd; p += 1 {
		cfgTmp := gProfileMaps[p].cfg1
		if idc == cfgTmp.idc {
			value := srcIop & cfgTmp.mask
			if value == cfgTmp.value || !strict {
				return p
			}
		}

		cfgTmp = gProfileMaps[p].cfg2
		if idc == cfgTmp.idc {
			value := srcIop & cfgTmp.mask
			if value == cfgTmp.value || !strict {
				return p
			}
		}

		cfgTmp = gProfileMaps[p].cfg3
		if idc == cfgTmp.idc {
			value := srcIop & cfgTmp.mask
			if value == cfgTmp.value || !strict {
				return p
			}
		}
	}

	return UnknownProfile
}

func GetProfileLevelByLvlIdc(idc uint8) H264ProfileLevel {
	for l := Lvl1; l < LvlEnd; l += 1 {
		cfg := gProfileLevelMaps[l]
		if cfg.value == idc {
			return l
		}
	}

	return Lvl0
}

func ParseProfileLevelID(profileLevelID string) (error, H264Profile, H264ProfileLevel, uint64, uint64, uint64) {
	if len(profileLevelID) != 6 {
		return fmt.Errorf("invalid profileLevelID: " + profileLevelID), UnknownProfile, Lvl0, 0, 0, 0
	}

	idc, err1 := strconv.ParseUint(profileLevelID[:2], 16, 8)
	iop, err2 := strconv.ParseUint(profileLevelID[2:4], 16, 8)
	lvl, err3 := strconv.ParseUint(profileLevelID[4:], 16, 8)
	if err1 != nil {
		return err1, UnknownProfile, Lvl0, 0, 0, 0
	}
	if err2 != nil {
		return err2, UnknownProfile, Lvl0, 0, 0, 0
	}
	if err3 != nil {
		return err3, UnknownProfile, Lvl0, 0, 0, 0
	}

	profile := GetProfileByIdcIop(uint8(idc), uint8(iop), true)
	if profile == UnknownProfile {
		return fmt.Errorf("get profile id failed! profile_level_id is %s", profileLevelID), UnknownProfile, Lvl0, 0, 0, 0
	}

	level := GetProfileLevelByLvlIdc(uint8(lvl))
	if level == Lvl0 {
		return fmt.Errorf("get profile level id failed! %s", profileLevelID), UnknownProfile, Lvl0, 0, 0, 0
	}

	return nil, profile, level, 0, 0, 0
}

func GenProfileLevelID(profile H264Profile, level H264ProfileLevel) string {
	proCfg := gProfileMaps[profile]
	lvlCfg := gProfileLevelMaps[level]
	profileLevelID := fmt.Sprintf("%02x%02x%02x", proCfg.defaultIdc, proCfg.defaultIop, lvlCfg.value)
	return profileLevelID
}

func ParseBase64SPS(sps string) (error, []byte, *h264parser.SPSInfo) {
	spsData, err := base64.StdEncoding.DecodeString(sps)
	if err != nil {
		return err, nil, nil
	}

	err, info := ParseSPS(spsData)
	return err, spsData, info
}

func ParseSPS(sps []byte) (error, *h264parser.SPSInfo) {
	info, err := h264parser.ParseSPS(sps)
	return err, &info
}

func GenSpropParameterSets(sps, pps []byte) (string, string, string) {
	str1 := base64.StdEncoding.EncodeToString(sps)
	str2 := base64.StdEncoding.EncodeToString(pps)
	return str1 + "," + str2, str1, str2
}

type pcConfig struct {
	idc   uint8
	mask  uint8
	value uint8
}

type profileCfg struct {
	name       string
	cfg1       pcConfig
	cfg2       pcConfig
	cfg3       pcConfig
	defaultIdc uint8
	defaultIop uint8
}

type profileLvlCfg struct {
	name                 string
	value                uint8
	maxFrameSizeInMB     uint32
	maxDecodingSpeedInMB uint32
	maxDPBInMB           uint32
}

var gProfileMaps = map[H264Profile]profileCfg{
	/*
	 * Profile     profile_idc        profile-iop             value
	 *             (hexadecimal)      (binary) (mask)
	 * CB          42 (B)             x1xx0000 01001111(4F) : x1xx0000 & 01001111 == 0x40
	 *    same as: 4D (M)             1xxx0000 10001111(8F) : 1xxx0000 & 10001111 == 0x80
	 *    same as: 58 (E)             11xx0000 11001111(CF) : 11xx0000 & 11001111 == 0xC0
	 * B           42 (B)             x0xx0000 01001111(4F) : x0xx0000 & 01001111 == 0x00
	 *    same as: 58 (E)             10xx0000 11001111(CF) : 10xx0000 & 11001111 == 0x80
	 * M           4D (M)             0x0x0000 10101111(AF) : 0x0x0000 & 10101111 == 0x00
	 * E           58                 00xx0000 11001111(CF) : 00xx0000 & 11001111 == 0x00
	 * H           64                 00000000 11111111(FF) : 00000000 & 11111111 == 0x00
	 * H10         6E                 00000000 11111111(FF) : 00000000 & 11111111 == 0x00
	 * H42         7A                 00000000 11111111(FF) : 00000000 & 11111111 == 0x00
	 * H44         F4                 00000000 11111111(FF) : 00000000 & 11111111 == 0x00
	 * H10I        6E                 00010000 11111111(FF) : 00010000 & 11111111 == 0x10
	 * H42I        7A                 00010000 11111111(FF) : 00010000 & 11111111 == 0x10
	 * H44I        F4                 00010000 11111111(FF) : 00010000 & 11111111 == 0x10
	 * C44I        2C                 00010000 11111111(FF) : 00010000 & 11111111 == 0x10
	 * */
	//H264Profile                    name       config
	CBaseLine:      {"CBP", pcConfig{0x42 /*66*/, 0x4F, 0x40}, pcConfig{0x4D /*77*/, 0x8F, 0x80}, pcConfig{0x58 /*88*/, 0xCF, 0xC0}, 0x42, 0x40},
	BaseLine:       {"BP", pcConfig{0x42 /*66*/, 0x4F, 0x00}, pcConfig{0x58 /*88*/, 0xCF, 0x80}, pcConfig{0x00, 0x00, 0x00}, 0x42, 0x80},
	Main:           {"MP", pcConfig{0x4D /*77*/, 0xAF, 0x00}, pcConfig{0x00, 0x00, 0x00}, pcConfig{0x00, 0x00, 0x00}, 0x4d, 0x00},
	Extended:       {"XP", pcConfig{0x58 /*88*/, 0xCF, 0x00}, pcConfig{0x00, 0x00, 0x00}, pcConfig{0x00, 0x00, 0x00}, 0x58, 0x00},
	High:           {"HiP", pcConfig{0x64 /*100*/, 0xFF, 0x00}, pcConfig{0x00, 0x00, 0x00}, pcConfig{0x00, 0x00, 0x00}, 0x64, 0x00},
	High10:         {"Hi10P", pcConfig{0x6E /*110*/, 0xFF, 0x00}, pcConfig{0x00, 0x00, 0x00}, pcConfig{0x00, 0x00, 0x00}, 0x6e, 0x00},
	High42:         {"Hi422P", pcConfig{0x7A /*122*/, 0xFF, 0x00}, pcConfig{0x00, 0x00, 0x00}, pcConfig{0x00, 0x00, 0x00}, 0x7a, 0x00},
	High44:         {"Hi444PP", pcConfig{0xF4 /*244*/, 0xFF, 0x00}, pcConfig{0x00, 0x00, 0x00}, pcConfig{0x00, 0x00, 0x00}, 0xf4, 0x00},
	UnknownProfile: {"", pcConfig{0x00 /*0*/, 0x00, 0x00}, pcConfig{0x00, 0x00, 0x00}, pcConfig{0x00, 0x00, 0x00}, 0x00, 0x00},
}

var gProfileLevelMaps = map[H264ProfileLevel]profileLvlCfg{
	//H264ProfileLevel               name   value  MB/frame MB/s    MB/buffer
	Lvl1:   {"1", 10, 99, 1485, 396},
	Lvl1b:  {"1b", 9, 99, 1485, 396},
	Lvl1_1: {"1.1", 11, 396, 3000, 900},
	Lvl1_2: {"1.2", 12, 396, 6000, 2376},
	Lvl1_3: {"1.3", 13, 396, 11880, 2376},
	Lvl2:   {"2", 20, 396, 11880, 2376},
	Lvl2_1: {"2.1", 21, 792, 19800, 4752},
	Lvl2_2: {"2.2", 22, 1620, 20250, 8100},
	Lvl3:   {"3", 30, 1620, 40500, 8100},
	Lvl3_1: {"3.1", 31, 3600, 108000, 18000},
	Lvl3_2: {"3.2", 32, 5120, 216000, 20480},
	Lvl4:   {"4", 40, 8192, 245760, 32768},
	Lvl4_1: {"4.1", 41, 8192, 245760, 32768},
	Lvl4_2: {"4.2", 42, 8704, 522240, 34816},
	Lvl5:   {"5", 50, 22080, 589824, 110400},
	Lvl5_1: {"5.1", 51, 36864, 983040, 184320},
	Lvl5_2: {"5.2", 52, 36864, 2073600, 184320},
	Lvl0:   {"", 0, 0, 0, 0},
}
