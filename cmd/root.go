package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

var cfgFile string
var jaegerAgentHost string
var jaegerAgentPort int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jaeger-me",
	Short: "jaeger-me example",
	Long:  "jaeger-me example demonstrates integration with jaeger implementation with remote jaeger possibly on istio",
	Run:   jaegerMe,
}

func jaegerMe(cmd *cobra.Command, args []string) {
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: jaegerAgentHost + ":" + strconv.Itoa(jaegerAgentPort),
		},
	}

	tracer, closer, err := cfg.New(
		"jaeger-me",
		config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		panic(err)
	}
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	span := tracer.StartSpan("say-hello")
	span.SetTag("hello-to", "govinda")
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	helloStr := formatString(ctx, "govinda")
	printHello(ctx, helloStr)
}

func formatString(ctx context.Context, helloTo string) string {
	span, _ := opentracing.StartSpanFromContext(ctx, "formatString")
	defer span.Finish()

	helloStr := fmt.Sprintf("Hello, %s!", helloTo)
	span.LogFields(
		log.String("event", "string-format"),
		log.String("value", helloStr),
	)

	return helloStr
}

func printHello(ctx context.Context, helloStr string) {
	span, _ := opentracing.StartSpanFromContext(ctx, "printHello")
	defer span.Finish()

	println(helloStr)
	span.LogKV("event", "println")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().StringVar(&jaegerAgentHost, "jaegerAgentHost", "localhost", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().IntVar(&jaegerAgentPort, "jaegerAgentPort", 5775, "config file (default is ./config.yaml)")
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
