package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	// "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	grpchealth "github.com/bufbuild/connect-grpchealth-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	api "github.com/okpalaChidiebere/chirper-app-api-image/api"
	"github.com/okpalaChidiebere/chirper-app-api-image/config"
	imagefilterservice "github.com/okpalaChidiebere/chirper-app-api-image/v0/image-filter/business_logic"
	presignerrepo "github.com/okpalaChidiebere/chirper-app-api-image/v0/image-filter/data_access"
)

func main() {

	var (
		cfg aws.Config
		err error
		mConfig    = config.NewConfig()
	)

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if mConfig.IsLocal() {
		cfg, err = awsconfig.LoadDefaultConfig(ctx, 
		awsconfig.WithSharedConfigProfile(mConfig.Aws.Aws_profile))
		if err != nil {
			log.Printf("unable to load local SDK config, %v\n", err)
			os.Exit(2)
		}
	} else {
		cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(mConfig.Aws.Aws_region))
		if err != nil {
			log.Printf("unable to load SDK config, %v\n", err)
			os.Exit(2)
		}
	}



	httpClient := http.Client{}
	// dynamodbClient := dynamodb.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3Client)

	presignRepo := presignerrepo.NewPresignerRepository(presignClient)

	imagefilterService := imagefilterservice.New(presignRepo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	s := api.Servers{
		ImageFilterServer: api.NewImageFilterServer(imagefilterService, &httpClient),
	}

	mux := http.NewServeMux()

	apiServer := s.NewAPIServer(mux)

	sh_map := s.GetAllServiceHandlers()

	apiServer.RegisterServiceHandlers(sh_map)

	var services []string
	for key := range sh_map {
		services = append(services, key)
	}
	reflector := grpcreflect.NewStaticReflector(services...)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
  	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	//add support for gRPC style health checks and http style health checks as well
	//documentation: https://github.com/bufbuild/connect-grpchealth-go
	checker := grpchealth.NewStaticChecker(services...)
	mux.Handle(grpchealth.NewHandler(checker))

	httpLis, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Printf("HTTP server: failed to listen: error %v", err)
		os.Exit(2)
	}

	httpServer := &http.Server{
		// Use h2c so we can serve HTTP/2 without TLS.
		Handler: h2c.NewHandler(mux, &http2.Server{}),
		// Don't forget timeouts!
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	http.Handle("/favicon.ico", http.NotFoundHandler())

	go func() {
		log.Printf("server listening at %v", httpLis.Addr())
		err = httpServer.Serve(httpLis)
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Println("server closed")
		} else if err != nil {
			panic(err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	fmt.Println("shutting down gracefully, press Ctrl+C again to force")

	// Perform application shutdown with a maximum timeout of 10 seconds.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}
}
