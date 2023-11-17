package app

import (
	"context"
	"errors"
	"fmt"

	petstore_client "github.com/nagymarci/petstore-client"
)

var ErrNotAvailable = errors.New("pet not available")

type client struct {
	client *petstore_client.APIClient
}

func (c *client) Buy(ctx context.Context, id int64) (int64, error) {
	pet, resp, err := c.client.PetAPI.GetPetById(ctx, id).Execute()
	if err != nil {
		return 0, fmt.Errorf("get pet failed for id %d: %w", id, err)
	}
	defer resp.Body.Close()

	if pet.GetStatus() != "available" {
		return 0, ErrNotAvailable
	}

	request := petstore_client.NewOrder()
	request.PetId = pet.Id
	order, orderResp, err := c.client.StoreAPI.PlaceOrder(ctx).Order(*request).Execute()
	if err != nil {
		return 0, fmt.Errorf("create order failed for pet %d: %w", *pet.Id, err)
	}
	defer orderResp.Body.Close()

	return order.GetId(), nil
}

func NewClient(url string) *client {
	config := petstore_client.NewConfiguration()
	config.Servers[0].URL = url

	petstore := petstore_client.NewAPIClient(config)
	return &client{
		client: petstore,
	}
}
