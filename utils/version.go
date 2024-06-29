package utils

import (
	"fmt"
	"strings"
	"time"
)

type Building struct {
	IsDebug         bool   // 是否为DEBUG版本
	AppShortId      string // 应用短ID, E.G.: [mcs]
	AppId           string // 应用ID, E.G.: [avd/zerogo/app/mcs2]
	VersionName     string // 版本名称, E.G.: [1.x.x]
	VersionCode     string // 版本ID, E.G.: [20231121.86400]
	VersionFullName string // 版本全称, E.G.: [1.x.x-20231121.86400]
	VersionSHA      string // 版本SHA值(GIT ID), E.G.: [2c0866ef0]
	BuildType       string // 打包类型, E.G.: [default]
	BuildTime       string // 打包时间, E.G.: [2023.11.21 14:18:20]
	BuildHost       string // 版本来源, E.G.: [Tuyj-T470p]
	Flavor          string // 渠道标记, E.G.: [S]
}

var AppBuilding = Building{}

func LoadBuilding(IsDebug bool, AppShortId, AppId, VersionName,
	VersionSHA, BuildType, BuildTime, BuildHost, Flavor string) {

	BuildTime = strings.Replace(BuildTime, "_", " ", 1)
	versionCode := "n/a"
	parsedTime, err := time.Parse("2006.01.02 15:04:05", BuildTime)
	if err != nil {
		fmt.Println("Error parsing time:", err, BuildTime)
	} else {
		seconds := parsedTime.Second() + (parsedTime.Minute() * 60) + (parsedTime.Hour() * 3600)
		versionCode = fmt.Sprintf("%s.%v", parsedTime.Format("20060102"), seconds)
	}

	AppBuilding.IsDebug = IsDebug
	AppBuilding.AppShortId = AppShortId
	AppBuilding.AppId = AppId
	AppBuilding.VersionName = VersionName
	AppBuilding.VersionCode = versionCode
	AppBuilding.VersionFullName = VersionName + "-" + versionCode + "-" + VersionSHA
	AppBuilding.VersionSHA = VersionSHA
	AppBuilding.BuildType = BuildType
	AppBuilding.BuildTime = BuildTime
	AppBuilding.BuildHost = BuildHost
	AppBuilding.Flavor = Flavor

	fmt.Println("==============App info==============")
	fmt.Println("  IsDebug:", AppBuilding.IsDebug)
	fmt.Println("  AppShortId:", AppBuilding.AppShortId)
	fmt.Println("  AppId:", AppBuilding.AppId)
	fmt.Println("  VersionName:", AppBuilding.VersionName)
	fmt.Println("  VersionCode:", AppBuilding.VersionCode)
	fmt.Println("  VersionFullName:", AppBuilding.VersionFullName)
	fmt.Println("  VersionSHA:", AppBuilding.VersionSHA)
	fmt.Println("  BuildType:", AppBuilding.BuildType)
	fmt.Println("  BuildTime:", AppBuilding.BuildTime)
	fmt.Println("  BuildHost:", AppBuilding.BuildHost)
	fmt.Println("  Flavor:", AppBuilding.Flavor)
	fmt.Println("====================================")
}
