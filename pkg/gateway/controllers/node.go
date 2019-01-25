package controllers

import (
	"github.com/gladiusio/gladius-common/pkg/blockchain"
	"github.com/gladiusio/gladius-common/pkg/db/models"
	"github.com/gladiusio/gladius-common/pkg/utils"
	"net/http"
	"github.com/rs/zerolog/log"
	"encoding/json"
	"github.com/spf13/viper"
	"github.com/gladiusio/gladius-p2p/pkg/p2p/message"
	"github.com/gladiusio/gladius-p2p/pkg/p2p/signature"
)

func ApplyToPool(poolAddress string, ga *blockchain.GladiusAccountManager) {
	var poolURL string

	if viper.GetString("Pool.URL") != "" {
		poolURL = viper.GetString("Pool.URL")
	} else {
		var err error
		poolURL, err = blockchain.PoolRetrieveApplicationServerUrl(poolAddress, ga)

		if err != nil {
			log.Error().Err(err).Msg("Pool URL could not be found for provided address")
			return
		}
	}

	accountAddress, _ := ga.GetAccountAddress()

	var requestPayload = models.NodeRequestPayload {
		EstimatedSpeed: viper.GetInt("Profile.EstimatedSpeed"),
		Wallet: accountAddress.String(),
		Name: viper.GetString("Profile.Name"),
		Email: viper.GetString("Profile.Email"),
		Bio: viper.GetString("Profile.Bio"),
		IPAddress: "",
	}

	payload, err := json.Marshal(requestPayload)
	if err != nil {
		log.Error().Err(err).Msg("Could not marshal request payload")
		return
	}

	unsignedMessage := message.New(payload)
	signedMessage, err := signature.CreateSignedMessage(unsignedMessage, ga)
	if err != nil {
		log.Error().Err(err).Msg("Could not create signed message")
		return
	}

	_, err = utils.SendRequest(http.MethodPost, poolURL + "applications/new", signedMessage)

	if err != nil {
		log.Error().Err(err).Msg("Could not complete application")
		return
	}
}