package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	petstore_client "github.com/nagymarci/petstore-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuy(t *testing.T) {
	srv := createServer(t)
	defer srv.Close()

	subject := NewClient(srv.URL)

	tests := map[string]struct {
		input    int64
		expected int64
		err      error
	}{
		"success": {
			input:    availablePet,
			expected: availablePet * 10,
		},
		"not_available": {
			input: notAvailablePet,
			err:   ErrNotAvailable,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			id, err := subject.Buy(context.Background(), test.input)
			require.ErrorIs(t, err, test.err)

			assert.Equal(t, test.expected, id)
		})
	}

}

func createServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/pet/", func(w http.ResponseWriter, r *http.Request) {
		pathParams := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		require.True(t, len(pathParams) == 2 && r.Method == http.MethodGet, "unexpected path or method")

		petId, err := strconv.ParseInt(pathParams[1], 10, 0)
		require.NoError(t, err, "no valid petId")

		pet, ok := pets[petId]
		require.True(t, ok, "pet not found %d", petId)

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(pet.responseCode)
		require.NoError(t, json.NewEncoder(w).Encode(pet.pet))
	})

	mux.HandleFunc("/store/", func(w http.ResponseWriter, r *http.Request) {
		pathParams := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		require.True(t, len(pathParams) == 2 &&
			pathParams[1] == "order" &&
			r.Method == http.MethodPost, "unexpected path or method")

		var order petstore_client.Order
		require.NoError(t, json.NewDecoder(r.Body).Decode(&order))

		response, ok := orders[order.GetPetId()]
		require.True(t, ok, "order not found for pet %d", order.GetPetId())

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(response.responseCode)
		require.NoError(t, json.NewEncoder(w).Encode(response.order))

	})
	srv := httptest.NewServer(mux)

	return srv
}

const availablePet int64 = 1
const notAvailablePet int64 = 2

var pets = map[int64]struct {
	pet          petstore_client.Pet
	responseCode int
}{
	availablePet: {
		pet: petstore_client.Pet{
			Id:     petstore_client.PtrInt64(availablePet),
			Status: petstore_client.PtrString("available"),
		},
		responseCode: http.StatusOK,
	},
	notAvailablePet: {
		pet: petstore_client.Pet{
			Id:     petstore_client.PtrInt64(availablePet),
			Status: petstore_client.PtrString("pending"),
		},
		responseCode: http.StatusOK,
	},
}

var orders = map[int64]struct {
	order        petstore_client.Order
	responseCode int
}{
	availablePet: {
		order: petstore_client.Order{
			Id: petstore_client.PtrInt64(availablePet * 10),
		},
		responseCode: http.StatusOK,
	},
}
