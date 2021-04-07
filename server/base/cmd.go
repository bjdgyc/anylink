package base

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// 提交id
	CommitId string
	// 配置文件
	cfgFile string
	// pass明文
	passwd string
	// 生成密钥
	secret bool
	// 显示版本信息
	rev bool
	// 获取env名称
	env bool

	// Used for flags.
	runSrv bool

	rootCmd *cobra.Command
)

// Execute executes the root command.
func execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// viper.Debug()

	if !runSrv {
		os.Exit(0)
	}
}

func init() {
	rootCmd = &cobra.Command{
		Use:   "anylink",
		Short: "AnyLink VPN Server",
		Long:  `AnyLink is a VPN Server application`,
		Run: func(cmd *cobra.Command, args []string) {
			// fmt.Println("cmd：", cmd.Use, args)
			runSrv = true
		},
	}

	cobra.OnInitialize(func() {
		viper.SetConfigFile(cfgFile)
		viper.AutomaticEnv()

		err := viper.ReadInConfig()
		if err != nil {
			fmt.Println("Using config file:", err)
		}
	})

	viper.SetEnvPrefix("link")

	// 基础配置
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "./conf/server.toml", "config file")

	for _, v := range configs {
		if v.Typ == cfgStr {
			rootCmd.Flags().String(v.Name, v.ValStr, v.Usage)
		}
		if v.Typ == cfgInt {
			rootCmd.Flags().Int(v.Name, v.ValInt, v.Usage)
		}
		if v.Typ == cfgBool {
			rootCmd.Flags().Bool(v.Name, v.ValBool, v.Usage)
		}

		_ = viper.BindPFlag(v.Name, rootCmd.Flags().Lookup(v.Name))
		_ = viper.BindEnv(v.Name)
		// viper.SetDefault(v.Name, v.Value)
	}

	rootCmd.AddCommand(initToolCmd())
}

func initToolCmd() *cobra.Command {
	toolCmd := &cobra.Command{
		Use:   "tool",
		Short: "AnyLink tool",
		Long:  `AnyLink tool is a application`,
	}

	toolCmd.Flags().BoolVarP(&rev, "version", "v", false, "display version info")
	toolCmd.Flags().BoolVarP(&secret, "secret", "s", false, "generate a random jwt secret")
	toolCmd.Flags().StringVarP(&passwd, "passwd", "p", "", "convert the password plaintext")
	toolCmd.Flags().BoolVarP(&env, "env", "e", false, "list the config name and env key")

	toolCmd.Run = func(cmd *cobra.Command, args []string) {
		switch {
		case rev:
			fmt.Printf("%s v%s build on %s [%s, %s] commit_id(%s) \n",
				APP_NAME, APP_VER, runtime.Version(), runtime.GOOS, runtime.GOARCH, CommitId)
		case secret:
			rand.Seed(time.Now().UnixNano())
			s, _ := utils.RandSecret(40, 60)
			s = strings.Trim(s, "=")
			fmt.Printf("Secret:%s\n", s)
		case passwd != "":
			pass, _ := utils.PasswordHash(passwd)
			fmt.Printf("Passwd:%s\n", pass)
		case env:
			for k, v := range envs {
				fmt.Printf("%s => %s\n", k, v)
			}
		default:
			fmt.Println("Using [anylink tool -h] for help")
		}
	}

	return toolCmd
}
