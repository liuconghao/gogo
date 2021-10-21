package cmd

import (
	"flag"
	"fmt"
	"getitle/src/core"
	"getitle/src/scan"
	. "getitle/src/utils"
	"github.com/panjf2000/ants/v2"
	"os"
	"strings"
	"time"
)

func CMD(k string) {
	defer ants.Release()
	connected = checkconn()
	if !strings.Contains(strings.Join(os.Args, ""), k) {
		inforev()
	}
	var config Config
	//默认参数信息
	flag.StringVar(&config.IP, "ip", "", "")
	flag.StringVar(&config.Ports, "p", "top1", "")
	flag.StringVar(&config.ListFile, "l", "", "")
	flag.StringVar(&config.JsonFile, "j", "", "")
	flag.IntVar(&config.Threads, "t", 4000, "")
	flag.StringVar(&config.Mod, "m", "default", "")
	flag.StringVar(&config.SmartPort, "sp", "default", "")
	flag.StringVar(&config.IpProbe, "ipp", "default", "")
	flag.BoolVar(&config.Spray, "s", false, "")
	flag.StringVar(&config.Filename, "f", "", "")
	flag.BoolVar(&config.NoSpray, "ns", false, "")

	//全局变量初始化
	flag.StringVar(&core.Output, "o", "full", "")
	flag.BoolVar(&core.Clean, "c", false, "")
	flag.StringVar(&core.FileOutput, "O", "json", "")
	flag.IntVar(&scan.Delay, "d", 2, "")
	flag.IntVar(&scan.HttpsDelay, "D", 2, "")
	flag.StringVar(&scan.Payloadstr, "payload", "", "")
	flag.BoolVar(&core.Noscan, "no", false, "")

	// 一些特殊参数初始化
	key := flag.String("k", "", "")
	version := flag.Bool("v", false, "")
	version2 := flag.Bool("vv", false, "")
	exploit := flag.Bool("e", false, "")
	exploitConfig := flag.String("E", "none", "")
	printType := flag.String("P", "no", "")
	formatoutput := flag.String("F", "", "")
	autofile := flag.Bool("af", false, "")
	noup := flag.Bool("nu", false, "")
	uploadfile := flag.String("uf", "", "")
	flag.Parse()
	// 密钥
	if *key != k {
		//rev()
		os.Exit(0)
	}

	// 输出 config
	printConfigs(*printType)

	// 格式化
	if *formatoutput != "" {
		core.FormatOutput(*formatoutput, config.Filename)
		os.Exit(0)
	} else if *uploadfile != "" {
		uploadfiles([]string{*uploadfile})
	}

	starttime := time.Now()

	parseVersion(*version, *version2)
	parseExploit(*exploit, *exploitConfig)
	parseFilename(*autofile, &config)

	config = core.Init(config)
	core.RunTask(config)

	//关闭文件写入管道
	close(core.Datach)
	close(core.LogDetach)

	time.Sleep(500 * time.Microsecond)
	if connected && !*noup && config.Filename != "" { // 如果出网则自动上传结果到云服务器
		uploadfiles([]string{config.Filename, config.SmartFilename})
	}

	time.Sleep(time.Microsecond * 500)
	fmt.Printf("\n[*] Alive sum: %d, Target sum : %d\n", core.Alivesum, scan.Sum)
	fmt.Println("[*] Totally run: " + time.Since(starttime).String())

}

func printConfigs(t string) {
	if t == "no" {
		return
	}
	if t == "port" {
		core.Printportconfig()
	} else if t == "nuclei" {
		core.PrintNucleiPoc()
	} else if t == "inter" {
		core.PrintInterConfig()
	} else {
		fmt.Println("choice port|nuclei|inter")
	}
	os.Exit(0)
}
