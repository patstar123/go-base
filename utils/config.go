package utils

import (
	"lx/meeting/base"
	"lx/meeting/base/logger"
	"os"
	"path/filepath"
)

func GetAppPath() (base.Result, string) {
	abs, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return base.INTERNAL_ERROR.AppendErr("filepath.Abs failed", err), ""
	}

	return base.SUCCESS, abs
}

func GetAppRealPath() (base.Result, string) {
	exePath, err := os.Executable()
	if err != nil {
		return base.INTERNAL_ERROR.AppendErr("os.Executable failed", err), ""
	}

	abs, err := filepath.Abs(filepath.Dir(exePath))
	if err != nil {
		return base.INTERNAL_ERROR.AppendErr("filepath.Abs failed", err), ""
	}

	return base.SUCCESS, abs
}

func GetConfig(configFile string, outStructPtr any) base.Result {
	config, err := os.ReadFile(configFile)
	if err != nil {
		writeConfig(configFile, outStructPtr)
		res := base.INTERNAL_ERROR.AppendErr("os.ReadFile("+configFile+") error", err)
		logger.Warnw("GetConfig failed", res)
		return res
	}

	err = yaml.Unmarshal(config, outStructPtr)
	if err != nil {
		res := base.INTERNAL_ERROR.AppendErr("yaml.Unmarshal("+string(config)+") error", err)
		logger.Warnw("GetConfig failed", res)
		return res
	}

	return base.SUCCESS
}

func writeConfig(filepath string, outStruct any) error {
	newData, err := yaml.Marshal(outStruct)
	if err != nil {
		logger.Warnw("write yaml file failed", err)
		return err
	}

	err = os.WriteFile(filepath, newData, 0644)
	if err != nil {
		logger.Warnw("write yaml file failed", err)
		return err
	}

	return nil
}
