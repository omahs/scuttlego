package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
)

var extractFromConfigSet = wire.NewSet(
	extractNetworkKeyFromConfig,
	extractMessageHMACFromConfig,
	extractLoggingSystemFromConfig,
	extractPeerManagerConfigFromConfig,
	extractHopsFromConfig,
)

func extractNetworkKeyFromConfig(config Config) boxstream.NetworkKey {
	return config.NetworkKey
}

func extractMessageHMACFromConfig(config Config) formats.MessageHMAC {
	return config.MessageHMAC
}

func extractLoggingSystemFromConfig(config Config) logging.LoggingSystem {
	return config.LoggingSystem
}

func extractPeerManagerConfigFromConfig(config Config) domain.PeerManagerConfig {
	return config.PeerManagerConfig
}

func extractHopsFromConfig(config Config) graph.Hops {
	return *config.Hops
}
