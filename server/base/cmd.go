package base

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// 提交id
	CommitId string
	// pass明文
	passwd string
	// 生成密钥
	secret bool
	// 显示版本信息
	rev bool
	// 输出debug信息
	debug bool

	// Used for flags.
	runSrv bool

	linkViper *viper.Viper
	rootCmd   *cobra.Command
)

// Execute executes the root command.
func execute() {
	initCmd()

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// viper.Debug()
	ref := reflect.ValueOf(linkViper)
	s := ref.Elem()
	ee := s.FieldByName("env")
	if ee.Kind() != reflect.Map {
		panic("Viper env is err")
	}
	rr := ee.MapRange()
	for rr.Next() {
		// fmt.Println(rr.Key(), rr.Value().Index(0))
		envs[rr.Key().String()] = rr.Value().Index(0).String()
	}

	if !runSrv {
		os.Exit(0)
	}
}

func initCmd() {
	linkViper = viper.New()
	rootCmd = &cobra.Command{
		Use:   "anylink",
		Short: "AnyLink VPN Server",
		Long:  `AnyLink is a VPN Server application`,
		Run: func(cmd *cobra.Command, args []string) {
			// fmt.Println("cmd：", cmd.Use, args)
			runSrv = true

			if rev {
				printVersion()
				os.Exit(0)
			}
		},
	}

	linkViper.SetEnvPrefix("link")

	// 基础配置

	for _, v := range configs {
		if v.Typ == cfgStr {
			rootCmd.Flags().StringP(v.Name, v.Short, v.ValStr, v.Usage)
		}
		if v.Typ == cfgInt {
			rootCmd.Flags().IntP(v.Name, v.Short, v.ValInt, v.Usage)
		}
		if v.Typ == cfgBool {
			rootCmd.Flags().BoolP(v.Name, v.Short, v.ValBool, v.Usage)
		}

		_ = linkViper.BindPFlag(v.Name, rootCmd.Flags().Lookup(v.Name))
		_ = linkViper.BindEnv(v.Name)
		// viper.SetDefault(v.Name, v.Value)
	}

	rootCmd.Flags().BoolVarP(&rev, "version", "v", false, "display version info")
	rootCmd.AddCommand(initToolCmd())

	cobra.OnInitialize(func() {
		linkViper.AutomaticEnv()
		conf := linkViper.GetString("conf")

		_, err := os.Stat(conf)
		if errors.Is(err, os.ErrNotExist) {
			// 没有配置文件，不做处理
			panic(err)
		}

		linkViper.SetConfigFile(conf)
		err = linkViper.ReadInConfig()
		if err != nil {
			fmt.Println("Using config file:", err)
		}
	})
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
	toolCmd.Flags().BoolVarP(&debug, "debug", "d", false, "list the config viper.Debug() info")

	toolCmd.Run = func(cmd *cobra.Command, args []string) {
		switch {
		case rev:
			printVersion()
		case secret:
			s, _ := utils.RandSecret(40, 60)
			s = strings.Trim(s, "=")
			fmt.Printf("Secret:%s\n", s)
		case passwd != "":
			pass, _ := utils.PasswordHash(passwd)
			fmt.Printf("Passwd:%s\n", pass)
		case debug:
			linkViper.Debug()
		default:
			fmt.Println("Using [anylink tool -h] for help")
		}
	}

	return toolCmd
}

func printVersion() {
	fmt.Printf("%s v%s build on %s [%s, %s] commit_id(%s) \n",
		APP_NAME, APP_VER, runtime.Version(), runtime.GOOS, runtime.GOARCH, CommitId)
}
