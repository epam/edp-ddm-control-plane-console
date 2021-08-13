package main

import (
	"ddm-admin-console/auth"
	oauth "ddm-admin-console/auth"
	"ddm-admin-console/cluster"
	"ddm-admin-console/config"
	"ddm-admin-console/dashboard"
	"ddm-admin-console/registry"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/jenkins"
	"ddm-admin-console/service/k8s"
	"ddm-admin-console/service/openshift"
	"encoding/gob"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/leonelquinteros/gotext"

	"golang.org/x/oauth2"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	configPath := flag.String("c", "default.env", "config file path")
	flag.Parse()

	cnf, err := loadConfig(*configPath)
	if err != nil {
		panic(err)
	}

	logger, err := getLogger(cnf.LogLevel, cnf.LogEncoding)
	if err != nil {
		panic(err)
	}

	logger.Info("init gin router")
	gin.SetMode(cnf.GinMode)
	r := gin.New()
	r.SetFuncMap(template.FuncMap{"i18n": i18n})
	r.LoadHTMLGlob("templates/**/*")
	r.Static("/static", "./static")
	store := cookie.NewStore([]byte(cnf.SessionSecret))
	r.Use(sessions.Sessions("cookie-session", store))

	logger.Info("init apps")
	if err := initApps(logger, cnf, r); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	logger.Info("init i18n")
	gotext.Configure("locale", "uk_UA", "default")

	logger.Info("run router on port", zap.String("port", cnf.HTTPPort))
	if err := r.Run(fmt.Sprintf(":%s", cnf.HTTPPort)); err != nil {
		panic(err)
	}
}

func loadConfig(path string) (*config.Settings, error) {
	if err := godotenv.Load(path); err != nil {
		return nil, errors.Wrap(err, "unable to load config file")
	}

	var cnf config.Settings
	if err := envconfig.Process("", &cnf); err != nil {
		return nil, errors.Wrap(err, "unable to parse env variables")
	}

	return &cnf, nil
}

func getLogger(level, encoding string) (*zap.Logger, error) {
	levels := map[string]zapcore.Level{
		"DEBUG":   zap.DebugLevel,
		"INFO":    zap.InfoLevel,
		"WARNING": zap.WarnLevel,
		"ERROR":   zap.ErrorLevel,
		"DPANIC":  zap.DPanicLevel,
		"PANIC":   zap.PanicLevel,
		"FATAL":   zap.FatalLevel,
	}

	logLevel, ok := levels[level]
	if !ok {
		logLevel = zap.InfoLevel
	}

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(logLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         encoding,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, errors.Wrap(err, "unable to build logger")
	}

	return logger, nil
}

func initApps(logger *zap.Logger, cnf *config.Settings, r *gin.Engine) error {
	restConf, err := initKubeConfig()
	if err != nil {
		return errors.Wrap(err, "unable to init kube config")
	}

	oa, err := initOauth(restConf, cnf, r)
	if err != nil {
		return errors.Wrap(err, "unable to init oauth")
	}

	edpComponentService, err := edpComponent.Make(restConf, cnf.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to init edp component service")
	}

	codebaseService, err := codebase.Make(restConf, cnf.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to init codebase service")
	}

	k8sService, err := k8s.Make(restConf, cnf.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to init k8s service")
	}

	jenkinsService, err := jenkins.Make(restConf, cnf.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to init jenkins service")
	}

	openShiftService, err := openshift.Make(restConf, cnf.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to init open shift service")
	}

	appRouter := router.Make(r, logger)

	_, err = dashboard.Make(appRouter, edpComponentService, oa, k8sService, openShiftService, codebaseService,
		cnf.ClusterCodebaseName)
	if err != nil {
		return errors.Wrap(err, "unable to make dashboard app")
	}

	_, err = registry.Make(appRouter, logger, codebaseService, edpComponentService, k8sService, jenkinsService, cnf)
	if err != nil {
		return errors.Wrap(err, "unable to make registry app")
	}

	_, err = cluster.Make(appRouter, logger, codebaseService, jenkinsService, edpComponentService, k8sService, cnf)
	if err != nil {
		return errors.Wrap(err, "unable to init cluster app")
	}

	return nil
}

func initKubeConfig() (*rest.Config, error) {
	k8sConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restConfig, err := k8sConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get k8s client config")
	}

	return restConfig, nil
}

func initOauth(k8sConfig *rest.Config, cfg *config.Settings, r *gin.Engine) (*auth.OAuth2, error) {
	transport, err := rest.TransportFor(k8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create transport for k8s config")
	}

	oa, err := oauth.InitOauth2(
		cfg.OCClientID,
		cfg.OCClientSecret,
		k8sConfig.Host,
		cfg.Host+"/auth/callback",
		&http.Client{Transport: transport})
	if err != nil {
		return nil, errors.Wrap(err, "unable to init oauth2 client")
	}

	gob.Register(&oauth2.Token{})
	r.Use(oauth.MakeGinMiddleware(oa, router.AuthTokenSessionKey, router.AuthTokenValidSessionKey, "/admin/"))
	r.Use(router.UserDataMiddleware)

	return oa, nil
}

func i18n(word ...string) string {
	message := strings.TrimSpace(strings.Join(word, " "))
	return gotext.Get(message)
}
