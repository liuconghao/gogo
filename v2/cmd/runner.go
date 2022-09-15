package cmd

import (
	"fmt"
	. "github.com/chainreactors/gogo/v2/internal/core"
	. "github.com/chainreactors/gogo/v2/internal/plugin"
	. "github.com/chainreactors/gogo/v2/pkg"
	nucleihttp "github.com/chainreactors/gogo/v2/pkg/nuclei/protocols/http"
	. "github.com/chainreactors/gogo/v2/pkg/utils"
	. "github.com/chainreactors/logs"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

func NewRunner() *Runner {
	return &Runner{
		Config: Config{},
	}
}

type Runner struct {
	InputOption  `group:"Input Options"`
	OutputOption `group:"Output Options"`
	MiscOption   `group:"Miscellaneous Options"`
	SmartOption  `group:"Smart Options"`
	ConfigOption `group:"Configuration Options"`
	//Arp          bool

	start  time.Time
	Config Config
}

func (r *Runner) preInit() bool {
	// 初始化日志工具"
	if r.Debug {
		Log = NewLogger(0, r.Quiet)
	} else {
		Log = NewLogger(1, r.Quiet)
	}
	Log.LogFileName = ".sock.lock"
	Log.Init()

	RunOpt = RunnerOpts{
		Delay:      r.Delay,
		HttpsDelay: r.HttpsDelay,
		SuffixStr:  r.SuffixStr,
	}
	Opt.PluginDebug = r.PluginDebug
	Key = []byte(r.Key)

	r.Config = Config{
		IP:          r.IP,
		Ports:       r.Ports,
		ListFile:    r.ListFile,
		JsonFile:    r.JsonFile,
		IsListInput: r.IsListInput,
		IsJsonInput: r.IsJsonInput,
		SmartPort:   r.SmartPort,
		IpProbe:     r.IpProbe,
		NoSpray:     r.NoSpray,
		Filename:    r.Filename,
		FilePath:    r.FilePath,
		Compress:    r.Compress,
		Threads:     r.Threads,
		PortSpray:   r.PortSpray,
		Mod:         r.Mod,
	}

	if r.FileOutputf == Default {
		r.Config.FileOutputf = "json"
	} else {
		r.Config.FileOutputf = r.FileOutputf
	}

	if r.Outputf == Default {
		r.Config.Outputf = "full"
	} else {
		r.Config.Outputf = r.Outputf
	}

	r.Config.Compress = !r.Config.Compress
	if r.AutoFile {
		r.Config.Filenamef = "auto"
	} else if r.HiddenFile {
		r.Config.Filenamef = "hidden"
	}

	// 一些特殊的分支, 这些分支将会直接退出程序
	if r.Ver {
		fmt.Println(ver)
		return false
	}

	if r.FormatterFilename != "" {
		FormatOutput(r.FormatterFilename, r.Config.Filename, r.Config.Outputf, r.Config.FileOutputf, r.filters)
		return false
	}
	// 输出 Config
	if r.Printer != "" {
		printConfigs(r.Printer)
		return false
	}

	if r.Proxy != "" {
		uri, err := url.Parse(r.Proxy)
		if err == nil {
			ProxyUrl = uri
			Proxy = http.ProxyURL(uri)
			nucleihttp.Proxy = Proxy
			ProxyDialTimeout = func(network, address string, duration time.Duration) (net.Conn, error) {
				forward := &net.Dialer{Timeout: duration}
				dial, err := proxy.FromURL(uri, forward)
				if err != nil {
					return nil, err
				}
				conn, err := dial.Dial(network, address)
				if err != nil {
					return nil, err
				}
				return conn, nil
			}

		} else {
			Log.Warnf("parse proxy error %s, skip proxy!", err.Error())
		}
		//if !r.Debug {
		//	Log.Error("-proxy is debug only flag, please add -debug. skipped proxy")
		//} else {
		//	Log.Importantf("DEBUG ONLY, set http proxy: " + r.Proxy)
		//
		//}
	}
	//if r.UploadFile != "" {
	//	// 指定上传文件
	//	uploadfiles(strings.Split(r.UploadFile, ","))
	//	return false
	//}
	return true
}

func (r *Runner) init() {
	// 初始化各种全局变量
	// 初始化指纹优先级
	if r.Version {
		RunOpt.VersionLevel = 1
		//} else if r.Version2 {
		//	RunOpt.VersionLevel = 2
	} else {
		RunOpt.VersionLevel = 0
	}

	// 初始化漏洞
	if r.Exploit {
		RunOpt.Exploit = "auto"
	} else {
		RunOpt.Exploit = r.ExploitName
	}

	if r.NoScan {
		Opt.Noscan = r.NoScan
	}

	if r.Clean {
		Log.Clean = !Log.Clean
	}

	//if !Win {
	//	//if r.iface == "eth0" {
	//	//	Log.Warn("no interface name input, use default interface name: eth0")
	//	//}
	//	var err error
	//	RunOpt.Interface, err = net.InterfaceByName(r.iface)
	//	if err != nil {
	//		Log.Warn("interface error, " + err.Error())
	//		//Log.Warn("interface error, " + err.Error())
	//		//Log.Warn("interface error, " + err.Error())
	//	}
	//}

	if r.extracts != "" {
		exts := strings.Split(r.extracts, ",")
		for _, extract := range exts {
			if reg, ok := PresetExtracts[extract]; ok {
				Extractors[extract] = reg
			}
		}
	}
	for _, extract := range r.extract {
		if reg, ok := PresetExtracts[extract]; ok {
			Extractors[extract] = reg
		} else {
			Extractors[extract] = CompileRegexp(extract)
		}
	}

	// 加载配置文件中的全局变量
	templatesLoader()
	nucleiLoader(r.ExploitFile, r.payloads)
}

func (r *Runner) prepareConfig(config Config) *Config {
	if r.Config.Ports == "" {
		config.Ports = "top1"
	}

	//if r.Arp {
	//	config.AliveSprayMod = append(config.AliveSprayMod, "arp")
	//}

	if r.Ping {
		config.AliveSprayMod = append(config.AliveSprayMod, "icmp")
	}

	//if Config.Filename == "" {
	//	Config.Filename = GetFilename(&Config, Config.FileOutputf)
	//} else {
	//	Config.Filename = path.Join(Config.FilePath, Config.Filename)
	//}

	//if Config.IsSmart() {
	//	if r.NoScan && !r.AutoFile && !r.HiddenFile {
	//		Config.SmartFilename = Config.Filename
	//	} else {
	//		Config.SmartFilename = GetFilename(&Config, "cidr")
	//	}
	//}

	//if Config.HasAlivedScan() {
	//	Config.AlivedFilename = GetFilename(&Config, "alived")
	//}
	return &config
}

func (r *Runner) run() {
	r.start = time.Now()
	if r.WorkFlowName == "" && !r.IsWorkFlow {
		r.runWithCMD()
	} else {
		var workflowMap = WorkflowMap{}
		if r.IsWorkFlow {
			workflowMap["tmp"] = ParseWorkflowsFromInput(LoadFile(os.Stdin))
			r.WorkFlowName = "tmp"
		} else if IsExist(r.WorkFlowName) {
			workflowMap["tmp"] = ParseWorkflowsFromInput(LoadFile(Open(r.WorkFlowName)))
			r.WorkFlowName = "tmp"
		} else {
			workflowMap = LoadWorkFlow()
		}
		r.runWithWorkFlow(workflowMap)
	}
}

func (r *Runner) runWithCMD() {
	config := r.prepareConfig(r.Config)
	if config.Mod == SUPERSMARTB {
		config.FileOutputf = SUPERSMARTB
	}
	if config.Filename == "" && config.IsSmart() {
		config.SmartFilename = GetFilename(config, "cidr")
	}
	if config.Filename == "" && config.HasAlivedScan() {
		config.AlivedFilename = GetFilename(config, "alived")
	}

	if config.Filename != "" || config.Filenamef != "" {
		Log.Warn("The result file has been specified, other files will not be created.")
		config.Filename = GetFilename(config, config.FileOutputf)
	}
	preparedConfig, err := InitConfig(config)
	if err != nil {
		Fatal(err.Error())
	}
	RunTask(*preparedConfig) // 运行
	r.close(config)
}

func (r *Runner) runWithWorkFlow(workflowMap WorkflowMap) {
	if workflows := workflowMap.Choice(r.WorkFlowName); len(workflows) > 0 {
		for _, workflow := range workflows {
			Log.Important("workflow " + workflow.Name + " starting")
			config := workflow.PrepareConfig(r.Config)

			if config.Mod == SUPERSMARTB {
				config.FileOutputf = SUPERSMARTB
			}

			if config.Filename == "" && config.IsSmart() {
				config.SmartFilename = GetFilename(config, "cidr")
			}
			if config.Filename == "" && config.HasAlivedScan() {
				config.AlivedFilename = GetFilename(config, "alived")
			}

			if config.Filename != "" || config.Filenamef != "" {
				Log.Warn("The result file has been specified, other files will not be created.")
				config.Filename = GetFilename(config, config.FileOutputf)
			}

			// 全局变量的处理
			if !r.NoScan {
				Opt.Noscan = workflow.NoScan
			}

			if r.Version {
				RunOpt.VersionLevel = 1
			} else {
				RunOpt.VersionLevel = workflow.Version
			}

			if RunOpt.Exploit != "none" {
				if r.Exploit {
					RunOpt.Exploit = "auto"
				} else {
					RunOpt.Exploit = r.ExploitName
				}
			} else {
				RunOpt.Exploit = workflow.Exploit
			}

			preparedConfig, err := InitConfig(config)
			if err != nil {
				Fatal(err.Error())
			}
			RunTask(*preparedConfig) // 运行
			r.close(config)
			r.resetGlobals()
		}
	} else {
		Fatal("not fount workflow " + r.WorkFlowName)
	}
}

func (r *Runner) close(config *Config) {
	config.Close() // 关闭result与extract写入管道

	if r.HiddenFile {
		Chtime(config.Filename)
		if config.SmartFilename != "" {
			Chtime(config.SmartFilename)
		}
	}

	// 任务统计
	Log.Importantf("Alive sum: %d, Target sum : %d", Opt.AliveSum, RunOpt.Sum)
	Log.Important("Totally: " + time.Since(r.start).String())

	// 输出文件名
	if config.File != nil && config.File.InitSuccess {
		Log.Importantf("Results: " + config.Filename)
	}
	if config.SmartFile != nil && config.SmartFile.InitSuccess {
		Log.Important("Smart result: " + config.SmartFilename)
	}
	if config.AliveFile != nil && config.AliveFile.Initialized {
		Log.Important("Alived result: " + config.AlivedFilename)
	}
	if IsExist(config.Filename + "_extract") {
		Log.Important("extractor result: " + config.Filename + "_extract")
	}

	// 扫描结果文件自动上传
	//if connected && !r.NoUpload { // 如果出网则自动上传结果到云服务器
	//	uploadfiles([]string{Config.Filename, Config.SmartFilename})
	//}
}

func (r *Runner) resetGlobals() {
	Opt.Noscan = false
	RunOpt.Exploit = "none"
	RunOpt.VersionLevel = 0
}

func printConfigs(t string) {
	if t == "port" {
		LoadPortConfig()
		Printportconfig()
	} else if t == "nuclei" {
		nucleiLoader("", nil)
		PrintNucleiPoc()
	} else if t == "workflow" {
		PrintWorkflow()
	} else if t == "extract" {
		PrintExtract()
	} else {
		fmt.Println("choice port|nuclei|workflow|extract")
	}
}

func nucleiLoader(pocfile string, payloads []string) {
	ExecuterOptions = ParserCmdPayload(payloads)
	TemplateMap = LoadNuclei(pocfile)
}

func templatesLoader() {
	LoadPortConfig()
	AllFingers = LoadFinger("http")
	Mmh3Fingers, Md5Fingers = LoadHashFinger(AllFingers)
	TcpFingers = LoadFinger("tcp").GroupByPort()
	HttpFingers = AllFingers.GroupByPort()
	CommonCompiled = map[string]*regexp.Regexp{
		"title":     CompileRegexp("(?Uis)<title>(.*)</title>"),
		"server":    CompileRegexp("(?i)Server: ([\x20-\x7e]+)"),
		"xpb":       CompileRegexp("(?i)X-Powered-By: ([\x20-\x7e]+)"),
		"sessionid": CompileRegexp("(?i) (.*SESS.*?ID)"),
	}
}
