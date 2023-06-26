package main

import (
	"fmt"
	"os"
	"time"

	"github.com/lzfDream/ReadExcel/config"
	"github.com/lzfDream/ReadExcel/output"
	"github.com/lzfDream/ReadExcel/parse"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	begin := time.Now()
	cfg := config.GetConfig()

	entries, err := os.ReadDir(cfg.InputPath)
	if err != nil {
		logrus.Errorln("读取目录失败：", err)
		return
	}
	defines := make([]parse.ExcelDefine, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		logrus.Infof("开始读取文件: %s", fileName)
		startTime := time.Now()

		f, err := excelize.OpenFile(cfg.InputPath + "/" + fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func() {
			if err := f.Close(); err != nil {
				logrus.Errorln(err)
				return
			}
		}()

		for _, sheetName := range f.GetSheetList() {
			logrus.Infof("开始解析 %s:%s", fileName, sheetName)
			define, clientData, err := parse.ReadSheet(fileName, sheetName, f)
			if err != nil {
				logrus.Errorln(err)
				return
			}
			defines = append(defines, define)

			if len(clientData) == 0 {
				logrus.Infof("%s client 数据为空, 被跳过", define.Desc())
			} else {
				err = output.OutputData(cfg.Client, define, clientData)
				if err != nil {
					logrus.Errorln(err)
					return
				}
			}
		}
		elapsed := time.Since(startTime)
		logrus.Infof("读取%s完成, 耗时%s\n\n", fileName, elapsed)
	}
	err = output.OutputCode(cfg.Client, defines, parse.GroupType_Client)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	logrus.Infof("导出成功, 总耗时%s", time.Since(begin))
}
