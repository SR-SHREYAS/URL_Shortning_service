package job

import (
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"

	"github.com/SR-SHREYAS/URL_Shortning_service/internal/config"
	"github.com/SR-SHREYAS/URL_Shortning_service/internal/lib/geoip"
)

type JobService struct {
	Client *asynq.Client
	server *asynq.Server
	logger *zerolog.Logger
	cfg    *config.Config
	geo    *geoip.Client

	clicks ClickRecorder
}

func NewJobService(logger *zerolog.Logger, cfg *config.Config) *JobService {
	redisAddr := cfg.Redis.Address

	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})

	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr: redisAddr,
		},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6, // Higher priority queue for important emails
				"default":  3, // Default priority for most emails
				"low":      1, // Lower priority for non-urgent emails
			},
		},
	)

	return &JobService{
		Client: client,
		server: server,
		logger: logger,
		cfg:    cfg,

		geo: geoip.New(cfg.Service.GeoIPDBPath),
	}
}

func (j *JobService) Start() error {
	// Register task handlers
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskWelcome, j.handleWelcomeEmailTask)
	mux.HandleFunc(TaskClickEnrich, j.handleClickEnrichTask)

	j.logger.Info().Msg("Starting background job server")

	if err := j.server.Start(mux); err != nil {
		return err
	}

	return nil
}

func (j *JobService) Stop() {
	j.logger.Info().Msg("Stopping background job server")

	j.server.Shutdown()
	j.Client.Close()

	_ = j.geo.Close()
}
